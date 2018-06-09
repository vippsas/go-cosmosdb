package cosmosdb

// Database
type Database struct {
	Resource
	Colls string `json:"_colls,omitempty"`
	Users string `json:"_users,omitempty"`
}
