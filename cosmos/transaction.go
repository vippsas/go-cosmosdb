package cosmos

import (
	"reflect"
	"time"

	"github.com/pkg/errors"
	cosmosapi "github.com/vippsas/go-cosmosdb/cosmosapi"
)

// Transaction is simply a wrapper around Session which unlocks some of
// the methods that should only be called inside an idempotent closure
type Transaction struct {
	fetchedId uniqueKey // the id that was fetched in the single allowed Get()
	toPut     Model     // the entity that was queued for put in the single allowed Put()
	session   Session
}

var rollbackError = errors.New("__rollback__")

var ContentionError = errors.New("Contention error; optimistic concurrency control did not succeed after all the retries")
var NotImplementedError = errors.New("Not implemented")
var PutWithoutGetError = errors.New("Attempting to put an entity that has not been get first")

func Rollback() error {
	return rollbackError
}

// Transaction <todo rest of docs>. Note: On commit, the Etag is updated on all relevant
// entities (but normally these should never be used outside)
func (session Session) Transaction(closure func(*Transaction) error) error {
	if session.ConflictRetries == 0 {
		return errors.Errorf("Number of retries set to 0")
	}
	for i := 0; i != session.ConflictRetries; i++ {
		txn := Transaction{session: session}

		closureErr := closure(&txn)
		if closureErr == nil && txn.toPut != nil {
			putErr := txn.commit()
			if errors.Cause(putErr) == cosmosapi.ErrPreconditionFailed {
				// contention, loop around
				time.Sleep(100 * time.Millisecond) // TODO: randomization; use scaled put walltime
				continue
			}
			return putErr
		} else {
			// Implement Rollback() -- do not commit but do not return error either
			if errors.Cause(closureErr) == rollbackError {
				closureErr = nil
			}
			return closureErr
		}
	}
	return errors.WithStack(ContentionError)
}

func (txn *Transaction) commit() error {
	// Sanity check -- help the poor developer out by not allowing put without get
	base, partitionValue := txn.session.Collection.GetEntityInfo(txn.toPut)
	uniqueKey, err := newUniqueKey(partitionValue, base.Id)
	if err != nil {
		return err
	}
	if uniqueKey != txn.fetchedId {
		return errors.WithStack(PutWithoutGetError)
	}

	if err := prePut(txn.toPut.(Model), txn); err != nil {
		return err
	}

	// Execute the put
	newBase, response, err := txn.session.Collection.put(txn.session.Context, txn.toPut, base, partitionValue, true)

	// no matter what happened, if we got a session token we want to update to it
	if response.SessionToken != "" {
		txn.session.state.sessionToken = response.SessionToken
	}

	if err == nil {
		// Successful PUT, so
		// a) update Etag on the entity (this intentionally affects callers copy if caller still has one, which should
		//    not usually be the case..)
		// below reflect is doing: txn.toPut.BaseModel = newBase
		reflect.ValueOf(txn.toPut).Elem().FieldByName("BaseModel").Set(reflect.ValueOf(BaseModel(*newBase)))

		// b) add updated entity to the session's entity cache.
		// If there is an error here it would be in JSON serialized; in that case panic, it should
		// never happen since we just serialized in the same way above...
		if jsonSerializationErr := txn.session.cacheSet(partitionValue, base.Id, txn.toPut); jsonSerializationErr != nil {
			panic(errors.Errorf("This should never happen: The entity successfully serialized to JSON the first time, but not the second ... %s", jsonSerializationErr))
		}

	} else if errors.Cause(err) == cosmosapi.ErrPreconditionFailed {
		// We know that this object is staled, make sure to remove it from cache
		txn.session.Drop(partitionValue, base.Id)
	}

	return err

}

func (txn *Transaction) Get(partitionValue interface{}, id string, target Model) (err error) {
	uniqueKey, err := newUniqueKey(partitionValue, id)
	if err != nil {
		return err
	}
	if txn.fetchedId != "" && txn.fetchedId != uniqueKey {
		return errors.Wrap(NotImplementedError, "Fetching more than one entity in transaction not supported yet")
	}

	var found bool
	found, err = txn.session.cacheGet(partitionValue, id, target)
	if err != nil {
		// Trouble in JSON deserialization from cache; a bug in deserialization hooks or similar... return it
		return err
	}
	if found {
		// do nothing, cacheGet already unserialized to target
	} else {
		// post-get hook will be done by Collection.get()
		err = txn.session.Collection.get(
			txn.session.Context,
			partitionValue,
			id,
			target,
			cosmosapi.ConsistencyLevelSession,
			txn.session.Token())
		if err == nil {
			txn.session.cacheSet(partitionValue, id, target)
		}
	}

	if err == nil {
		txn.fetchedId = uniqueKey
		err = postGet(target, txn)
	}
	return
}

func (txn *Transaction) Put(entityPtr Model) {
	txn.toPut = entityPtr
}
