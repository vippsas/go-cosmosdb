package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithAllFields(t *testing.T) {
	cd := getCollectionDefinitions("test_data/all_fields.json")
	assert.NotNil(t, cd)
	assert.Len(t, cd, 1)
}

func TestWithoutPartitionKey(t *testing.T) {
	cd := getCollectionDefinitions("test_data/not_partition_key.json")
	assert.NotNil(t, cd)
	assert.Len(t, cd, 1)
}

func TestWithoutIndexingPolicy(t *testing.T) {
	cd := getCollectionDefinitions("test_data/not_indexing_policy.json")
	assert.NotNil(t, cd)
	assert.Len(t, cd, 1)
}
