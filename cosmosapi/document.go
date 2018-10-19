package cosmosapi

import (
	"context"
	"net/http"
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
	PartitionKeyValue   interface{}
	IsUpsert            bool
	IndexingDirective   IndexingDirective
	PreTriggersInclude  []string
	PostTriggersInclude []string
}

type DocumentResponse struct {
	RUs          float64
	SessionToken string
}

func parseDocumentResponse(resp *http.Response) (parsed DocumentResponse) {
	parsed.SessionToken = resp.Header.Get(HEADER_SESSION_TOKEN)
	parsed.RUs, _ = strconv.ParseFloat(resp.Header.Get(HEADER_REQUEST_CHARGE), 64)
	return
}

func (ops CreateDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	if ops.PartitionKeyValue != nil {
		v, err := MarshalPartitionKeyHeader(ops.PartitionKeyValue)
		if err != nil {
			return nil, err
		}
		headers[HEADER_PARTITIONKEY] = v
	}

	headers[HEADER_UPSERT] = strconv.FormatBool(ops.IsUpsert)

	if ops.IndexingDirective != "" {
		headers[HEADER_INDEXINGDIRECTIVE] = string(ops.IndexingDirective)
	}

	if ops.PreTriggersInclude != nil && len(ops.PreTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_PRE_INCLUDE] = strings.Join(ops.PreTriggersInclude, ",")
	}

	if ops.PostTriggersInclude != nil && len(ops.PostTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_POST_INCLUDE] = strings.Join(ops.PostTriggersInclude, ",")
	}

	return headers, nil
}

func (c *Client) CreateDocument(ctx context.Context, dbName, colName string,
	doc interface{}, ops CreateDocumentOptions) (*Resource, DocumentResponse, error) {

	// add optional headers (after)
	headers := map[string]string{}
	var err error
	headers, err = ops.AsHeaders()
	if err != nil {
		return nil, DocumentResponse{}, err
	}

	resource := &Resource{}
	link := createDocsLink(dbName, colName)

	response, err := c.create(ctx, link, doc, resource, headers)
	if err != nil {
		return nil, DocumentResponse{}, err
	}
	return resource, parseDocumentResponse(response), nil
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
	PartitionKeyValue interface{}
	ConsistencyLevel  ConsistencyLevel
	SessionToken      string
}

func (ops GetDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	headers[HEADER_IF_NONE_MATCH] = strconv.FormatBool(ops.IfNoneMatch)

	if ops.PartitionKeyValue != nil {
		v, err := MarshalPartitionKeyHeader(ops.PartitionKeyValue)
		if err != nil {
			return nil, err
		}
		headers[HEADER_PARTITIONKEY] = v
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
	ops GetDocumentOptions, out interface{}) error {

	headers, err := ops.AsHeaders()
	if err != nil {
		return err
	}

	link := createDocLink(dbName, colName, id)

	err = c.get(ctx, link, out, headers)
	if err != nil {
		return err
	}

	return nil
}

type ReplaceDocumentOptions struct {
	PartitionKeyValue   interface{}
	IndexingDirective   IndexingDirective
	PreTriggersInclude  []string
	PostTriggersInclude []string
	IfMatch             string
}

func (ops ReplaceDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	if ops.PartitionKeyValue != nil {
		v, err := MarshalPartitionKeyHeader(ops.PartitionKeyValue)
		if err != nil {
			return nil, err
		}
		headers[HEADER_PARTITIONKEY] = v
	}

	if ops.IndexingDirective != "" {
		headers[HEADER_INDEXINGDIRECTIVE] = string(ops.IndexingDirective)
	}

	if ops.PreTriggersInclude != nil && len(ops.PreTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_PRE_INCLUDE] = strings.Join(ops.PreTriggersInclude, ",")
	}

	if ops.PostTriggersInclude != nil && len(ops.PostTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_POST_INCLUDE] = strings.Join(ops.PostTriggersInclude, ",")
	}

	if ops.IfMatch != "" {
		headers[HEADER_IF_MATCH] = ops.IfMatch
	}

	return headers, nil
}

// ReplaceDocument replaces a whole document.
func (c *Client) ReplaceDocument(ctx context.Context, dbName, colName, id string,
	doc interface{}, ops ReplaceDocumentOptions) (*Resource, DocumentResponse, error) {

	headers := map[string]string{}
	var err error
	headers, err = ops.AsHeaders()
	if err != nil {
		return nil, DocumentResponse{}, err
	}

	link := createDocLink(dbName, colName, id)
	resource := &Resource{}

	response, err := c.replace(ctx, link, doc, resource, headers)
	if err != nil {
		return nil, DocumentResponse{}, err
	}

	return resource, parseDocumentResponse(response), nil
}

// DeleteDocumentOptions contains all options that can be used for deleting
// documents.
type DeleteDocumentOptions struct {
	PartitionKeyValue   interface{}
	PreTriggersInclude  []string
	PostTriggersInclude []string
	/* TODO */
}

func (ops DeleteDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	//TODO: DRY
	if ops.PartitionKeyValue != nil {
		v, err := MarshalPartitionKeyHeader(ops.PartitionKeyValue)
		if err != nil {
			return nil, err
		}
		headers[HEADER_PARTITIONKEY] = v
	}

	if ops.PreTriggersInclude != nil && len(ops.PreTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_PRE_INCLUDE] = strings.Join(ops.PreTriggersInclude, ",")
	}

	if ops.PostTriggersInclude != nil && len(ops.PostTriggersInclude) > 0 {
		headers[HEADER_TRIGGER_POST_INCLUDE] = strings.Join(ops.PostTriggersInclude, ",")
	}

	return headers, nil
}

func (c *Client) DeleteDocument(ctx context.Context, dbName, colName, id string, ops DeleteDocumentOptions) error {
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

// QueryDocumentsOptions bundles all options supported by Cosmos DB when
// querying for documents.
type QueryDocumentsOptions struct {
	PartitionKeyValue    interface{}
	IsQuery              bool
	ContentType          string
	MaxItemCount         int
	Continuation         string
	EnableCrossPartition bool
	ConsistencyLevel     ConsistencyLevel
	SessionToken         string
}

const QUERY_CONTENT_TYPE = "application/query+json"

// DefaultQueryDocumentOptions returns QueryDocumentsOptions populated with
// sane defaults. For QueryDocumentsOptions Cosmos DB requires some specific
// options which are not obvious. This function helps to get things right.
func DefaultQueryDocumentOptions() QueryDocumentsOptions {
	return QueryDocumentsOptions{
		IsQuery:     true,
		ContentType: QUERY_CONTENT_TYPE,
	}
}

func (ops QueryDocumentsOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	//TODO: DRY
	if ops.PartitionKeyValue != nil {
		v, err := MarshalPartitionKeyHeader(ops.PartitionKeyValue)
		if err != nil {
			return nil, err
		}
		headers[HEADER_PARTITIONKEY] = v
	} else if ops.EnableCrossPartition {
		headers[HEADER_CROSSPARTITION] = "true"
	}

	headers[HEADER_IS_QUERY] = strconv.FormatBool(ops.IsQuery)

	if ops.ContentType != QUERY_CONTENT_TYPE {
		return nil, ErrWrongQueryContentType
	} else {
		headers[HEADER_CONTYPE] = ops.ContentType
	}

	// TODO: Add missing headers

	return headers, nil
}

// QueryDocuments queries a collection in cosmosdb with the provided query.
// To correctly parse the returned results you currently have to pass in
// a slice for the returned documents, not a single document.
func (c *Client) QueryDocuments(ctx context.Context, dbName, collName string, qry Query, docs interface{}, ops QueryDocumentsOptions) (QueryDocumentsResponse, error) {

	headers, err := ops.AsHeaders()
	if err != nil {
		return QueryDocumentsResponse{}, err
	}

	link := createDocsLink(dbName, collName)

	results := QueryDocumentsResponse{
		Documents: docs,
	}

	response, err := c.query(ctx, link, qry, &results, headers)
	if err != nil {
		return QueryDocumentsResponse{}, err
	}
	results.RUs, _ = strconv.ParseFloat(response.Header.Get(HEADER_REQUEST_CHARGE), 64)
	results.Continuation = response.Header.Get(HEADER_CONTINUATION)

	return results, nil
}
