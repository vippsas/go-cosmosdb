package cosmosapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/logging"
)

const (
	apiVersion = "2018-12-31"
)

var (
	// TODO: useful?
	IgnoreContext bool
	// TODO: check thread safety
	ResponseHook            func(ctx context.Context, method string, headers map[string][]string)
	errUnexpectedHTTPStatus = errors.New("Unexpected HTTP return status")
)

type ResponseBase struct {
	RequestCharge float64
}

func parseHttpResponse(httpResponse *http.Response) (ResponseBase, error) {
	response := ResponseBase{}
	if header := httpResponse.Header.Get(HEADER_REQUEST_CHARGE); header != "" {
		if requestCharge, err := strconv.ParseFloat(header, 64); err != nil {
			return response, errors.WithStack(err)
		} else {
			response.RequestCharge = requestCharge
		}
	}
	return response, nil
}

// Config is required as input parameter for the constructor creating a new
// cosmosdb client.
type Config struct {
	MasterKey  string
	MaxRetries int
}

type Client struct {
	Url    string
	Config Config
	Client *http.Client
	Log    logging.ExtendedLogger
}

// New makes a new client to communicate to a cosmosdb instance.
// If no http.Client is provided it defaults to the http.DefaultClient
// The log argument can either be an StdLogger (log.Logger), an ExtendedLogger (like logrus.Logger)
// or nil (logging disabled)
func New(url string, cfg Config, cl *http.Client, log logging.StdLogger) *Client {
	client := &Client{
		Url:    strings.Trim(url, "/"),
		Config: cfg,
		Client: cl,
	}

	if client.Client == nil {
		client.Client = http.DefaultClient
	}

	client.Log = logging.Adapt(log)

	return client
}

func (c *Client) get(ctx context.Context, link string, ret interface{}, headers map[string]string) (*http.Response, error) {
	return c.method(ctx, "GET", link, ret, nil, headers)
}

func (c *Client) create(ctx context.Context, link string, body, ret interface{}, headers map[string]string) (*http.Response, error) {
	data, err := stringify(body)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)

	return c.method(ctx, "POST", link, ret, buf, headers)
}

func (c *Client) replace(ctx context.Context, link string, body, ret interface{}, headers map[string]string) (*http.Response, error) {
	data, err := stringify(body)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)

	return c.method(ctx, "PUT", link, ret, buf, headers)
}

func (c *Client) delete(ctx context.Context, link string, headers map[string]string) (*http.Response, error) {
	return c.method(ctx, "DELETE", link, nil, nil, headers)
}

func (c *Client) query(ctx context.Context, link string, body, ret interface{}, headers map[string]string) (*http.Response, error) {
	return c.create(ctx, link, body, ret, headers)
}

func (c *Client) method(ctx context.Context, method, link string, ret interface{}, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, path(c.Url, link), body)
	if err != nil {
		c.Log.Errorln(err)
		return nil, err
	}
	defaultHeaders, err := defaultHeaders(method, link, c.Config.MasterKey)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create request headers")
	}
	if headers == nil {
		headers = map[string]string{}
	}
	for k, v := range defaultHeaders {
		// insert if not already present
		headers[k] = v
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return c.do(ctx, req, ret)
}

func retriable(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable
}

// Request Error
type RequestError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Implement Error function
func (e RequestError) Error() string {
	return fmt.Sprintf("%v, %v", e.Code, e.Message)
}

func (c *Client) checkResponse(resp *http.Response) error {
	if retriable(resp.StatusCode) {
		return errRetry
	}
	if cosmosError, ok := CosmosHTTPErrors[resp.StatusCode]; ok {
		return cosmosError
	}
	return errUnexpectedHTTPStatus

}

// Private Do function, DRY
func (c *Client) do(ctx context.Context, r *http.Request, data interface{}) (*http.Response, error) {
	cli := c.Client
	if cli == nil {
		cli = http.DefaultClient
	}
	if !IgnoreContext {
		r = r.WithContext(ctx)
	}
	// save body to be able to retry the request
	b := []byte{}
	if r.Body != nil {
		var err error
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	var resp *http.Response
	for retryCount := 0; retryCount <= c.Config.MaxRetries; retryCount++ {
		var err error
		if retryCount > 0 {
			delay := backoffDelay(retryCount)
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return nil, ctx.Err()
			case <-t.C:
			}
		}

		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		c.Log.Debugf("Cosmos request: %s %s (headers: %s) (attempt: %d/%d)\n", r.Method, r.URL, r.Header, retryCount+1, c.Config.MaxRetries)
		resp, err = cli.Do(r)
		if err != nil {
			return nil, err
		}
		c.Log.Debugf("Cosmos response: %s (headers: %s)", resp.Status, resp.Header)
		err = c.handleResponse(ctx, r, resp, data)
		if err == errRetry {
			continue
		}
		return resp, err
	}
	return resp, ErrMaxRetriesExceeded
}

func (c Client) handleResponse(ctx context.Context, req *http.Request, resp *http.Response, ret interface{}) error {
	defer resp.Body.Close()
	if ResponseHook != nil {
		ResponseHook(ctx, req.Method, resp.Header)
	}
	err := c.checkResponse(resp)

	if err != nil {
		b, readErr := ioutil.ReadAll(resp.Body)
		if readErr == nil {
			c.Log.Debugln("Error response from Cosmos DB: " + string(b))
		}
		return err
	}

	if ret == nil {
		return nil
	}
	if resp.ContentLength == 0 {
		return nil
	}
	err = readJson(resp.Body, ret)
	// even if JSON parsing failed, we still want to consume all bytes from Body
	// in order to reuse the connection.
	io.Copy(ioutil.Discard, resp.Body)
	return err
}
