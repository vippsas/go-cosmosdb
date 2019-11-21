package cosmosapi

import (
	"context"
)

type StoredProcedure struct {
	Resource
	Body string `json:"body"`
}

type StoredProcedures struct {
	Resource
	StoredProcedures []StoredProcedure `json:"StoredProcedures"`
	Count            int               `json:"_count,omitempty"`
}

func newSproc(name, body string) *StoredProcedure {
	return &StoredProcedure{
		Resource{Id: name},
		body,
	}
}

func (c *Client) CreateStoredProcedure(
	ctx context.Context, dbName, colName, sprocName, body string,
) (*StoredProcedure, error) {
	ret := &StoredProcedure{}
	link := createSprocsLink(dbName, colName)

	_, err := c.create(ctx, link, newSproc(sprocName, body), ret, nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) ReplaceStoredProcedure(
	ctx context.Context, dbName, colName, sprocName, body string) (*StoredProcedure, error) {
	ret := &StoredProcedure{}
	link := createSprocLink(dbName, colName, sprocName)

	_, err := c.replace(ctx, link, newSproc(sprocName, body), ret, nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) DeleteStoredProcedure(ctx context.Context, dbName, colName, sprocName string) error {
	_, err := c.delete(ctx, createSprocLink(dbName, colName, sprocName), nil)
	return err
}

func (c *Client) GetStoredProcedure(ctx context.Context, dbName, colName, sprocName string) (*StoredProcedure, error) {
	ret := &StoredProcedure{}
	link := createSprocLink(dbName, colName, sprocName)

	_, err := c.get(ctx, link, ret, nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) ListStoredProcedures(ctx context.Context, dbName, colName string) (*StoredProcedures, error) {
	ret := &StoredProcedures{}
	link := createSprocsLink(dbName, colName)

	_, err := c.get(ctx, link, ret, nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type ExecuteStoredProcedureOptions struct {
	PartitionKeyValue interface{}
}

// ExecuteStoredProcedureResponse returns metadata about the procedure call.
type ExecuteStoredProcedureResponse struct {
	ResponseBase
}

func (ops ExecuteStoredProcedureOptions) AsHeaders() (map[string]string, error) {
	headers := make(map[string]string)
	if ops.PartitionKeyValue != nil {
		v, err := MarshalPartitionKeyHeader(ops.PartitionKeyValue)
		if err != nil {
			return nil, err
		}
		headers[HEADER_PARTITIONKEY] = v
	}
	return headers, nil
}

// ExecuteStoredProcedure executes the procedure and returns metadata about the
// call, or error if the procedure failed.
func (c *Client) ExecuteStoredProcedure(
	ctx context.Context, dbName, colName, sprocName string,
	ops ExecuteStoredProcedureOptions,
	ret interface{}, args ...interface{},
) (*ExecuteStoredProcedureResponse, error) {
	headers, err := ops.AsHeaders()
	if err != nil {
		return nil, err
	}
	link := createSprocLink(dbName, colName, sprocName)
	httpResp, err := c.create(ctx, link, args, ret, headers)
	if err != nil {
		return nil, err
	}
	resp := &ExecuteStoredProcedureResponse{}
	resp.ResponseBase, err = parseHttpResponse(httpResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
