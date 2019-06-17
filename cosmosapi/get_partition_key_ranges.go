package cosmosapi

import (
	"context"
)

func (c *Client) GetPartitionKeyRanges(
	ctx context.Context,
	databaseName, collectionName string,
	options *GetPartitionKeyRangesOptions,
) (response GetPartitionKeyRangesResponse, err error) {
	link := CreateCollLink(databaseName, collectionName) + "/pkranges"
	var responseBody getPartitionKeyRangesResponseBody
	if headers, err := options.AsHeaders(); err != nil {
		return response, err
	} else if _, err := c.get(ctx, link, &responseBody, headers); err != nil {
		return response, err
	} else {
		response.PartitionKeyRanges = responseBody.PartitionKeyRanges
		response.Id = responseBody.Id
		response.Rid = responseBody.Rid
		return response, err
	}
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
}

func (ops GetPartitionKeyRangesOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}
	return headers, nil
}

type GetPartitionKeyRangesResponse struct {
	Id                 string
	Rid                string
	PartitionKeyRanges []PartitionKeyRange
}
