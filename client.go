package cosmosdb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func init() {
	// Using standard log lib to avoid external dependencies for now.
	// TODO: Consider if debug is neccesary
	log.SetOutput(ioutil.Discard)
}

var (
	// TODO: useful?
	IgnoreContext bool
	// TODO: check thread safety
	ResponseHook            func(ctx context.Context, method string, headers map[string][]string)
	errUnexpectedHTTPStatus = errors.New("Unexpected HTTP return status")
)

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
}

// New makes a new client to communicate to a cosmosdb instance.
// If no http.Client is provided it defaults to the http.DefaultClient
func New(url string, cfg Config, cl *http.Client) *Client {
	client := &Client{
		Url:    strings.Trim(url, "/"),
		Config: cfg,
		Client: cl,
	}

	if client.Client == nil {
		client.Client = http.DefaultClient
	}

	return client
}

func (c *Client) get(ctx context.Context, link string, ret interface{}, headers map[string]string) error {
	_, err := c.method(ctx, "GET", link, ret, nil, headers)
	return err
}

func (c *Client) create(ctx context.Context, link string, body, ret interface{}, headers map[string]string) error {
	data, err := stringify(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)

	_, err = c.method(ctx, "POST", link, ret, buf, headers)
	return err
}

func (c *Client) replace(ctx context.Context, link string, body, ret interface{}, headers map[string]string) error {
	data, err := stringify(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)

	_, err = c.method(ctx, "PUT", link, ret, buf, headers)
	return err
}

func (c *Client) delete(ctx context.Context, link string, headers map[string]string) error {
	_, err := c.method(ctx, "DELETE", link, nil, nil, headers)
	return err
}

func (c *Client) query(ctx context.Context, link string, body, ret interface{}, headers map[string]string) error {
	return c.create(ctx, link, body, ret, headers)
}

func (c *Client) method(ctx context.Context, method, link string, ret interface{}, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, path(c.Url, link), body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Printf("Will call: %s\n", req.URL)
	//r := ResourceRequest(link, req)

	defaultHeaders, err := defaultHeaders(method, link, c.Config.MasterKey)
	if err != nil {
		log.Println(err)
		return nil, err
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

	log.Printf("Headers: %s\n", req.Header)

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

func (c *Client) checkResponse(ctx context.Context, retryCount int, resp *http.Response) error {
	if retriable(resp.StatusCode) {
		if retryCount < c.Config.MaxRetries {
			delay := backoffDelay(retryCount)
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
				return errRetry
			}
		}
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
	retryCount := 0
	for {
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		log.Printf("Executing request\n")
		resp, err := cli.Do(r)
		if err != nil {
			return nil, err
		}
		if ResponseHook != nil {
			ResponseHook(ctx, r.Method, resp.Header)
		}
		err = c.checkResponse(ctx, retryCount, resp)
		if err == errRetry {
			resp.Body.Close()
			retryCount++
			continue
		}
		defer resp.Body.Close()

		if err != nil {
			return resp, err
		}

		if data == nil {
			return resp, nil
		}
		return resp, readJson(resp.Body, data)
	}
}
