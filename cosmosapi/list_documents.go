package cosmosapi

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

// ListDocument reads either all documents or the incremental feed, aka. change feed.
func (c *Client) ListDocuments(
	ctx context.Context,
	databaseName, collectionName string,
	options *ListDocumentsOptions,
	documentList interface{},
) (response ListDocumentsResponse, err error) {
	link := createDocsLink(databaseName, collectionName)
	var responseBody listDocumentsResponseBody
	headers, err := options.AsHeaders()
	if err != nil {
		return response, err
	}
	httpResponse, err := c.get(ctx, link, &responseBody, headers)
	if err != nil {
		return response, err
	} else if httpResponse.StatusCode == http.StatusNotModified {
		return response, err
	} else if err = json.Unmarshal(responseBody.Documents, documentList); err != nil {
		return response, err
	}
	r, err := response.parse(httpResponse)
	return *r, err
}

type listDocumentsResponseBody struct {
	Rid       string          `json:"_rid"`
	Count     int             `json:"_count"`
	Documents json.RawMessage `json:"Documents"`
}

type ListDocumentsOptions struct {
	MaxItemCount        int
	AIM                 string
	Continuation        string
	IfNoneMatch         string
	PartitionKeyRangeId string
}

func (ops ListDocumentsOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}
	if ops.MaxItemCount != 0 {
		headers[HEADER_MAX_ITEM_COUNT] = strconv.Itoa(ops.MaxItemCount)
	}
	if ops.AIM != "" {
		headers[HEADER_A_IM] = ops.AIM
	}
	if ops.Continuation != "" {
		headers[HEADER_CONTINUATION] = ops.Continuation
	}
	if ops.IfNoneMatch != "" {
		headers[HEADER_IF_NONE_MATCH] = ops.IfNoneMatch
	}
	if ops.PartitionKeyRangeId != "" {
		headers[HEADER_PARTITION_KEY_RANGE_ID] = ops.PartitionKeyRangeId
	}
	return headers, nil
}

type ListDocumentsResponse struct {
	RequestCharge float64
	SessionToken  string
	Continuation  string
	Etag          string
}

func (r *ListDocumentsResponse) parse(httpResponse *http.Response) (*ListDocumentsResponse, error) {
	r.SessionToken = httpResponse.Header.Get(HEADER_SESSION_TOKEN)
	r.Continuation = httpResponse.Header.Get(HEADER_CONTINUATION)
	r.Etag = httpResponse.Header.Get(HEADER_ETAG)
	if _, ok := httpResponse.Header[HEADER_REQUEST_CHARGE]; ok {
		if requestCharge, err := strconv.ParseFloat(httpResponse.Header.Get(HEADER_REQUEST_CHARGE), 64); err != nil {
			return r, errors.WithStack(err)
		} else {
			r.RequestCharge = requestCharge
		}
	}
	return r, nil
}
