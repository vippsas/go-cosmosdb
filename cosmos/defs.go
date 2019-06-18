package cosmos

import (
	"context"

	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

type BaseModel cosmosapi.Resource

// This method will return true if the document is new (document was not found on get, or get has not been attempted)
func (bm *BaseModel) IsNew() bool {
	return bm.Etag == ""
}

type Model interface {
	// This method is called on entities after a successful Get() (whether from database or cache).
	// If the result of a Collection.StaleGet() is used, txn==nil; if Transaction.Get() is used,
	// txn is set.
	PostGet(txn *Transaction) error
	// This method is called on entities right before the write to database.
	// If Collection.RacingPut() is used, txn==nil; if we are inside a transaction
	// commit, txn is set.
	PrePut(txn *Transaction) error
	// Exported by BaseModel
	IsNew() bool
}

// Client is an interface exposing the public API of the cosmosapi.Client struct
type Client interface {
	GetDocument(ctx context.Context, dbName, colName, id string, ops cosmosapi.GetDocumentOptions, out interface{}) (cosmosapi.DocumentResponse, error)
	CreateDocument(ctx context.Context, dbName, colName string, doc interface{}, ops cosmosapi.CreateDocumentOptions) (*cosmosapi.Resource, cosmosapi.DocumentResponse, error)
	ReplaceDocument(ctx context.Context, dbName, colName, id string, doc interface{}, ops cosmosapi.ReplaceDocumentOptions) (*cosmosapi.Resource, cosmosapi.DocumentResponse, error)
	QueryDocuments(ctx context.Context, dbName, collName string, qry cosmosapi.Query, docs interface{}, ops cosmosapi.QueryDocumentsOptions) (cosmosapi.QueryDocumentsResponse, error)
	ListDocuments(ctx context.Context, dbName, colName string, ops *cosmosapi.ListDocumentsOptions, docs interface{}) (cosmosapi.ListDocumentsResponse, error)
	GetCollection(ctx context.Context, dbName, colName string) (*cosmosapi.Collection, error)
	DeleteCollection(ctx context.Context, dbName, colName string) error
	DeleteDatabase(ctx context.Context, dbName string, ops *cosmosapi.RequestOptions) error
	ExecuteStoredProcedure(ctx context.Context, dbName, colName, sprocName string, ops cosmosapi.ExecuteStoredProcedureOptions, ret interface{}, args ...interface{}) error
	GetPartitionKeyRanges(ctx context.Context, dbName, colName string, options *cosmosapi.GetPartitionKeyRangesOptions) (cosmosapi.GetPartitionKeyRangesResponse, error)
	ListOffers(ctx context.Context, ops *cosmosapi.RequestOptions) (*cosmosapi.Offers, error)
	ReplaceOffer(ctx context.Context, offerOps cosmosapi.OfferReplaceOptions, ops *cosmosapi.RequestOptions) (*cosmosapi.Offer, error)
}
