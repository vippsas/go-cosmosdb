package cosmos

import (
	"context"
	"net/http"
	"sync"
)

type contextKey int

const (
	ckStateContainer contextKey = iota + 1
)

var (
	sessionSlotCountMu sync.Mutex
	sessionSlotCount   int
)

// WithSessions initializes a container for the session states on the context. This enables restoring the cosmos
// session from the context. Can be used recursively to reset the session states.
func WithSessions(ctx context.Context) context.Context {
	return context.WithValue(ctx, ckStateContainer, newStateContainer())
}

// SessionMiddleware is a convenience middleware for initializing the session state container on the request context.
// See: WithSessions()
func SessionsMiddleware(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(WithSessions(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

func initForContextSessions(coll *Collection) {
	if coll.sessionSlotIndex != 0 {
		return
	}
	sessionSlotCountMu.Lock()
	defer sessionSlotCountMu.Unlock()
	sessionSlotCount++
	coll.sessionSlotIndex = sessionSlotCount // Important that this is never 0 as the default value indicates a collection that hasn't been registered
}

func setStateFromContext(ctx context.Context, session *Session) {
	sc := getStateContainer(ctx)
	sc.setState(session)
}

func getStateContainer(ctx context.Context) *stateContainer {
	val := ctx.Value(ckStateContainer)
	if val == nil {
		panic("Sessions not initialized on context. Try calling cosmos.WithSessions(ctx)")
	}
	return val.(*stateContainer)
}

type stateContainer struct {
	mu     sync.Mutex
	states map[int]*sessionState
}

func (sc *stateContainer) setState(session *Session) {
	idx := session.Collection.sessionSlotIndex
	if idx == 0 {
		panic("Storing session state on context requires that Collection.Init() has been called on the collection")
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if state, ok := sc.states[idx]; ok {
		session.state = state
	} else {
		sc.states[idx] = session.state
	}
}

func newStateContainer() *stateContainer {
	// update of sessionSlotCount is atomic, so no need to lock here
	return &stateContainer{states: make(map[int]*sessionState, sessionSlotCount)}
}
