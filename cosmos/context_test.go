package cosmos

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func patchSessionGetterCount() func() {
	sessionSlotCountMu.Lock()
	defer sessionSlotCountMu.Unlock()
	origVal := sessionSlotCount
	sessionSlotCount = 0
	return func() {
		sessionSlotCountMu.Lock()
		defer sessionSlotCountMu.Unlock()
		sessionSlotCount = origVal
	}
}

func TestSessionGetter(tt *testing.T) {
	for name, test := range map[string]func(t *testing.T){
		"WithSessions": func(t *testing.T) {
			ctx := context.Background()
			require.Panics(t, func() { getStateContainer(ctx) })
			ctx = WithSessions(ctx)
			require.NotNil(t, getStateContainer(ctx))
		},
		"Middleware": func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://test.test", nil)
			require.NoError(t, err)
			require.Panics(t, func() { getStateContainer(req.Context()) })
			SessionsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.NotNil(t, getStateContainer(req.Context()))
			}))
		},
		"Collection.Init": func(t *testing.T) {
			coll := Collection{}
			require.Equal(t, 0, coll.sessionSlotIndex)
			coll = coll.Init()
			require.Equal(t, 1, coll.sessionSlotIndex)
			coll = coll.Init()
			require.Equal(t, 1, coll.sessionSlotIndex)
		},
		"Collection.SessionContext": func(t *testing.T) {
			ctx := context.Background()
			coll := Collection{}.Init()
			require.Panics(t, func() { coll.SessionContext(ctx) })
			ctx = WithSessions(ctx)
			session := coll.SessionContext(ctx)
			session2 := coll.SessionContext(ctx)
			if session.state != session2.state {
				t.Errorf("Both sessions must point to the same state")
			}
			session3 := Collection{}.Init().SessionContext(ctx)
			if session.state == session3.state {
				t.Error("Sessions from different collections must not share state")
			}
		},
		"Collection.Session": func(t *testing.T) {
			ctx := context.Background()
			coll := Collection{}.Init()
			require.Panics(t, func() { coll.SessionContext(ctx) })
			ctx = WithSessions(ctx)
			session := coll.Session()
			session2 := coll.Session()
			if session.state == session2.state {
				t.Errorf("Sessions states must be different")
			}
		},
		"Reset state": func(t *testing.T) {
			ctx := context.Background()
			ctx = WithSessions(ctx)
			coll := Collection{}.Init()
			session := coll.SessionContext(ctx)
			session2 := coll.SessionContext(WithSessions(ctx))
			if session.state == session2.state {
				t.Errorf("Sessions must point to different states")
			}
		},
	} {
		tt.Run(name, func(t *testing.T) {
			defer patchSessionGetterCount()()
			test(t)
		})
	}
}
