package cosmosdb

import (
	"context"
)

// Database
type Database struct {
	Resource
	Colls string `json:"_colls,omitempty"`
	Users string `json:"_users,omitempty"`
}

func (c *Client) GetDatabase(ctx context.Context, link string, ops *RequestOptions) (*Database, error) {
	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	db := &Database{}

	err := c.get(ctx, link, db, nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}
