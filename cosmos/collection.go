package cosmos

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

type Logger interface {
	Print(args ...interface{})
	Printf(fmt string, args ...interface{})
}

type Collection struct {
	Client       Client
	DbName       string
	Name         string
	PartitionKey string
	Context      context.Context
	Log          Logger
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

func (c Collection) get(ctx context.Context, partitionValue interface{}, id string, target interface{}, consistency cosmosapi.ConsistencyLevel, sessionToken string) error {
	err := c.getExisting(ctx, partitionValue, id, target, consistency, sessionToken)
	if err != nil && errors.Cause(err) == cosmosapi.ErrNotFound {
		err = nil
		// To be bullet-proof, make sure to zero out the target. It could e.g. be used for other purposes in a loop,
		// it is nice to be able to rely on zeroing out on not-found
		val := reflect.ValueOf(target).Elem()
		zero := reflect.Zero(val.Type())
		val.Set(zero)
		// Then write the ID information so that Put() will work after populating the entity
		structT := val.Type()
		n := structT.NumField()
		found := false
		for i := 0; i != n; i++ {
			if structT.Field(i).Tag.Get("json") == c.PartitionKey {
				val.Field(i).Set(reflect.ValueOf(partitionValue))
				found = true
				break
			}
		}
		if !found {
			panic(errors.Errorf("Did not find any struct fields with tag json:\"%s\"", c.PartitionKey))
		}
		val.FieldByName("BaseModel").Addr().Interface().(*BaseModel).Id = id
	}

	return err
}

func (c Collection) getExisting(ctx context.Context, partitionValue interface{}, id string, target interface{}, consistency cosmosapi.ConsistencyLevel, sessionToken string) error {
	opts := cosmosapi.GetDocumentOptions{
		PartitionKeyValue: partitionValue,
		ConsistencyLevel:  consistency,
		SessionToken:      sessionToken,
	}
	err := c.Client.GetDocument(ctx, c.DbName, c.Name, id, opts, target)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("id='%s' partitionValue='%s'", id, partitionValue))
	}
	return nil
}

// StaleGet reads an element from the database. `target` should be a pointer to a struct
// that empeds BaseModel. If the document does not exist, the recipient
// struct is filled with the zero-value, including Etag which will become an empty String.
func (c Collection) StaleGet(partitionValue interface{}, id string, target interface{}) error {
	err := c.get(c.GetContext(), partitionValue, id, target, cosmosapi.ConsistencyLevelEventual, "")
	if err == nil {
		err = postGet(target.(Model), nil)
	}
	return err
}

// StaleGetExisting is similar to StaleGet, but returns an error if
// the document is not found instead of an empty document.  Test for
// this condition using errors.Cause(e) == cosmosapi.ErrNotFound
func (c Collection) StaleGetExisting(partitionValue interface{}, id string, target interface{}) error {
	err := c.getExisting(c.GetContext(), partitionValue, id, target, cosmosapi.ConsistencyLevelEventual, "")
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
func (c Collection) GetEntityInfo(entityPtr interface{}) (res BaseModel, partitionValue interface{}) {
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
	res = v.FieldByName("BaseModel").Interface().(BaseModel)
	n := structT.NumField()
	found := false
	for i := 0; i != n; i++ {
		field := structT.Field(i)
		if field.Tag.Get("json") == c.PartitionKey {
			partitionValue = v.Field(i).Interface()
			found = true
			break
		}
	}
	if !found {
		panic(errors.New(""))
	}
	return
}

func (c Collection) put(ctx context.Context, entity interface{}, base BaseModel, partitionValue interface{}, consistent bool) (
	resource *cosmosapi.Resource, response cosmosapi.DocumentResponse, err error) {

	// if consistent = false, we always use the database upsert primitive (non-consistent put)
	// Otherwise, we demand non-existence if entity.Etag==nil, and replace with Etag if entity.Etag!=nil
	if !consistent || base.Etag == "" {
		opts := cosmosapi.CreateDocumentOptions{
			PartitionKeyValue: partitionValue,
			IsUpsert:          !consistent,
		}
		resource, response, err = c.Client.CreateDocument(ctx, c.DbName, c.Name, entity, opts)
		if consistent && errors.Cause(err) == cosmosapi.ErrConflict {
			// For consistent creation with Etag="" we translate ErrConflict on creation to ErrPreconditionFailed
			err = errors.WithStack(cosmosapi.ErrPreconditionFailed)
		}
	} else {
		opts := cosmosapi.ReplaceDocumentOptions{
			PartitionKeyValue: partitionValue,
			IfMatch:           base.Etag,
		}
		resource, response, err = c.Client.ReplaceDocument(ctx, c.DbName, c.Name, base.Id, entity, opts)
	}
	err = errors.WithStack(err)
	return
}

// PutInconsistent simply does a raw write of document passed in without any considerations about races
// or consistency. An "upsert" will be performed without any Etag checks. `doc` should be a pointer to the struct
func (c Collection) RacingPut(entity interface{}) error {
	base, partitionValue := c.GetEntityInfo(entity)

	if err := prePut(entity.(Model), nil); err != nil {
		return err
	}

	_, _, err := c.put(c.GetContext(), entity, base, partitionValue, false)
	return err
}

func (c Collection) Query(query string, entities interface{}) error {
	_, err := c.Client.QueryDocuments(c.Context, c.DbName, c.Name, cosmosapi.Query{Query: query}, entities, cosmosapi.DefaultQueryDocumentOptions())
	return err
}
