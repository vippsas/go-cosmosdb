package cosmosapi

import (
	"encoding/json"
)

func MarshalPartitionKeyHeader(partitionKeyValue interface{}) (string, error) {
	switch partitionKeyValue.(type) {
	// for now we disallow float, as using floats as keys is conceptually flawed (floats are not exact values)
	case nil, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	default:
		return "", ErrInvalidPartitionKeyType
	}
	res, err := json.Marshal([]interface{}{partitionKeyValue})
	if err != nil {
		return "", err
	}
	return string(res), nil
}
