package cosmosdb

import (
	"context"
)

type Collection struct {
	Resource
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	Docs           string          `json:"_docs,omitempty"`
	Udf            string          `json:"_udfs,omitempty"`
	Sporcs         string          `json:"_sporcs,omitempty"`
	Triggers       string          `json:"_triggers,omitempty"`
	Conflicts      string          `json:"_conflicts,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`
}

type IndexingPolicy struct {
	IndexingMode IndexingMode   `json:"indexingMode,omitempty"`
	Automatic    bool           `json:"automatic"`
	Included     []IncludedPath `json:"includedPaths,omitempty"`
	Excluded     []ExcludedPath `json:"excludedPaths,omitempty"`
}

type IndexingMode string

const (
	Consistent = IndexingMode("Consistent")
	Lazy       = IndexingMode("Lazy")
)

type PartitionKey struct {
	Paths []string `json:"paths"`
	Kind  string   `json:"kind"`
}

type CollectionCreateOptions struct {
	Id             string          `json:"id"`
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`
}

type CollectionReplaceOptions struct {
	Id             string          `json:"id"`
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`
}

func (c *Client) CreateCollection(ctx context.Context, dbName string,
	colOps CollectionCreateOptions, ops *RequestOptions) (*Collection, error) {

	return nil, ErrorNotImplemented
}

func (c *Client) ListCollections(ctx context.Context, dbName string,
	ops *RequestOptions) ([]Collection, error) {
	return nil, ErrorNotImplemented
}

func (c *Client) GetCollection(ctx context.Context, dbName, colName string,
	ops *RequestOptions) (*Collection, error) {
	return nil, ErrorNotImplemented
}

func (c *Client) DeleteCollection(ctx context.Context, dbName, colName string,
	ops *RequestOptions) error {
	return ErrorNotImplemented
}

func (c *Client) ReplaceCollection(ctx context.Context, dbName, colName string,
	colOps CollectionReplaceOptions, ops *RequestOptions) (*Collection, error) {

	return nil, ErrorNotImplemented
}

// TODO: add model for partition key ranges
func (c *Client) GetPartitionKeyRanges(ctx context.Context, dbName, colName string,
	ops *RequestOptions) error {
	return ErrorNotImplemented
}
