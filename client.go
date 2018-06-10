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
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	HEADER_XDATE        = "X-Ms-Date"
	HEADER_AUTH         = "Authorization"
	HEADER_VER          = "X-Ms-Version"
	HEADER_CONTYPE      = "Content-Type"
	HEADER_CONLEN       = "Content-Length"
	HEADER_IS_QUERY     = "X-Ms-Documentdb-Isquery"
	HEADER_UPSERT       = "X-Ms-Documentdb-Is-Upsert"
	HEADER_CONTINUATION = "X-Ms-Continuation"
	HEADER_IF_MATCH     = "If-Match"
	HEADER_CHARGE       = "X-Ms-Request-Charge"

	HEADER_CROSSPARTITION = "x-ms-documentdb-query-enablecrosspartition"
	HEADER_PARTITIONKEY   = "x-ms-documentdb-partitionkey"
)

// Client represents a connection to cosmosdb. Not in the sense of a database
// connection but in the sense of containing all required information to get
var (
	errRetry              = errors.New("retry")
	IgnoreContext         bool
	ErrPreconditionFailed = errors.New("precondition failed")
	ResponseHook          func(ctx context.Context, method string, headers map[string][]string)
)

type Config struct {
	MasterKey  string
	MaxRetries int
}

type Client struct {
	Url    string
	Config Config
	Client *http.Client
}

func CreateDatabaseLink(dbName string) string {
	return "dbs/" + dbName
}

func CreateCollLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName
}

func CreateDocsLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/docs"
}

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

func (c *Client) GetDatabase(ctx context.Context, link string, ops *RequestOptions) (*Database, error) {
	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	db := &Database{}

	err := c.get(ctx, link, db, nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c *Client) GetDocument(ctx context.Context, link string, ops *RequestOptions, out *interface{}) error {
	return ErrorNotImplemented
}

func (c *Client) CreateDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions) error {
	fmt.Printf("CreateDocument: %s\n", link)

	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	fmt.Printf("CreateDocument: Headers: %s\n", headers)

	err := c.create(ctx, link, doc, nil, headers)
	if err != nil {
		return err
	}

	return nil
}

type RequestOptions map[RequestOption]string
type RequestOption string

var (
	ReqOpAllowCrossPartition = RequestOption("x-ms-documentdb-query-enablecrosspartition")
	ReqOpPartitionKey        = RequestOption(HEADER_PARTITIONKEY)
)

func (c *Client) get(ctx context.Context, link string, ret interface{}, headers map[string]string) error {
	_, err := c.method(ctx, "GET", link, ret, nil, headers)
	return err
}

// Create resource
func (c *Client) create(ctx context.Context, link string, body, ret interface{}, headers map[string]string) error {
	data, err := stringify(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	fmt.Printf("Request body: \n%s\n", data)

	fmt.Printf("Will call c.method\n")
	_, err = c.method(ctx, "POST", link, ret, buf, headers)
	return err
}

type AuthorizationPayload struct {
	Verb         string
	ResourceType string
	ResourceLink string
	Date         string
}

func stringToSign(p AuthorizationPayload) string {
	return strings.ToLower(p.Verb) + "\n" +
		strings.ToLower(p.ResourceType) + "\n" +
		p.ResourceLink + "\n" +
		strings.ToLower(p.Date) + "\n" +
		"" + "\n"
}

func makeSignedPayload(verb, link, date, key string) (string, error) {
	if strings.HasPrefix(link, "/") == true {
		link = link[1:]
	}

	rLink, rType := resourceTypeFromLink(verb, link)

	pl := AuthorizationPayload{
		Verb:         verb,
		ResourceType: rType,
		ResourceLink: rLink,
		Date:         date,
	}

	s := stringToSign(pl)
	fmt.Printf("payload to sign: %s\n", s)

	return authorize(s, key)
}

func makeAuthHeader(sPayload string) string {
	masterToken := "master"
	tokenVersion := "1.0"
	return url.QueryEscape(
		"type=" + masterToken + "&ver=" + tokenVersion + "&sig=" + sPayload,
	)
}

func defaultHeaders(method, link, key string) (map[string]string, error) {
	h := map[string]string{}
	h[HEADER_XDATE] = time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	h[HEADER_VER] = "2017-02-22"
	//h[HEADER_CROSSPARTITION] = "true"

	sign, err := makeSignedPayload(method, link, h[HEADER_XDATE], key)
	if err != nil {
		return h, err
	}

	masterToken := "master"
	tokenVersion := "1.0"
	h[HEADER_AUTH] = url.QueryEscape("type=" + masterToken + "&ver=" + tokenVersion + "&sig=" + sign)
	fmt.Printf("Auth header: %s\n", h[HEADER_AUTH])

	return h, nil
}

// Private generic method resource
func (c *Client) method(ctx context.Context, method, link string, ret interface{}, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, path(c.Url, link), body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("Will call: %s\n", req.URL)
	//r := ResourceRequest(link, req)

	defaultHeaders, err := defaultHeaders(method, link, c.Config.MasterKey)
	if err != nil {
		fmt.Println(err)
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

	fmt.Printf("Headers: %s\n", req.Header)

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
		fmt.Printf("Executing request\n")
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

func resourceTypeFromLink(verb, link string) (rLink, rType string) {
	if strings.HasPrefix(link, "/") == false {
		link = "/" + link
	}
	if strings.HasSuffix(link, "/") == false {
		link = link + "/"
	}

	parts := strings.Split(link, "/")
	l := len(parts)

	switch verb {
	case "GET":
		if l%2 == 0 {
			rLink = strings.Join(parts[1:l-1], "/")
			rType = parts[l-3]
		} else {
			rLink = strings.Join(parts[1:l-1], "/")
			rType = parts[l-2]
		}
	case "POST":
		if l%2 == 0 {
			rLink = strings.Join(parts[1:l-2], "/")
			rType = parts[l-3]
		} else {
			rLink = strings.Join(parts[1:l-2], "/")
			rType = parts[l-2]
		}

	default:
		if l%2 == 0 {
			rLink = strings.Join(parts[0:l-2], "/")
			rType = parts[l-3]
		} else {
			//rLink = strings.Join(parts[0:l-2], "/")
			rLink = link
			rType = parts[l-2]
		}
	}

	return
}

// Get path and return resource Id and Type
// (e.g: "/dbs/b5NCAA==/" ==> "b5NCAA==", "dbs")
func parse(id string) (rId, rType string) {
	if strings.HasPrefix(id, "/") == false {
		id = "/" + id
	}
	if strings.HasSuffix(id, "/") == false {
		id = id + "/"
	}

	parts := strings.Split(id, "/")
	l := len(parts)

	if l%2 == 0 {
		rId = parts[l-2]
		rType = parts[l-3]
	} else {
		rId = parts[l-3]
		rType = parts[l-2]
	}
	return
}
