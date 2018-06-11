package cosmosdb

import (
	"encoding/json"
	"io"
	"math/rand"
	"strings"
	"time"
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
