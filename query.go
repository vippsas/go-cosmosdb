package cosmosdb

type Query struct {
	Query  string       `json:"query"`
	Params []QueryParam `json:"parameters,omitempty"`
	Token  string       `json:"-"` // continuation token
}

type QueryParam struct {
	Name  string      `json:"name"` // should contain a @ character
	Value interface{} `json:"value"`
}

// TODO: add missing fields
type QueryDocumentsResponse struct {
	Documents    interface{}
	Count        int `json:"_count"`
	Continuation string
	RUs          int
}

// NewQuery create a query with given parameters.
//
// Example:
//	NewQuery(
// 		`SELECT * FROM root r WHERE (r.id = @id)`,
//		map[string]interface{}{"@id": "foo"},
//	)
//func NewQuery(qu string, params map[string]interface{}) *Query {
//q := &Query{Query: qu}
//q.Params = make([]QueryParam, 0, len(params))
//for name, val := range params {
//q.Params = append(q.Params, QueryParam{Name: name, Value: val})
//}
//return q
//}
