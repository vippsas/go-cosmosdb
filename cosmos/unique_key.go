package cosmos

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// In Cosmos DB document IDs are only unique within a partition key value. For cases where we need a globally unique
// identifier, such as caching, `uniqueKey` can be used.
// Documents also have the _rid property which is also globally unique, but not always practical to use as it requires
// fetching an existing document.
type uniqueKey string

func newUniqueKey(partitionKeyValue interface{}, id string) (uniqueKey, error) {
	// Use JSON for the cache key to match how Cosmos represents values
	d, err := json.Marshal([]interface{}{partitionKeyValue, id})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return uniqueKey(d), nil
}
