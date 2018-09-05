package cosmosdb

import (
	"encoding/json"
	"io"
	"math/rand"
	"strings"
	"time"
)

const (
	// Request headers
	HEADER_XDATE             = "X-Ms-Date"
	HEADER_AUTH              = "Authorization"
	HEADER_VER               = "X-Ms-Version"
	HEADER_CONTYPE           = "Content-Type"
	HEADER_CONLEN            = "Content-Length"
	HEADER_IS_QUERY          = "x-ms-documentdb-isquery"
	HEADER_UPSERT            = "x-Ms-Documentdb-Is-Upsert"
	HEADER_IF_MATCH          = "If-Match"
	HEADER_IF_NONE_MATCH     = "If-None-Match"
	HEADER_CHARGE            = "X-Ms-Request-Charge"
	HEADER_CONSISTENCY_LEVEL = "x-ms-consistency-level"
	HEADER_OFFER_THROUGHPUT  = "x-ms-offer-throughput"
	HEADER_OFFER_TYPE        = "x-ms-offer-type"

	HEADER_CROSSPARTITION       = "x-ms-documentdb-query-enablecrosspartition"
	HEADER_PARTITIONKEY         = "x-ms-documentdb-partitionkey"
	HEADER_INDEXINGDIRECTIVE    = "x-ms-indexing-directive"
	HEADER_TRIGGER_PRE_INCLUDE  = "x-ms-documentdb-pre-trigger-include"
	HEADER_TRIGGER_PRE_EXCLUDE  = "x-ms-documentdb-pre-trigger-exclude"
	HEADER_TRIGGER_POST_INCLUDE = "x-ms-documentdb-post-trigger-include"
	HEADER_TRIGGER_POST_EXCLUDE = "x-ms-documentdb-post-trigger-exclude"

	// Both request and response
	HEADER_SESSION_TOKEN = "x-ms-session-token"
	HEADER_CONTINUATION  = "x-ms-continuation"

	// Response headers
	HEADER_REQUEST_CHARGE = "x-ms-request-charge"
)

type RequestOptions map[RequestOption]string

type RequestOption string

var (
	ReqOpAllowCrossPartition = RequestOption("x-ms-documentdb-query-enablecrosspartition")
	ReqOpPartitionKey        = RequestOption(HEADER_PARTITIONKEY)
)

// defaultHeaders returns a map containing the default headers required
// for all requests to the cosmos db api.
func defaultHeaders(method, link, key string) (map[string]string, error) {
	h := map[string]string{}
	h[HEADER_XDATE] = time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	h[HEADER_VER] = "2017-02-22" // TODO: move to package level

	sign, err := signedPayload(method, link, h[HEADER_XDATE], key)
	if err != nil {
		return h, err
	}

	h[HEADER_AUTH] = authHeader(sign)

	return h, nil
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
