package cosmosapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

type ListCollectionsOptions struct {
	MaxItemCount int
	Continuation string
}

type ListCollectionsResponse struct {
	RequestCharge float64
	SessionToken  string
	Continuation  string
	Etag          string
	Collections   DocumentCollection
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/list-collections
func (c *Client) ListCollections(
	ctx context.Context,
	dbName string,
	options ListCollectionsOptions,
) (ListCollectionsResponse, error) {
	url := createDatabaseLink(dbName) + "/colls"
	response := ListCollectionsResponse{}
	headers, err := options.asHeaders()
	if err != nil {
		return response, errors.WithMessage(err, "Failed to list collections")
	}
	docCol := DocumentCollection{}
	httpResponse, err := c.get(ctx, url, &docCol, headers)
	if err != nil {
		return response, errors.WithMessage(err, "Failed to list collections")
	}
	response, err = response.parse(httpResponse)
	if err != nil {
		return response, errors.WithMessage(err, "Failed to list collections")
	}
	response.Collections = docCol
	return response, nil
}

func (ops ListCollectionsOptions) asHeaders() (map[string]string, error) {
	headers := map[string]string{}
	if ops.MaxItemCount != 0 {
		headers[HEADER_MAX_ITEM_COUNT] = strconv.Itoa(ops.MaxItemCount)
	}
	if ops.Continuation != "" {
		headers[HEADER_CONTINUATION] = ops.Continuation
	}
	return headers, nil
}

func (r ListCollectionsResponse) parse(httpResponse *http.Response) (ListCollectionsResponse, error) {
	r.SessionToken = httpResponse.Header.Get(HEADER_SESSION_TOKEN)
	r.Continuation = httpResponse.Header.Get(HEADER_CONTINUATION)
	r.Etag = httpResponse.Header.Get(HEADER_ETAG)
	if _, ok := httpResponse.Header[HEADER_REQUEST_CHARGE]; ok {
		requestCharge, err := strconv.ParseFloat(httpResponse.Header.Get(HEADER_REQUEST_CHARGE), 64)
		if err != nil {
			return r, errors.WithStack(err)
		}
		r.RequestCharge = requestCharge
	}
	return r, nil
}
