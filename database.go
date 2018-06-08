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

// Collection returns retrieves a collection from a cosmos db instance and returns
// it as a Collection struct.
func (db Database) Collection(ctx context.Context, name string) (*Collection, error) {
	return nil, ErrorNotImplemented
}

func (db Database) CreateCollection(ctx context.Context, doc interface{}) (*Collection, error) {
	return nil, ErrorNotImplemented
}
