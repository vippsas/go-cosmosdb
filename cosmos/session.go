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
	entityCache map[string]string
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
			entityCache: make(map[string]string),
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
func (session Session) Drop(id string) {
	delete(session.state.entityCache, id)
}

// Convenience method for doing a simple Get within a session without explicitly starting a transaction
func (session Session) Get(partitionValue interface{}, id string, target interface{}) error {
	return session.Transaction(func(txn *Transaction) error {
		return txn.Get(partitionValue, id, target)
	})
}

func (session Session) cacheSet(id string, entity interface{}) error {
	serialized, err := json.Marshal(entity)
	if err != nil {
		return errors.WithStack(err)
	}
	session.state.entityCache[id] = string(serialized)
	return nil
}

func (session Session) cacheGet(id string, entityPtr interface{}) (found bool, err error) {
	serialized, ok := session.state.entityCache[id]
	if !ok {
		return false, nil
	} else {
		return true, json.Unmarshal([]byte(serialized), entityPtr)
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
