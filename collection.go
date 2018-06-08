package cosmosdb

import (
	"context"
)

// Collection
type Collection struct {
	Resource
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	Docs           string          `json:"_docs,omitempty"`
	Udf            string          `json:"_udfs,omitempty"`
	Sporcs         string          `json:"_sporcs,omitempty"`
	Triggers       string          `json:"_triggers,omitempty"`
	Conflicts      string          `json:"_conflicts,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`
}

// Collection returns retrieves a collection from a cosmos db instance and returns
// it as a Collection struct.
//
// The method uses reflection to search for either a `_self` link field, `_rid`
// field or `id` field. If any one of them is found, it is used to fetch the
// collection from the database.
func (c Collection) CreateDocument(ctx context.Context, doc interface{}, out *interface{}) (Document, error) {
	return Document{}, nil
}

// List Documents returns a closure that can be repeatedly called
// TODO: maybe that could be a reader instead?
func (c Collection) ListDocuments(ctx context.Context) func() ([]Document, error) {
	return func() ([]Document, error) {
		return []Document{Document{}}, nil
	}
}

func (c Collection) Document(ctx context.Context, doc, out interface{}) error {
	return ErrorNotImplemented
}

func (c Collection) DeleteDocument(ctx context.Context, doc interface{}) error {
	return ErrorNotImplemented
}
