package cosmosdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Client represents a connection to cosmosdb. Not in the sense of a database
// connection but in the sense of containing all required information to get
var (
	errRetry              = errors.New("retry")
	IgnoreContext         bool
	ErrPreconditionFailed = errors.New("precondition failed")
	ResponseHook          func(ctx context.Context, method string, headers map[string][]string)
)

type queryKey struct{}
type sprocKey struct{}
type collKey struct{}

type Config struct {
	MasterKey  string
	MaxRetries int
}

type QueryParam struct {
	Name  string      `json:"name"` // should contain a @ character
	Value interface{} `json:"value"`
}

type Query struct {
	Text   string       `json:"query"`
	Params []QueryParam `json:"parameters,omitempty"`
	Token  string       `json:"-"` // continuation token
}

type Client struct {
	Url    string
	Config Config
	Client *http.Client
}

func New() (*Client, error) {
	return nil, ErrorNotImplemented
}

type PartitionResolver interface {
	// Returns partition key from a document
	GetPartitionKey()
}

// CreateAttachment
func (c *Client) CreateDatabase(ctx context.Context, uri string, ops *RequestOptions) {}

func (c *Client) CreateDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions) error {

	// add headers: default and options

	// check if collection uses partitions
	//if pkResolve :=  c.collectionRegister[link]; pkResolve != nil {
	//pk :=  pkResolve(doc)
	//}

	c.create(ctx, link, doc, nil, nil)

	return nil
}

//func (c *Client) ListDatabases()                                                      {}
//func (c *Client) GetDatabase()                                                        {}

func do() error {
	return nil
}

type RequestOptions map[RequestOption]string
type RequestOption string

var (
	ReqOpAllowCrossPartition = RequestOption("x-ms-documentdb-query-enablecrosspartition")
)

//type AddressingMode string

//var (
//AddressingModeSelf = AddressingMode("Self")
//AddressingModeUser = AddressingMode("User")
//)

// Create resource
func (c *Client) create(ctx context.Context, link string, body, ret interface{}, headers map[string]string) error {
	data, err := stringify(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)

	_, err = c.method(ctx, "POST", link, ret, buf, headers)
	return err
}

func defaultHeaders(method, link, key string) (map[string]string, error) {
	return nil, ErrorNotImplemented
}

// Private generic method resource
func (c *Client) method(ctx context.Context, method, link string, ret interface{}, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, path(c.Url, link), body)
	if err != nil {
		return nil, err
	}
	//r := ResourceRequest(link, req)

	defaultHeaders, err := defaultHeaders(method, link, c.Config.MasterKey)
	if err != nil {
		return nil, err
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
	if resp.StatusCode == http.StatusPreconditionFailed {
		return ErrPreconditionFailed
	}
	if resp.StatusCode >= 300 {
		err := &RequestError{}
		readJson(resp.Body, &err)
		return err
	}

	return nil
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

func backoffDelay(retryCount int) time.Duration {
	minTime := 300

	if retryCount > 13 {
		retryCount = 13
	} else if retryCount > 8 {
		retryCount = 8
	}

	delay := (1 << uint(retryCount)) * (rand.Intn(minTime) + minTime)
	return time.Duration(delay) * time.Millisecond
}

// Generate link
func path(url string, args ...string) (link string) {
	args = append([]string{url}, args...)
	link = strings.Join(args, "/")
	return
}

// Read json response to given interface(struct, map, ..)
func readJson(reader io.Reader, data interface{}) error {
	return json.NewDecoder(reader).Decode(data)
}

// Stringify body data
func stringify(body interface{}) (bt []byte, err error) {
	switch t := body.(type) {
	case string:
		bt = []byte(t)
	case []byte:
		bt = t
	default:
		bt, err = json.Marshal(t)
	}
	return
}
