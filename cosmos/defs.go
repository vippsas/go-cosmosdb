package cosmos

import (
	"context"

	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

type BaseModel cosmosapi.Resource

type Model interface {
	// This method is called on entities after a successful Get() (whether from database or cache).
	// If the result of a Collection.StaleGet() is used, txn==nil; if Transaction.Get() is used,
	// txn is set.
	PostGet(txn *Transaction) error
	// This method is called on entities right before the write to database.
	// If Collection.RacingPut() is used, txn==nil; if we are inside a transaction
	// commit, txn is set.
	PrePut(txn *Transaction) error
}

// Client is an interface exposing the public API of the cosmosapi.Client struct
type Client interface {
	GetDocument(ctx context.Context, dbName, colName, id string, ops cosmosapi.GetDocumentOptions, out interface{}) error
	CreateDocument(ctx context.Context, dbName, colName string, doc interface{}, ops cosmosapi.CreateDocumentOptions) (*cosmosapi.Resource, cosmosapi.DocumentResponse, error)
	ReplaceDocument(ctx context.Context, dbName, colName, id string, doc interface{}, ops cosmosapi.ReplaceDocumentOptions) (*cosmosapi.Resource, cosmosapi.DocumentResponse, error)
	QueryDocuments(ctx context.Context, dbName, collName string, qry cosmosapi.Query, docs interface{}, ops cosmosapi.QueryDocumentsOptions) (cosmosapi.QueryDocumentsResponse, error)
	DeleteCollection(ctx context.Context, dbName, colName string) error
	DeleteDatabase(ctx context.Context, dbName string, ops *cosmosapi.RequestOptions) error
}
