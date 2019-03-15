package cosmosapi

import (
	"context"
	"net/http"
	"strconv"
)

type Query struct {
	Query  string       `json:"query"`
	Params []QueryParam `json:"parameters,omitempty"`
	Token  string       `json:"-"` // continuation token
}

type QueryParam struct {
	Name  string      `json:"name"` // should contain a @ character
	Value interface{} `json:"value"`
}

// TODO: add missing fields
type QueryDocumentsResponse struct {
	ResponseBase
	Documents    interface{}
	Count        int `json:"_count"`
	Continuation string
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

// QueryDocuments queries a collection in cosmosdb with the provided query.
// To correctly parse the returned results you currently have to pass in
// a slice for the returned documents, not a single document.
func (c *Client) QueryDocuments(ctx context.Context, dbName, collName string, qry Query, docs interface{}, ops QueryDocumentsOptions) (QueryDocumentsResponse, error) {
	response := QueryDocumentsResponse{}
	headers, err := ops.asHeaders()
	if err != nil {
		return response, err
	}
	link := createDocsLink(dbName, collName)
	response.Documents = docs
	httpResponse, err := c.query(ctx, link, qry, &response, headers)
	if err != nil {
		return response, err
	}
	return response.parse(httpResponse)
}

// DefaultQueryDocumentOptions returns QueryDocumentsOptions populated with
// sane defaults. For QueryDocumentsOptions Cosmos DB requires some specific
// options which are not obvious. This function helps to get things right.
func DefaultQueryDocumentOptions() QueryDocumentsOptions {
	return QueryDocumentsOptions{
		IsQuery:     true,
		ContentType: QUERY_CONTENT_TYPE,
	}
}

func (ops QueryDocumentsOptions) asHeaders() (map[string]string, error) {
	headers := map[string]string{}

	// TODO: DRY
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

	if ops.MaxItemCount != 0 {
		headers[HEADER_MAX_ITEM_COUNT] = strconv.Itoa(ops.MaxItemCount)
	}

	if ops.Continuation != "" {
		headers[HEADER_CONTINUATION] = ops.Continuation
	}

	if ops.EnableCrossPartition == true {
		headers[HEADER_CROSSPARTITION] = strconv.FormatBool(ops.EnableCrossPartition)
	}

	if ops.ConsistencyLevel != "" {
		headers[HEADER_CONSISTENCY_LEVEL] = string(ops.ConsistencyLevel)
	}

	if ops.SessionToken != "" {
		headers[HEADER_SESSION_TOKEN] = ops.SessionToken
	}

	return headers, nil
}

func (r QueryDocumentsResponse) parse(httpResponse *http.Response) (QueryDocumentsResponse, error) {
	responseBase, err := parseHttpResponse(httpResponse)
	r.ResponseBase = responseBase
	r.Continuation = httpResponse.Header.Get(HEADER_CONTINUATION)
	return r, err
}
