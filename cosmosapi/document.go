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

type GetDocumentOptions struct {
	IfNoneMatch       string
	PartitionKeyValue interface{}
	ConsistencyLevel  ConsistencyLevel
	SessionToken      string
}

func (ops GetDocumentOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}

	headers[HEADER_IF_NONE_MATCH] = ops.IfNoneMatch

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
	ops GetDocumentOptions, out interface{}) (DocumentResponse, error) {
	headers, err := ops.AsHeaders()
	if err != nil {
		return DocumentResponse{}, err
	}

	link := createDocLink(dbName, colName, id)

	resp, err := c.get(ctx, link, out, headers)
	if err != nil {
		return DocumentResponse{}, err
	}
	return parseDocumentResponse(resp), nil
}

type ReplaceDocumentOptions struct {
	PartitionKeyValue   interface{}
	IndexingDirective   IndexingDirective
	PreTriggersInclude  []string
	PostTriggersInclude []string
	IfMatch             string
	ConsistencyLevel    ConsistencyLevel
	SessionToken        string
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

	if ops.ConsistencyLevel != "" {
		headers[HEADER_CONSISTENCY_LEVEL] = string(ops.ConsistencyLevel)
	}

	if ops.SessionToken != "" {
		headers[HEADER_SESSION_TOKEN] = ops.SessionToken
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

	// TODO: DRY
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

func (c *Client) DeleteDocument(ctx context.Context, dbName, colName, id string, ops DeleteDocumentOptions) (DocumentResponse, error) {
	headers, err := ops.AsHeaders()
	if err != nil {
		return DocumentResponse{}, err
	}

	link := createDocLink(dbName, colName, id)

	resp, err := c.delete(ctx, link, headers)
	if err != nil {
		return DocumentResponse{}, err
	}

	return parseDocumentResponse(resp), nil
}
