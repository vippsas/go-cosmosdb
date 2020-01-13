package cosmosapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

func (c *Client) GetPartitionKeyRanges(
	ctx context.Context,
	databaseName, collectionName string,
	options *GetPartitionKeyRangesOptions,
) (response GetPartitionKeyRangesResponse, err error) {
	link := CreateCollLink(databaseName, collectionName) + "/pkranges"
	var responseBody getPartitionKeyRangesResponseBody
	headers, err := options.AsHeaders()
	if err != nil {
		return response, err
	}
	httpResponse, err := c.get(ctx, link, &responseBody, headers)
	if err != nil {
		return response, err
	}
	response.PartitionKeyRanges = responseBody.PartitionKeyRanges
	response.Id = responseBody.Id
	response.Rid = responseBody.Rid
	err = response.parseHeaders(httpResponse)
	if err != nil {
		return response, errors.WithMessage(err, "Failed to get partition key ranges")
	}
	return response, err
}

type getPartitionKeyRangesResponseBody struct {
	Rid                string              `json:"_rid"`
	Id                 string              `json:"id"`
	PartitionKeyRanges []PartitionKeyRange `json:"PartitionKeyRanges"`
}

type PartitionKeyRange struct {
	Id           string   `json:"id"`
	MaxExclusive string   `json:"maxExclusive"`
	MinInclusive string   `json:"minInclusive"`
	Parents      []string `json:"parents"`
}

type GetPartitionKeyRangesOptions struct {
	MaxItemCount int
	Continuation string
}

func (ops GetPartitionKeyRangesOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}
	if ops.MaxItemCount != 0 {
		headers[HEADER_MAX_ITEM_COUNT] = strconv.Itoa(ops.MaxItemCount)
	}
	if ops.Continuation != "" {
		headers[HEADER_CONTINUATION] = ops.Continuation
	}
	return headers, nil
}

type GetPartitionKeyRangesResponse struct {
	Id                 string
	Rid                string
	PartitionKeyRanges []PartitionKeyRange
	RequestCharge      float64
	SessionToken       string
	Continuation       string
	Etag               string
}

func (r *GetPartitionKeyRangesResponse) parseHeaders(httpResponse *http.Response) error {
	r.SessionToken = httpResponse.Header.Get(HEADER_SESSION_TOKEN)
	r.Continuation = httpResponse.Header.Get(HEADER_CONTINUATION)
	r.Etag = httpResponse.Header.Get(HEADER_ETAG)
	if _, ok := httpResponse.Header[HEADER_REQUEST_CHARGE]; ok {
		requestCharge, err := strconv.ParseFloat(httpResponse.Header.Get(HEADER_REQUEST_CHARGE), 64)
		if err != nil {
			return errors.WithStack(err)
		}
		r.RequestCharge = requestCharge
	}
	return nil
}
