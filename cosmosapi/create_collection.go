package cosmosapi

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type CreateCollectionOptions struct {
	Id             string          `json:"id"`
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`

	// RTUs [400 - 250000]. Do not use in combination with OfferType
	OfferThroughput OfferThroughput `json:"offerThroughput,omitempty"`
	// S1,S2,S3. Do not use in combination with OfferThroughput
	OfferType         OfferType `json:"offerType,omitempty"`
	DefaultTimeToLive int       `json:"defaultTtl,omitempty"`
}

type CreateCollectionResponse struct {
	ResponseBase
	Collection Collection
}

func (colOps CreateCollectionOptions) asHeaders() (map[string]string, error) {
	headers := make(map[string]string)

	if colOps.OfferThroughput > 0 {
		headers[HEADER_OFFER_THROUGHPUT] = fmt.Sprintf("%d", colOps.OfferThroughput)
	}

	if colOps.OfferThroughput >= 10000 && colOps.PartitionKey == nil {
		return nil, ErrThroughputRequiresPartitionKey
	}

	if colOps.OfferType != "" {
		headers[HEADER_OFFER_TYPE] = fmt.Sprintf("%s", colOps.OfferType)
	}

	return headers, nil
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/create-a-collection
func (c *Client) CreateCollection(
	ctx context.Context,
	dbName string,
	colOps CreateCollectionOptions,
) (CreateCollectionResponse, error) {
	response := CreateCollectionResponse{}
	headers, hErr := colOps.asHeaders()
	if hErr != nil {
		return response, hErr
	}

	if colOps.OfferThroughput > 0 {
		headers[HEADER_OFFER_THROUGHPUT] = fmt.Sprintf("%d", colOps.OfferThroughput)
	}

	if colOps.OfferThroughput >= 10000 && colOps.PartitionKey == nil {
		return response, errors.New(fmt.Sprintf("Must specify PartitionKey for collection '%s' when OfferThroughput is >= 10000", colOps.Id))
	}

	if colOps.OfferType != "" {
		headers[HEADER_OFFER_TYPE] = fmt.Sprintf("%s", colOps.OfferType)
	}

	link := CreateCollLink(dbName, "")
	collection := Collection{}

	httpResponse, err := c.create(ctx, link, colOps, &collection, headers)
	if err != nil {
		return response, err
	}
	response.Collection = collection
	return response.parse(httpResponse)
}

func (r CreateCollectionResponse) parse(httpResponse *http.Response) (CreateCollectionResponse, error) {
	responseBase, err := parseHttpResponse(httpResponse)
	r.ResponseBase = responseBase
	return r, err
}
