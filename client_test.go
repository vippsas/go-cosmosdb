package cosmosdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// from MS documentation
const TestKey = "dsZQi3KtZmCv1ljt3VNWNm7sQUF1y5rJfC6kv5JiwvW0EndXdDku/dkKBp8/ufDToSxLzR4y+O/0H/t4bQtVNw=="

type TestDoc struct {
	id string
}

// TestMakeAuthHeader test the example from the RestAPI documentation found
// here https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources
func TestMakeAuthHeader(t *testing.T) {
	key := "dsZQi3KtZmCv1ljt3VNWNm7sQUF1y5rJfC6kv5JiwvW0EndXdDku/dkKBp8/ufDToSxLzR4y+O/0H/t4bQtVNw=="

	links := []string{"/dbs/ToDoList", "dbs/ToDoList"}
	for _, l := range links {
		t.Run("case: "+l, func(t *testing.T) {

			sign, err := makeSignedPayload("GET", l, "Thu, 27 Apr 2017 00:51:12 GMT", key)
			require.Nil(t, err)

			result := makeAuthHeader(sign)
			expected := "type%3Dmaster%26ver%3D1.0%26sig%3Dc09PEVJrgp2uQRkr934kFbTqhByc7TVr3OHyqlu%2Bc%2Bc%3D"

			assert.Equal(t, expected, result)
		})
	}
}

func TestResourceTypeFromLink(t *testing.T) {
	cases := []struct {
		verb  string
		in    string
		rLink string
		rType string
	}{
		{"GET", "/dbs", "dbs", "dbs"},
		{"GET", "dbs", "dbs", "dbs"},
		{"GET", "/dbs/myDb", "dbs/myDb", "dbs"},
		{"GET", "/dbs/myDb/", "dbs/myDb", "dbs"},
		{"GET", "/dbs/myDb/colls", "dbs/myDb/colls", "colls"},
		{"GET", "/dbs/myDb/colls/", "dbs/myDb/colls", "colls"},
		{"GET", "/dbs/myDb/colls/someCol", "dbs/myDb/colls/someCol", "colls"},
		{"GET", "/dbs/myDb/colls/someCol/", "dbs/myDb/colls/someCol", "colls"},
		{"POST", "/dbs/myDb/colls/myColl/docs/", "dbs/myDb/colls/myColl", "docs"},
	}
	for _, c := range cases {
		t.Run("case: "+c.verb+": "+c.in, func(t *testing.T) {
			rLink, rType := resourceTypeFromLink(c.verb, c.in)
			assert.Equal(t, c.rType, rType)
			assert.Equal(t, c.rLink, rLink)
		})
	}
}

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

	link := CreateDatabaseLink("ToDoList")

	_, err := c.GetDatabase(context.Background(), link, nil)
	assert.NotNil(t, err)
}
