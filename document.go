package cosmosdb

// Document
type Document struct {
	Resource
	Attachments string `json:"attachments,omitempty"`
}
