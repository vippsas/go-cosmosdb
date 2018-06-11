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
	doc interface{}, ops *RequestOptions) (*Resource, error) {

	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	resource := &Resource{}

	err := c.create(ctx, link, doc, resource, headers)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (c *Client) UpsertDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions) error {
	return ErrorNotImplemented
}

// ListDocument reads either all documents or the incremental feed, aka.
// change feed.
// TODO: probably have to return continuation token for the feed
func (c *Client) ListDocument(ctx context.Context, link string,
	ops *RequestOptions, out interface{}) error {
	return ErrorNotImplemented
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

// ReplaceDocument replaces a whole document.
func (c *Client) ReplaceDocument(ctx context.Context, link string,
	doc interface{}, ops *RequestOptions, out interface{}) error {
	return ErrorNotImplemented
}

func (c *Client) DeleteDocument(ctx context.Context, link string, ops *RequestOptions) error {
	// add optional headers
	headers := map[string]string{}

	if ops != nil {
		for k, v := range *ops {
			headers[string(k)] = v
		}
	}

	err := c.delete(ctx, link, headers)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) QueryDocuments(ctx context.Context, link string, qry Query, ops *RequestOptions) error {
	return ErrorNotImplemented
}
