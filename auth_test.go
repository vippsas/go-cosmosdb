package cosmosdb

import (
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

			sign, err := signedPayload("GET", l, "Thu, 27 Apr 2017 00:51:12 GMT", key)
			require.Nil(t, err)

			result := authHeader(sign)
			expected := "type%3Dmaster%26ver%3D1.0%26sig%3Dc09PEVJrgp2uQRkr934kFbTqhByc7TVr3OHyqlu%2Bc%2Bc%3D"

			assert.Equal(t, expected, result)
		})
	}
}
