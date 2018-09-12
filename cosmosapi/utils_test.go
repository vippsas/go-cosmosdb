package cosmosapi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMarshalPartitionKeyHeader(t *testing.T) {
	checkMarshal := func(in, expect interface{}) {
		v, err := MarshalPartitionKeyHeader(in)
		if _, ok := expect.(error); ok {
			require.Equal(t, expect, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, expect, v)
		}
	}

	checkMarshal(nil, `[null]`)
	checkMarshal("foo", `["foo"]`)
	checkMarshal(1, `[1]`)
	checkMarshal(int32(1), `[1]`)
	checkMarshal(17179869184, `[17179869184]`) // in > 2^32

	checkMarshal(1234.0, ErrInvalidPartitionKeyType)
	checkMarshal(struct{}{}, ErrInvalidPartitionKeyType)
}
