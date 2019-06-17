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
	} else if err = unmarshalDocuments(responseBody.Documents, documentList); err != nil {
		return response, err
	}
	r, err := response.parse(httpResponse)
	return *r, err
}

func unmarshalDocuments(bytes []byte, documentList interface{}) error {
	if len(bytes) == 0 {
		return nil
	}
	return errors.Wrapf(json.Unmarshal(bytes, documentList), "Error unmarshaling <%s>", string(bytes))
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
	ResponseBase
	SessionToken string
	Continuation string
	Etag         string
}

func (r *ListDocumentsResponse) parse(httpResponse *http.Response) (*ListDocumentsResponse, error) {
	r.SessionToken = httpResponse.Header.Get(HEADER_SESSION_TOKEN)
	r.Continuation = httpResponse.Header.Get(HEADER_CONTINUATION)
	r.Etag = httpResponse.Header.Get(HEADER_ETAG)
	rb, err := parseHttpResponse(httpResponse)
	r.ResponseBase = rb
	return r, err
}
