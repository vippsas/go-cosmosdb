package cosmosdb

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

type IndexingPolicy struct {
	IndexingMode IndexingMode   `json:"indexingMode,omitempty"`
	Automatic    bool           `json:"automatic"`
	Included     []IncludedPath `json:"includedPaths,omitempty"`
	Excluded     []ExcludedPath `json:"excludedPaths,omitempty"`
}

type IndexingMode string

const (
	Consistent = IndexingMode("Consistent")
	Lazy       = IndexingMode("Lazy")
)

type PartitionKey struct {
	Paths []string `json:"paths"`
	Kind  string   `json:"kind"`
}
