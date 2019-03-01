package cosmos

import (
	"context"
	"fmt"
	"github.com/vippsas/go-cosmosdb/logging"
	"reflect"

	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

const (
	fmtUnexpectedIdError                = "Unexpeced Id on fetched document: expected '%s', got '%s'"
	fmtUnexpectedPartitionKeyValueError = "Unexpected partition key vaule on fetched document: expected '%v', got: '%v'"
)

type Collection struct {
	Client       Client
	DbName       string
	Name         string
	PartitionKey string
	Context      context.Context
	Log          logging.StdLogger
}

func (c Collection) GetContext() context.Context {
	if c.Context == nil {
		return context.Background()
	} else {
		return c.Context
	}
}

func (c Collection) WithContext(ctx context.Context) Collection {
	c.Context = ctx // note that c is not a pointer
	return c
}

func (c Collection) log() logging.ExtendedLogger {
	return logging.Adapt(c.Log)
}

func (c Collection) get(ctx context.Context, partitionValue interface{}, id string, target Model, consistency cosmosapi.ConsistencyLevel, sessionToken string) (cosmosapi.DocumentResponse, error) {
	docResp, err := c.getExisting(ctx, partitionValue, id, target, consistency, sessionToken)
	res, partitionValueField := c.getEntityInfo(target)
	if err != nil && errors.Cause(err) == cosmosapi.ErrNotFound {
		err = nil
		// To be bullet-proof, make sure to zero out the target. It could e.g. be used for other purposes in a loop,
		// it is nice to be able to rely on zeroing out on not-found
		val := reflect.ValueOf(target).Elem()
		zero := reflect.Zero(val.Type())
		val.Set(zero)
		// Then write the ID information so that Put() will work after populating the entity
		partitionValueField.Set(reflect.ValueOf(partitionValue))
		res.Id = id
	}
	if err == nil {
		if res.Id != id {
			return docResp, errors.Errorf(fmtUnexpectedIdError, id, res.Id)
		}
		if partitionValueField.Interface() != partitionValue {
			return docResp, errors.Errorf(fmtUnexpectedPartitionKeyValueError, partitionValue, partitionValueField.Interface())
		}
	}
	return docResp, err
}

func (c Collection) getExisting(ctx context.Context, partitionValue interface{}, id string, target Model, consistency cosmosapi.ConsistencyLevel, sessionToken string) (cosmosapi.DocumentResponse, error) {
	opts := cosmosapi.GetDocumentOptions{
		PartitionKeyValue: partitionValue,
		ConsistencyLevel:  consistency,
		SessionToken:      sessionToken,
	}
	docResp, err := c.Client.GetDocument(ctx, c.DbName, c.Name, id, opts, target)
	if err != nil {
		return docResp, errors.Wrap(err, fmt.Sprintf("id='%s' partitionValue='%s'", id, partitionValue))
	}
	return docResp, nil
}

// StaleGet reads an element from the database. `target` should be a pointer to a struct
// that empeds BaseModel. If the document does not exist, the recipient
// struct is filled with the zero-value, including Etag which will become an empty String.
func (c Collection) StaleGet(partitionValue interface{}, id string, target Model) error {
	_, err := c.get(c.GetContext(), partitionValue, id, target, cosmosapi.ConsistencyLevelEventual, "")
	if err == nil {
		err = postGet(target.(Model), nil)
	}
	return err
}

// StaleGetExisting is similar to StaleGet, but returns an error if
// the document is not found instead of an empty document.  Test for
// this condition using errors.Cause(e) == cosmosapi.ErrNotFound
func (c Collection) StaleGetExisting(partitionValue interface{}, id string, target Model) error {
	_, err := c.getExisting(c.GetContext(), partitionValue, id, target, cosmosapi.ConsistencyLevelEventual, "")
	if err == nil {
		err = postGet(target.(Model), nil)
	}
	return err
}

// GetEntityInfo uses reflection to return information about the entity
// without each entity having to implement getters. One should pass a pointer
// to a struct that embeds "BaseModel" as well as a field having the partition field
// name; failure to do so will panic.
//
// Note: GetEntityInfo will also always assert that the Model property is set to the declared
// value
func (c Collection) GetEntityInfo(entityPtr Model) (res BaseModel, partitionValue interface{}) {
	resPtr, partitionValueField := c.getEntityInfo(entityPtr)
	return *resPtr, partitionValueField.Interface()
}

func (c Collection) getEntityInfo(entityPtr Model) (res *BaseModel, partitionValueField reflect.Value) {
	if c.PartitionKey == "" {
		panic(errors.Errorf("Please initialize PartitionKey in your Collection struct"))
	}
	defer func() {
		if e := recover(); e != nil {
			panic(errors.Errorf("Need to pass in a pointer to a struct with fields named 'BaseModel' and a tag 'json:\"%s\"', got: %s", c.PartitionKey, fmt.Sprintf("%v", entityPtr)))
		}
	}()

	v := reflect.ValueOf(entityPtr).Elem()
	structT := v.Type()
	res = v.FieldByName("BaseModel").Addr().Interface().(*BaseModel)
	n := structT.NumField()
	found := false
	if c.PartitionKey == "id" {
		partitionValueField = reflect.ValueOf(res).Elem().FieldByName("Id")
		found = true
	} else {
		for i := 0; i != n; i++ {
			field := structT.Field(i)
			if field.Tag.Get("json") == c.PartitionKey {
				partitionValueField = v.Field(i)
				found = true
				break
			}
		}
	}
	if !found {
		panic(errors.New(""))
	}
	return
}

func (c Collection) put(ctx context.Context, entityPtr Model, base BaseModel, partitionValue interface{}, consistent bool) (
	resource *cosmosapi.Resource, response cosmosapi.DocumentResponse, err error) {

	// if consistent = false, we always use the database upsert primitive (non-consistent put)
	// Otherwise, we demand non-existence if entity.Etag==nil, and replace with Etag if entity.Etag!=nil
	if !consistent || base.Etag == "" {
		opts := cosmosapi.CreateDocumentOptions{
			PartitionKeyValue: partitionValue,
			IsUpsert:          !consistent,
		}
		resource, response, err = c.Client.CreateDocument(ctx, c.DbName, c.Name, entityPtr, opts)
		if consistent && errors.Cause(err) == cosmosapi.ErrConflict {
			// For consistent creation with Etag="" we translate ErrConflict on creation to ErrPreconditionFailed
			err = errors.WithStack(cosmosapi.ErrPreconditionFailed)
		}
	} else {
		opts := cosmosapi.ReplaceDocumentOptions{
			PartitionKeyValue: partitionValue,
			IfMatch:           base.Etag,
		}
		resource, response, err = c.Client.ReplaceDocument(ctx, c.DbName, c.Name, base.Id, entityPtr, opts)
	}
	err = errors.WithStack(err)
	return
}

// RacingPut simply does a raw write of document passed in without any considerations about races
// or consistency. An "upsert" will be performed without any Etag checks. `entityPtr` should be a pointer to the struct
func (c Collection) RacingPut(entityPtr Model) error {
	base, partitionValue := c.GetEntityInfo(entityPtr)

	if err := prePut(entityPtr.(Model), nil); err != nil {
		return err
	}

	_, _, err := c.put(c.GetContext(), entityPtr, base, partitionValue, false)
	return err
}

func (c Collection) Query(query string, entities interface{}) error {
	_, err := c.Client.QueryDocuments(c.Context, c.DbName, c.Name, cosmosapi.Query{Query: query}, entities, cosmosapi.DefaultQueryDocumentOptions())
	return err
}

// Execute a StoredProcedure on the collection
func (c Collection) ExecuteSproc(sprocName string, partitionKeyValue interface{}, ret interface{}, args ...interface{}) error {
	opts := cosmosapi.ExecuteStoredProcedureOptions{PartitionKeyValue: partitionKeyValue}
	return c.Client.ExecuteStoredProcedure(
		c.GetContext(), c.DbName, c.Name, sprocName, opts, ret, args...)
}

// Retrieve <maxItems> documents that have changed within the partition key range since <etag>. Note that according to
// https://docs.microsoft.com/en-us/rest/api/cosmos-db/list-documents (as of Jan 14 16:30:27 UTC 2019) <maxItems>, which
// corresponds to the x-ms-max-item-count HTTP request header, is (quote):
//
// "An integer indicating the maximum number of items to be returned per page."
//
// However incremental feed reads seems to always return maximum one page, ie. the continuation token (x-ms-continuation
// HTTP response header) is always empty.
func (c Collection) ReadFeed(etag, partitionKeyRangeId string, maxItems int, documents interface{}) (cosmosapi.ListDocumentsResponse, error) {
	ops := cosmosapi.ListDocumentsOptions{
		MaxItemCount:        maxItems,
		AIM:                 "Incremental feed",
		PartitionKeyRangeId: partitionKeyRangeId,
		IfNoneMatch:         etag,
	}
	response, err := c.Client.ListDocuments(c.GetContext(), c.DbName, c.Name, &ops, documents)
	return response, err
}

func (c Collection) GetPartitionKeyRanges() ([]cosmosapi.PartitionKeyRange, error) {
	ops := cosmosapi.GetPartitionKeyRangesOptions{}
	response, err := c.Client.GetPartitionKeyRanges(c.GetContext(), c.DbName, c.Name, &ops)
	return response.PartitionKeyRanges, err
}
