package cosmosdb

import (
	"context"
)

type Trigger struct {
	Resource
	Id        string           `json:"id"`
	Body      string           `json:"body"`
	Operation TriggerOperation `json:"triggerOperation"`
	Type      TriggerType      `json:"triggerType"`
}

type TriggerType string
type TriggerOperation string

type CollectionTriggers struct {
	Rid      string    `json:"_rid,omitempty"`
	Count    int32     `json:"_count,omitempty"`
	Triggers []Trigger `json:"Triggers"`
}

//const (
//	TriggerTypePost = TriggerType("Post")
//	TriggerTypePre  = TriggerType("Pre")
//
//	TriggerOpAll     = TriggerType("All")
//	TriggerOpCreate  = TriggerType("Create")
//	TriggerOpReplace = TriggerType("Replace")
//	TriggerOpDelete  = TriggerType("Delete")
//)

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/create-a-trigger
type TriggerCreateOptions struct {
	Id        string           `json:"id"`
	Body      string           `json:"body"`
	Operation TriggerOperation `json:"triggerOperation"`
	Type      TriggerType      `json:"triggerType"`
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/replace-a-trigger
type TriggerReplaceOptions struct {
	Id        string           `json:"id"`
	Body      string           `json:"body"`
	Operation TriggerOperation `json:"triggerOperation"`
	Type      TriggerType      `json:"triggerType"`
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/create-a-trigger
func (c *Client) CreateTrigger(ctx context.Context, dbName string, colName string,
	trigOps TriggerCreateOptions) (*Trigger, error) {

	trigger := &Trigger{}
	link := CreateTriggerLink(dbName, colName, "")

	err := c.create(ctx, link, trigOps, trigger, nil)

	if err != nil {
		return nil, err
	}
	return trigger, nil
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/list-triggers
func (c *Client) ListTriggers(ctx context.Context, dbName string, colName string) (*CollectionTriggers, error) {

	url := CreateCollLink(dbName, colName) + "/triggers"

	colTrigs := &CollectionTriggers{}
	err := c.get(ctx, url, colTrigs, nil)
	if err != nil {
		return nil, err
	}

	return colTrigs, nil
}

func (c *Client) DeleteTrigger(ctx context.Context, dbName, colName string) error {
	return ErrorNotImplemented
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/replace-a-trigger
func (c *Client) ReplaceTrigger(ctx context.Context, dbName, colName string,
	trigOps TriggerReplaceOptions) (*Trigger, error) {

	trigger := &Trigger{}
	link := CreateTriggerLink(dbName, colName, trigOps.Id)

	err := c.replace(ctx, link, trigOps, trigger, nil)
	if err != nil {
		return nil, err
	}

	return trigger, nil

}
