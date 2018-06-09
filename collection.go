package cosmosdb

// Collection
type Collection struct {
	Resource
	IndexingPolicy *IndexingPolicy `json:"indexingPolicy,omitempty"`
	Docs           string          `json:"_docs,omitempty"`
	Udf            string          `json:"_udfs,omitempty"`
	Sporcs         string          `json:"_sporcs,omitempty"`
	Triggers       string          `json:"_triggers,omitempty"`
	Conflicts      string          `json:"_conflicts,omitempty"`
	PartitionKey   *PartitionKey   `json:"partitionKey,omitempty"`
}
