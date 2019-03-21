package cosmos

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
)

const DefaultConflictRetries = 3

type sessionState struct {
	sessionToken string

	// The entity cache is a map of string -> interface to json serialization.struct (not
	// pointer-to-struct). All the structs are dedidcated copies owned
	// by the cache and addresses are never handed out.
	entityCache map[uniqueKey][]byte
}

type Session struct {
	Context         context.Context
	ConflictRetries int
	Collection      Collection
	state           *sessionState
}

func (c Collection) Session() Session {
	return Session{
		state: &sessionState{
			entityCache: make(map[uniqueKey][]byte),
		},
		Context:         c.GetContext(), // at least context.Background() at this point ...
		Collection:      c,
		ConflictRetries: DefaultConflictRetries,
	}
}

func (c Collection) ResumeSession(token string) Session {
	session := c.Session()
	session.state.sessionToken = token
	return session
}

func (session Session) Token() string {
	return session.state.sessionToken
}

func (session Session) WithContext(ctx context.Context) Session {
	session.Context = ctx // note: non-pointer receiver
	return session
}

func (session Session) WithRetries(n int) Session {
	session.ConflictRetries = n // note: non-pointer receiver
	return session
}

// Drop removes an entity from the session cache, so that the next fetch will always go
// out externally to fetch it.
func (session Session) Drop(partitionValue interface{}, id string) {
	key, err := newUniqueKey(partitionValue, id)
	if err != nil {
		// This shouldn't happen. If we're unable to create the cache key, we wouldn't be able to populate the cache
		// for the partition/id combination in the first place
		panic(err)
	}
	delete(session.state.entityCache, key)
}

// Convenience method for doing a simple Get within a session without explicitly starting a transaction
func (session Session) Get(partitionValue interface{}, id string, target Model) error {
	return session.Transaction(func(txn *Transaction) error {
		return txn.Get(partitionValue, id, target)
	})
}

func (session Session) cacheSetEmpty(partitionValue interface{}, id string) error {
	key, err := newUniqueKey(partitionValue, id)
	if err != nil {
		return err
	}
	session.state.entityCache[key] = nil
	return nil
}

func (session Session) cacheSet(partitionValue interface{}, id string, entity Model) error {
	key, err := newUniqueKey(partitionValue, id)
	if err != nil {
		return err
	}
	var serialized []byte = nil
	if !entity.IsNew() {
		serialized, err = json.Marshal(entity)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	session.state.entityCache[key] = serialized
	return nil
}

func (session Session) cacheGet(partitionKey interface{}, id string, entityPtr Model) (found bool, err error) {
	key, err := newUniqueKey(partitionKey, id)
	if err != nil {
		return false, err
	}
	serialized, ok := session.state.entityCache[key]
	if !ok {
		return false, nil
	} else if serialized != nil {
		return true, json.Unmarshal(serialized, entityPtr)
	} else {
		session.Collection.initializeEmptyDoc(partitionKey, id, entityPtr)
		return true, nil
	}
}

/*
Future optimization: Another cache strategy is to use reflect to copy data as done below.
However we then also need a pass to zero any attributes without JSON in them, or similar...


--

func (session Session) cacheSet(id string, entity interface{}) {
	// entity should be a pointer to a model. We want to cache it *by value*
	entityVal := reflect.ValueOf(entity).Elem()
	ptrToCopy := reflect.New(entityVal.Type())
	ptrToCopy.Elem().Set(entityVal)

	session.state.entityCache[id] = ptrToCopy.Elem().Interface()
}

func (session Session) cacheGet(id string) interface{} {
	result, _ := session.state.entityCache[id]
	return result
}
*/
