package cosmosapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

func (c *Client) GetPartitionKeyRanges(
	ctx context.Context,
	databaseName, collectionName string,
	options *GetPartitionKeyRangesOptions,
) (response GetPartitionKeyRangesResponse, err error) {
	link := CreateCollLink(databaseName, collectionName) + "/pkranges"
	var responseBody getPartitionKeyRangesResponseBody
	if options != nil {
		if options.MaxItemCount == 0 && options.Continuation == "" {
			// Caller presumably used the old version of the library, which didn't
			// take continuations into account. If they haven't set MaxItemCount or
			// Continuation, we assume they want all items.
			return c.getAllPartitionKeyRanges(ctx, databaseName, collectionName, options)
		}
	}
	headers, err := options.AsHeaders()
	if err != nil {
		return response, err
	}
	httpResponse, err := c.get(ctx, link, &responseBody, headers)
	if err != nil {
		return response, err
	}
	response.PartitionKeyRanges = responseBody.PartitionKeyRanges
	response.Id = responseBody.Id
	response.Rid = responseBody.Rid
	err = response.parseHeaders(httpResponse)
	if err != nil {
		return response, errors.WithMessage(err, "Failed to get partition key ranges")
	}
	return response, err
}

type getPartitionKeyRangesResponseBody struct {
	Rid                string              `json:"_rid"`
	Id                 string              `json:"id"`
	PartitionKeyRanges []PartitionKeyRange `json:"PartitionKeyRanges"`
}

type PartitionKeyRange struct {
	Id           string   `json:"id"`
	MaxExclusive string   `json:"maxExclusive"`
	MinInclusive string   `json:"minInclusive"`
	Parents      []string `json:"parents"`
}

type GetPartitionKeyRangesOptions struct {
	MaxItemCount int
	Continuation string
}

func (ops GetPartitionKeyRangesOptions) AsHeaders() (map[string]string, error) {
	headers := map[string]string{}
	if ops.MaxItemCount != 0 {
		headers[HEADER_MAX_ITEM_COUNT] = strconv.Itoa(ops.MaxItemCount)
	}
	if ops.Continuation != "" {
		headers[HEADER_CONTINUATION] = ops.Continuation
	}
	return headers, nil
}

type GetPartitionKeyRangesResponse struct {
	Id                 string
	Rid                string
	PartitionKeyRanges []PartitionKeyRange
	RequestCharge      float64
	SessionToken       string
	Continuation       string
	Etag               string
}

func (r *GetPartitionKeyRangesResponse) parseHeaders(httpResponse *http.Response) error {
	r.SessionToken = httpResponse.Header.Get(HEADER_SESSION_TOKEN)
	r.Continuation = httpResponse.Header.Get(HEADER_CONTINUATION)
	r.Etag = httpResponse.Header.Get(HEADER_ETAG)
	if _, ok := httpResponse.Header[HEADER_REQUEST_CHARGE]; ok {
		requestCharge, err := strconv.ParseFloat(httpResponse.Header.Get(HEADER_REQUEST_CHARGE), 64)
		if err != nil {
			return errors.WithStack(err)
		}
		r.RequestCharge = requestCharge
	}
	return nil
}

func (c *Client) getAllPartitionKeyRanges(ctx context.Context, databaseName, collectionName string, options *GetPartitionKeyRangesOptions) (GetPartitionKeyRangesResponse, error) {
	options.MaxItemCount = -1
	options.Continuation = ""
	p := c.NewPartitionKeyRangesPaginator(databaseName, collectionName, options)
	var pkranges GetPartitionKeyRangesResponse
	for p.Next() {
		newPk, err := p.CurrentPage(ctx)
		if err != nil {
			return pkranges, err
		}
		newPk.PartitionKeyRanges = append(pkranges.PartitionKeyRanges, newPk.PartitionKeyRanges...)
		newPk.RequestCharge += pkranges.RequestCharge
		pkranges = newPk
	}
	return pkranges, nil
}

// NewPartitionKeyRangesPaginator returns a paginator for ListObjectsV2. Use the
// Next method to get the next page, and CurrentPage to get the current response
// page from the paginator. Next will return false if there are no more pages,
// or an error was encountered.
//
// Note: This operation can generate multiple requests to a service.
//
//   // Example iterating over pages.
//   p := client.NewPartitionKeyRangesPaginator(input)
//
//   for p.Next() {
//       err, page := p.CurrentPage(context.TODO())
//       if err != nil {
//         return err
//       }
//   }
//
func (c *Client) NewPartitionKeyRangesPaginator(databaseName, collectionName string, options *GetPartitionKeyRangesOptions) *PartitionKeyRangesPaginator {
	var opts GetPartitionKeyRangesOptions
	if options != nil {
		opts = *options
	}
	return &PartitionKeyRangesPaginator{
		databaseName:   databaseName,
		collectionName: collectionName,
		options:        opts,
		client:         c,
	}
}

// PartitionKeyRangesPaginator is a paginator over the "Get Partition key
// ranges" API endpoint. This paginator is not threadsafe.
type PartitionKeyRangesPaginator struct {
	shouldFetchPage bool
	hasPage         bool

	err         error
	currentPage GetPartitionKeyRangesResponse

	client         *Client
	databaseName   string
	collectionName string
	options        GetPartitionKeyRangesOptions
}

// CurrentPage returns the current page of partition key ranges. Panics if
// Next() has not yet been called.
func (p *PartitionKeyRangesPaginator) CurrentPage(ctx context.Context) (GetPartitionKeyRangesResponse, error) {
	if !p.shouldFetchPage && !p.hasPage {
		panic("PartitionKeyRangesPaginator: Must call Next before CurrentPage")
	}
	if p.shouldFetchPage { // includes retries if the previous call errored out
		p.currentPage, p.err = p.client.GetPartitionKeyRanges(ctx, p.databaseName, p.collectionName, &p.options)
		if p.err == nil {
			p.shouldFetchPage = false
			p.hasPage = true
			p.options.Continuation = p.currentPage.Continuation
		}
	}
	return p.currentPage, p.err
}

// Next returns true if there are more pages to be read, and false if the
// previous CurrentPage call returned an error, or if there are no more pages to
// be read.
func (p *PartitionKeyRangesPaginator) Next() bool {
	if p.err != nil {
		return false
	}
	if !p.hasPage {
		p.shouldFetchPage = true
		return true
	}
	// Check if we have a continuation token
	if p.options.Continuation != "" {
		p.shouldFetchPage = true
		return true
	}
	return false
}
