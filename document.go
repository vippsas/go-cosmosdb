package cosmosdb

import (
	"context"
	"fmt"
	"strconv"
)

// Document
type Document struct {
	Resource
	Attachments string `json:"attachments,omitempty"`
}

type IndexingDirective string

const (
	IndexingDirectiveInclude = IndexingDirective("include")
	IndexingDirectiveExclude = IndexingDirective("exclude")
)

type CreateDocumentOptions struct {
	PartitionKeyValue *string
	IsUpsert          *bool
	IndexingDirective *IndexingDirective

	// Optional, not sure if this is a good idea
	// could be useful to know if the collection requires a partition key or not.
	Collection Collection
}

func (ops CreateDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	if ops.PartitionKeyValue != nil {
		headers[HEADER_PARTITIONKEY] = fmt.Sprintf("[\"%s\"]", *ops.PartitionKeyValue)
	}

	if ops.IsUpsert != nil {
		headers[HEADER_UPSERT] = strconv.FormatBool(*ops.IsUpsert)
	}

	if ops.IndexingDirective != nil {
		headers[HEADER_INDEXINGDIRECTIVE] = string(*ops.IndexingDirective)
	}

	return headers, nil
}

func (c *Client) CreateDocument(ctx context.Context, dbName, colName string,
	doc interface{}, ops *CreateDocumentOptions) (*Resource, error) {

	// add optional headers (before)
	//headers := map[string]string{}

	//if ops != nil {
	//for k, v := range *ops {
	//headers[string(k)] = v
	//}
	//}

	// add optional headers (after)
	headers := map[string]string{}
	var err error
	if ops != nil {
		headers, err = ops.AsHeaders()
		if err != nil {
			return nil, err
		}
	}

	resource := &Resource{}
	link := CreateDocsLink(dbName, colName)

	err = c.create(ctx, link, doc, resource, headers)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

type UpsertDocumentOptions struct {
	/* TODO */
}

func (c *Client) UpsertDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions) error {
	return ErrorNotImplemented
}

// ListDocument reads either all documents or the incremental feed, aka.
// change feed.
// TODO: probably have to return continuation token for the feed
func (c *Client) ListDocument(ctx context.Context, link string,
	ops *RequestOptions, out interface{}) error {
	return ErrorNotImplemented
}

func (c *Client) GetDocument(ctx context.Context, dbName, colName, id string,
	ops *RequestOptions, out interface{}) error {

	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	link := CreateDocLink(dbName, colName, id)

	err := c.get(ctx, link, out, headers)
	if err != nil {
		return err
	}

	return nil
}

// ReplaceDocument replaces a whole document.
func (c *Client) ReplaceDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions, out interface{}) error {
	return ErrorNotImplemented
}

func (c *Client) DeleteDocument(ctx context.Context, dbName, colName, id string, ops *RequestOptions) error {
	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	link := CreateDocLink(dbName, colName, id)

	err := c.delete(ctx, link, headers)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) QueryDocuments(ctx context.Context, link string, qry Query, ops *RequestOptions) error {
	return ErrorNotImplemented
}
