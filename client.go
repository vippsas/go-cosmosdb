package cosmosdb

import (
	"context"
)

// Client represents a connection to cosmosdb. Not in the sense of a database
// connection but in the sense of containing all required information to get

type Client struct {
}

func (db *Client) Collection(ctx context.Context, doc interface{}) (Collection, error) {
	return Collection{}, nil
}

type Collection struct {
}

func (c Collection) CreateDocument(ctx context.Context, doc interface{}) (Document, error) {
	return Document{}, nil
}

func (c Collection) DeleteDocument(ctx context.Context, doc interface{}) error { return nil }

type Document struct {
}
