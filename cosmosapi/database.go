package cosmosapi

import (
	"context"
)

// Database
type Database struct {
	Resource
	Colls string `json:"_colls,omitempty"`
	Users string `json:"_users,omitempty"`
}

type CreateDatabaseOptions struct {
	ID string `json:"id"`
}

func createDatabaseLink(dbName string) string {
	return "dbs/" + dbName
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/create-a-database
func (c *Client) CreateDatabase(ctx context.Context, dbName string, ops *RequestOptions) (*Database, error) {
	db := &Database{}

	_, err := c.create(ctx, createDatabaseLink(""), CreateDatabaseOptions{dbName}, db, nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c *Client) ListDatabases(ctx context.Context, ops *RequestOptions) ([]Database, error) {
	return nil, ErrorNotImplemented
}

func (c *Client) GetDatabase(ctx context.Context, dbName string, ops *RequestOptions) (*Database, error) {
	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	db := &Database{}

	err := c.get(ctx, createDatabaseLink(dbName), db, nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, dbName string, ops *RequestOptions) error {
	return ErrorNotImplemented
}
