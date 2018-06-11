package cosmosdb

import (
	"context"
)

// Document
type Document struct {
	Resource
	Attachments string `json:"attachments,omitempty"`
}

func (c *Client) CreateDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions) error {

	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	err := c.create(ctx, link, doc, nil, headers)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetDocument(ctx context.Context, link string,
	ops *RequestOptions, out interface{}) error {

	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	err := c.get(ctx, link, out, headers)
	if err != nil {
		return err
	}

	return nil
}
