package cosmosdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {
	t.Skip("currently broken")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

		assert.Equal(t, "type%3dmaster%26ver%3d1.0%26sig%3dc09PEVJrgp2uQRkr934kFbTqhByc7TVr3OHyqlu%2bc%2bc%3d", r.Header.Get(HEADER_AUTH))
	}))
	defer ts.Close()

	cfg := Config{
		MasterKey: TestKey,
	}
	c := New(ts.URL, cfg, nil)

	_, err := c.GetDatabase(context.Background(), "ToDoList", nil)
	assert.NotNil(t, err)
}
