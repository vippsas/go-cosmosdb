package cosmosapi

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrThroughputRequiresPartitionKey = errors.New("Must specify PartitionKey when OfferThroughput is >= 10000")
)

type Collection struct {
	Resource
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	Docs           string          `json:"_docs,omitempty"`
	Udf            string          `json:"_udfs,omitempty"`
	Sprocs         string          `json:"_sprocs,omitempty"`
	Triggers       string          `json:"_triggers,omitempty"`
	Conflicts      string          `json:"_conflicts,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`
}

type DocumentCollection struct {
	Rid                 string       `json:"_rid,omitempty"`
	Count               int32        `json:"_count,omitempty"`
	DocumentCollections []Collection `json:"DocumentCollections"`
}

type IndexingPolicy struct {
	IndexingMode IndexingMode     `json:"indexingMode,omitempty"`
	Automatic    bool             `json:"automatic"`
	Included     []IncludedPath   `json:"includedPaths,omitempty"`
	Excluded     []ExcludedPath   `json:"excludedPaths,omitempty"`
	Composite    []CompositeIndex `json:"compositeIndexes,omitempty"`
}

type IndexingMode string

//const (
//	Consistent = IndexingMode("Consistent")
//	Lazy       = IndexingMode("Lazy")
//)
//
//const (
//	OfferTypeS1 = OfferType("S1")
//	OfferTypeS2 = OfferType("S2")
//	OfferTypeS3 = OfferType("S3")
//)

type PartitionKey struct {
	Paths []string `json:"paths"`
	Kind  string   `json:"kind"`
}

type CollectionReplaceOptions struct {
	Resource
	Id                string          `json:"id"`
	IndexingPolicy    *IndexingPolicy `json:"indexingPolicy,omitempty"`
	PartitionKey      *PartitionKey   `json:"partitionKey,omitempty"`
	DefaultTimeToLive int             `json:"defaultTtl,omitempty"`
}

func (c *Client) GetCollection(ctx context.Context, dbName, colName string) (*Collection, error) {
	collection := &Collection{}
	link := CreateCollLink(dbName, colName)
	_, err := c.get(ctx, link, collection, nil)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func (c *Client) DeleteCollection(ctx context.Context, dbName, colName string) error {
	_, err := c.delete(ctx, CreateCollLink(dbName, colName), nil)
	return err
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/replace-a-collection
func (c *Client) ReplaceCollection(ctx context.Context, dbName string,
	colOps CollectionReplaceOptions) (*Collection, error) {

	collection := &Collection{}
	link := CreateCollLink(dbName, colOps.Id)

	_, err := c.replace(ctx, link, colOps, collection, nil)
	if err != nil {
		return nil, err
	}

	return collection, nil
}
