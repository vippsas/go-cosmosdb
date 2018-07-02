package cosmosdb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Document
type Document struct {
	Resource
	Attachments string `json:"attachments,omitempty"`
}

type IndexingDirective string
type ConsistencyLevel string

const (
	IndexingDirectiveInclude = IndexingDirective("include")
	IndexingDirectiveExclude = IndexingDirective("exclude")

	ConsistencyLevelStrong   = ConsistencyLevel("Strong")
	ConsistencyLevelBounded  = ConsistencyLevel("Bounded")
	ConsistencyLevelSession  = ConsistencyLevel("Session")
	ConsistencyLevelEventual = ConsistencyLevel("Eventual")
)

type CreateDocumentOptions struct {
	PartitionKeyValue   string
	IsUpsert            bool
	IndexingDirective   IndexingDirective
	PreTriggersInclude  []string
	PostTriggersInclude []string
}

func (ops CreateDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	if ops.PartitionKeyValue != "" {
		headers[HEADER_PARTITIONKEY] = fmt.Sprintf("[\"%s\"]", ops.PartitionKeyValue)
	}

	headers[HEADER_UPSERT] = strconv.FormatBool(ops.IsUpsert)

	if ops.IndexingDirective != "" {
		headers[HEADER_INDEXINGDIRECTIVE] = string(ops.IndexingDirective)
	}

	return headers, nil
}

func (c *Client) CreateDocument(ctx context.Context, dbName, colName string,
	doc interface{}, ops *CreateDocumentOptions) (*Resource, error) {

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
	PreTriggersInclude  []string
	PostTriggersInclude []string
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

type GetDocumentOptions struct {
	IfNoneMatch       bool
	PartitionKeyValue string
	ConsistencyLevel  ConsistencyLevel
	SessionToken      string
}

func (ops GetDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	headers[HEADER_IF_NONE_MATCH] = strconv.FormatBool(ops.IfNoneMatch)

	if ops.PartitionKeyValue != "" {
		headers[HEADER_PARTITIONKEY] = fmt.Sprintf("[\"%s\"]", ops.PartitionKeyValue)
	}

	if ops.ConsistencyLevel != "" {
		headers[HEADER_CONSISTENCY_LEVEL] = string(ops.ConsistencyLevel)
	}

	if ops.SessionToken != "" {
		headers[HEADER_SESSION_TOKEN] = ops.SessionToken
	}

	return headers, nil
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

	link := createDocLink(dbName, colName, id)

	err := c.get(ctx, link, out, headers)
	if err != nil {
		return err
	}

	return nil
}

type ReplaceDocumentOptions struct {
	PreTriggersInclude  []string
	PostTriggersInclude []string
	/* TODO */
}

// ReplaceDocument replaces a whole document.
func (c *Client) ReplaceDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions, out interface{}) error {
	return ErrorNotImplemented
}

// DeleteDocumentOptions contains all options that can be used for deleting
// documents.
type DeleteDocumentOptions struct {
	PartitionKeyValue   string
	PreTriggersInclude  []string
	PostTriggersInclude []string
	/* TODO */
}

func (ops DeleteDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	//TODO: DRY
	if ops.PartitionKeyValue != "" {
		headers[HEADER_PARTITIONKEY] = fmt.Sprintf("[\"%s\"]", ops.PartitionKeyValue)
	}

	if ops.PreTriggersInclude != nil && len(ops.PreTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_PRE_INCLUDE] = strings.Join(ops.PreTriggersInclude, ",")
	}

	if ops.PostTriggersInclude != nil && len(ops.PostTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_POST_INCLUDE] = strings.Join(ops.PostTriggersInclude, ",")
	}

	return headers, nil
}

func (c *Client) DeleteDocument(ctx context.Context, dbName, colName, id string, ops *DeleteDocumentOptions) error {
	headers, err := ops.AsHeaders()
	if err != nil {
		return err
	}

	link := createDocLink(dbName, colName, id)

	err = c.delete(ctx, link, headers)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) QueryDocuments(ctx context.Context, link string, qry Query, ops *RequestOptions) error {
	return ErrorNotImplemented
}
