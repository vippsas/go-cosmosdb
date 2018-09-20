package cosmos

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

//
// Our test model
//

type MyModel struct {
	BaseModel
	Model       string `json:"model" cosmosmodel:"MyModel/1"`
	UserId      string `json:"userId"`      // partition key
	X           int    `json:"x"`           // data
	SetByPrePut string `json:"setByPrePut"` // set by pre-put hook

	XPlusOne int `json:-` // computed field set by post-get hook
}

func (e *MyModel) PrePut(txn *Transaction) error {
	e.SetByPrePut = "set by pre-put, checked in mock"
	return nil
}

func (e *MyModel) PostGet(txn *Transaction) error {
	e.XPlusOne = e.X + 1
	return nil
}

//
// Our mock of Cosmos DB -- this mocks the interface provided by cosmosapi package
//

type mockCosmos struct {
	ReturnX         int
	ReturnEtag      string
	ReturnSession   string
	ReturnError     error
	GotId           string
	GotPartitionKey interface{}
	GotMethod       string
	GotUpsert       bool
	GotX            int
	GotSession      string
}

func (mock *mockCosmos) QueryDocuments(ctx context.Context, dbName, collName string, qry cosmosapi.Query, docs interface{}, ops cosmosapi.QueryDocumentsOptions) (cosmosapi.QueryDocumentsResponse, error) {
	panic("implement me")
}

func (mock *mockCosmos) DeleteCollection(ctx context.Context, dbName, colName string) error {
	panic("implement me")
}

func (mock *mockCosmos) DeleteDatabase(ctx context.Context, dbName string, ops *cosmosapi.RequestOptions) error {
	panic("implement me")
}

func (mock *mockCosmos) ExecuteStoredProcedure(ctx context.Context, dbName, colName, sprocName string, ops cosmosapi.ExecuteStoredProcedureOptions, ret interface{}, args ...interface{}) error {
	panic("implement me")
}

func (mock *mockCosmos) reset() {
	*mock = mockCosmos{}
}

func (mock *mockCosmos) GetDocument(ctx context.Context,
	dbName, colName, id string, ops cosmosapi.GetDocumentOptions, out interface{}) error {

	mock.GotId = id
	mock.GotMethod = "get"
	mock.GotSession = ops.SessionToken

	t := out.(*MyModel)
	t.X = mock.ReturnX
	t.BaseModel.Etag = mock.ReturnEtag
	t.BaseModel.Id = id
	return mock.ReturnError
}

func (mock *mockCosmos) CreateDocument(ctx context.Context,
	dbName, colName string, doc interface{}, ops cosmosapi.CreateDocumentOptions) (*cosmosapi.Resource, cosmosapi.DocumentResponse, error) {
	t := doc.(*MyModel)
	mock.GotMethod = "create"
	mock.GotPartitionKey = ops.PartitionKeyValue
	mock.GotId = t.Id
	mock.GotX = t.X
	mock.GotUpsert = ops.IsUpsert

	if t.SetByPrePut != "set by pre-put, checked in mock" {
		panic(errors.New("assertion failed"))
	}

	newBase := cosmosapi.Resource{
		Id:   t.Id,
		Etag: mock.ReturnEtag,
	}
	return &newBase, cosmosapi.DocumentResponse{SessionToken: mock.ReturnSession}, mock.ReturnError
}

func (mock *mockCosmos) ReplaceDocument(ctx context.Context,
	dbName, colName, id string, doc interface{}, ops cosmosapi.ReplaceDocumentOptions) (*cosmosapi.Resource, cosmosapi.DocumentResponse, error) {
	t := doc.(*MyModel)
	mock.GotMethod = "replace"
	mock.GotPartitionKey = ops.PartitionKeyValue
	mock.GotId = t.Id
	mock.GotX = t.X

	if t.SetByPrePut != "set by pre-put, checked in mock" {
		panic(errors.New("assertion failed"))
	}

	newBase := cosmosapi.Resource{
		Id:   t.Id,
		Etag: mock.ReturnEtag,
	}
	return &newBase, cosmosapi.DocumentResponse{SessionToken: mock.ReturnSession}, mock.ReturnError
}

type mockCosmosNotFound struct {
	mockCosmos
}

func (mockCosmosNotFound) QueryDocuments(ctx context.Context, dbName, collName string, qry cosmosapi.Query, docs interface{}, ops cosmosapi.QueryDocumentsOptions) (cosmosapi.QueryDocumentsResponse, error) {
	panic("implement me")
}

func (mockCosmosNotFound) ExecuteStoredProcedure(ctx context.Context, dbName, colName, sprocName string, ops cosmosapi.ExecuteStoredProcedureOptions, ret interface{}, args ...interface{}) error {
	panic("implement me")
}

func (mockCosmosNotFound) GetDocument(ctx context.Context, dbName, colName, id string, ops cosmosapi.GetDocumentOptions, out interface{}) error {
	return cosmosapi.ErrNotFound
}

func (mock *mockCosmosNotFound) DeleteCollection(ctx context.Context, dbName, colName string) error {
	panic("implement me")
}

func (mock *mockCosmosNotFound) DeleteDatabase(ctx context.Context, dbName string, ops *cosmosapi.RequestOptions) error {
	panic("implement me")
}

//
// Tests
//

func TestGetEntityInfo(t *testing.T) {
	c := Collection{
		Client:       &mockCosmosNotFound{},
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}
	e := MyModel{BaseModel: BaseModel{Id: "id1"}, UserId: "Alice"}
	res, pkey := c.GetEntityInfo(&e)
	require.Equal(t, "id1", res.Id)
	require.Equal(t, "Alice", pkey)
}

func TestCheckModel(t *testing.T) {
	e := MyModel{Model: "MyModel/1"}
	require.Equal(t, "MyModel/1", CheckModel(&e))
}

func TestCollectionStaleGet(t *testing.T) {
	c := Collection{
		Client:       &mockCosmosNotFound{},
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	var target MyModel
	target.X = 3
	target.Etag = "some-e-tag"

	err := c.StaleGetExisting("foo", "foo", &target)
	// StaleGetExisting: target not modified, returns not found error
	require.Equal(t, 3, target.X)
	require.Equal(t, cosmosapi.ErrNotFound, errors.Cause(err))

	// StaleGet: target zeroed, returns nil
	err = c.StaleGet("foo", "foo", &target)
	require.NoError(t, err)
	require.Equal(t, 0, target.X)
	require.Equal(t, "", target.Etag)
}

func TestCollectionRacingPut(t *testing.T) {
	mock := mockCosmos{}
	c := Collection{
		Client:       &mock,
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	entity := MyModel{
		BaseModel: BaseModel{
			Id: "id1",
		},
		X:      1,
		UserId: "alice",
	}

	require.NoError(t, c.RacingPut(&entity))
	require.Equal(t, mockCosmos{
		GotId:           "id1",
		GotPartitionKey: "alice",
		GotMethod:       "create",
		GotUpsert:       true,
		GotX:            1,
	}, mock)

	entity.Etag = "has an etag"

	// Should not affect RacingPut at all, it just does upserts..
	require.NoError(t, c.RacingPut(&entity))
	require.Equal(t, mockCosmos{
		GotId:           "id1",
		GotPartitionKey: "alice",
		GotMethod:       "create",
		GotUpsert:       true,
		GotX:            1,
	}, mock)

}

func TestTransactionCacheHappyDay(t *testing.T) {
	mock := mockCosmos{}
	c := Collection{
		Client:       &mock,
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	session := c.Session()

	var entity MyModel // in production code this should be declared inside closure, but want more control in this test

	require.NoError(t, session.Transaction(func(txn *Transaction) error {
		entity.X = -20
		mock.ReturnError = cosmosapi.ErrNotFound
		require.NoError(t, txn.Get("partitionvalue", "idvalue", &entity))
		require.Equal(t, "get", mock.GotMethod)
		// due to ErrNotFound, the Get() should zero-initialize to wipe the -20
		require.Equal(t, 0, entity.X)
		require.Equal(t, 1, entity.XPlusOne) // PostGetHook called

		require.Equal(t, "idvalue", mock.GotId)
		entity.X = 42
		mock.reset()
		txn.Put(&entity)
		// *not* put yet, so mock not called yet, and not in cache
		require.Equal(t, "", mock.GotMethod)
		require.Equal(t, 0, len(session.state.entityCache))
		mock.ReturnEtag = "etag-1" // Etag returned by mock on commit; this needs to find its way into cache
		mock.ReturnSession = "session-token-1"
		return nil
	}))
	// now after exiting closure the X=42-entity was put
	// also there was a create, not a replace, because entity.Etag was empty
	require.Equal(t, "create", mock.GotMethod)
	// cache should be populated
	jsonInCache, ok := session.state.entityCache["idvalue"]
	require.True(t, ok)
	require.Contains(t, jsonInCache, "\"etag-1\"")

	// Session token should be set from the create call
	require.Equal(t, "session-token-1", session.Token())

	// entity outside of scope should have updated etag (this should typically not be used by code,
	// but by writing this test it is in the contract as an edge case)
	require.Equal(t, "etag-1", entity.Etag)
	// Modify entity here just to make sure it doesn't reflect what is served by cache.
	entity.X = -10

	require.NoError(t, session.Transaction(func(txn *Transaction) error {
		mock.reset()
		require.NoError(t, txn.Get("partitionvalue", "idvalue", &entity))
		// Get() above hit cache, so mock was not called
		require.Equal(t, "", mock.GotMethod)
		require.Equal(t, 42, entity.X) // i.e., not the -10 value from above
		entity.X = 43
		txn.Put(&entity)
		mock.ReturnEtag = "etag-2"
		mock.ReturnSession = "session-token-2"
		return nil
	}))
	require.Equal(t, "replace", mock.GotMethod) // this time mock returned an etag on Get(), so we got a replace
	jsonInCache, ok = session.state.entityCache["idvalue"]
	require.True(t, ok)
	require.Contains(t, jsonInCache, "\"etag-2\"")

	// Session token should be set from the create call
	require.Equal(t, "session-token-2", session.Token())
}

func TestTransactionCollisionAndSessionTracking(t *testing.T) {
	mock := mockCosmos{}
	c := Collection{
		Client:       &mock,
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	session := c.Session()

	attempt := 0

	require.NoError(t, session.WithRetries(3).WithContext(context.Background()).Transaction(func(txn *Transaction) error {
		var entity MyModel

		mock.reset()

		require.NoError(t, txn.Get("partitionvalue", "idvalue", &entity))
		require.Equal(t, "get", mock.GotMethod)

		if attempt == 0 {
			require.Equal(t, "", mock.GotSession)
			mock.ReturnSession = "after-0"
			mock.ReturnError = cosmosapi.ErrPreconditionFailed
		} else if attempt == 1 {
			require.Equal(t, "after-0", mock.GotSession)
			mock.ReturnSession = "after-1"
			mock.ReturnError = cosmosapi.ErrPreconditionFailed
		} else if attempt == 2 {
			require.Equal(t, "after-1", mock.GotSession)
			mock.ReturnSession = "after-2"
			mock.ReturnError = nil
		}

		attempt++

		txn.Put(&entity)
		return nil
	}))

	require.Equal(t, 3, attempt)
	require.Equal(t, "after-2", session.Token())
}

func TestTransactionGetExisting(t *testing.T) {
	mock := mockCosmos{}
	c := Collection{
		Client:       &mock,
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	session := c.Session()

	require.NoError(t, session.WithRetries(3).WithContext(context.Background()).Transaction(func(txn *Transaction) error {
		var entity MyModel

		mock.ReturnEtag = "etag-1"
		mock.ReturnError = nil
		mock.ReturnX = 42
		require.NoError(t, txn.Get("partitionvalue", "idvalue", &entity))
		require.False(t, entity.IsNew())
		require.Equal(t, "get", mock.GotMethod)
		require.Equal(t, 42, entity.X)
		require.Equal(t, 43, entity.XPlusOne) // PostGetHook called
		return nil
	}))
}

func TestTransactionNonExisting(t *testing.T) {
	mock := mockCosmos{}
	c := Collection{
		Client:       &mock,
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	session := c.Session()

	require.NoError(t, session.Transaction(func(txn *Transaction) error {
		var entity MyModel
		require.NoError(t, txn.Get("partitionValue", "idvalue", &entity))
		require.True(t, entity.IsNew())
		return nil
	}))
	return
}

func TestTransactionRollback(t *testing.T) {
	mock := mockCosmos{}
	c := Collection{
		Client:       &mock,
		DbName:       "mydb",
		Name:         "mycollection",
		PartitionKey: "userId"}

	session := c.Session()

	require.NoError(t, session.Transaction(func(txn *Transaction) error {
		var entity MyModel

		require.NoError(t, txn.Get("partitionvalue", "idvalue", &entity))

		mock.reset()
		txn.Put(&entity)
		return Rollback()
	}))

	// no api call done due to rollback
	require.Equal(t, "", mock.GotMethod)

}
