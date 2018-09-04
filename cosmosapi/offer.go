package cosmosapi

import (
	"context"
)

type Offer struct {
	Resource
	OfferVersion    string                 `json:"offerVersion"`
	OfferType       OfferType              `json:"offerType"`
	Content         OfferThroughputContent `json:"content,omitempty"`
	OfferResourceId string                 `json:"offerResourceId"`
}

type OfferThroughput int32
type OfferType string

type OfferThroughputContent struct {
	Throughput OfferThroughput `json:"offerThroughput"`
}

type Offers struct {
	Rid    string  `json:"_rid,omitempty"`
	Count  int32   `json:"_count,omitempty"`
	Offers []Offer `json:"Offers"`
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/replace-an-offer
type OfferReplaceOptions struct {
	OfferVersion     string                 `json:"offerVersion"`
	OfferType        OfferType              `json:"offerType"`
	Content          OfferThroughputContent `json:"content,omitempty"`
	ResourceSelfLink string                 `json:"resource"`
	OfferResourceId  string                 `json:"offerResourceId"`
	Id               string                 `json:"id"`
	Rid              string                 `json:"_rid"`
}

func createOfferLink(offerId string) string {
	return "offers/" + offerId
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/get-an-offer
func (c *Client) GetOffer(ctx context.Context, offerId string, ops *RequestOptions) (*Offer, error) {
	offer := &Offer{}
	err := c.get(ctx, createOfferLink(offerId), offer, nil)

	if err != nil {
		return nil, err
	}
	return offer, nil
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/list-offers
func (c *Client) ListOffers(ctx context.Context, ops *RequestOptions) (*Offers, error) {

	url := createOfferLink("")

	offers := &Offers{}
	err := c.get(ctx, url, offers, nil)
	if err != nil {
		return nil, err
	}

	return offers, nil
}

// https://docs.microsoft.com/en-us/rest/api/cosmos-db/replace-an-offer
func (c *Client) ReplaceOffer(ctx context.Context, offerOps OfferReplaceOptions, ops *RequestOptions) (*Offer, error) {

	offer := &Offer{}
	link := createOfferLink(offerOps.Rid)

	_, err := c.replace(ctx, link, offerOps, offer, nil)
	if err != nil {
		return nil, err
	}

	return offer, nil

}
