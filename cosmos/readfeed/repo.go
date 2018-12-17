package readfeed

import (
	"github.com/vippsas/go-cosmosdb/cosmos"
)

type DocumentRepo interface {
	GetOrCreate(toCreate *Document) (Document *Document, created bool, err error)
	Get(payerId string, id string) (*Document, error)
	Update(payerId string, id string, update func(Document *Document) error) (Document *Document, err error)
}

type DocumentCosmosRepo struct {
	Collection cosmos.Collection
	session    cosmos.Session
	hasSession bool
}

func (r DocumentCosmosRepo) Session() *cosmos.Session {
	if !r.hasSession {
		r.session = r.Collection.Session()
		r.hasSession = true
	}
	return &r.session
}

func (r DocumentCosmosRepo) GetOrCreate(toCreate *Document) (ret *Document, created bool, err error) {
	ret = &Document{}
	err = r.Session().Transaction(func(txn *cosmos.Transaction) error {
		var err error
		err = txn.Get(toCreate.PayerId, toCreate.Id, ret)
		if err != nil {
			return err
		}
		if !ret.IsNew() {
			return nil
		}
		created = true
		ret = toCreate
		txn.Put(ret)
		return nil
	})
	return
}

func (r DocumentCosmosRepo) Get(payerId string, id string) (*Document, error) {
	Document := &Document{}
	err := r.Session().Transaction(func(txn *cosmos.Transaction) error {
		err := txn.Get(payerId, id, Document)
		return err
	})
	return Document, err
}

func (r DocumentCosmosRepo) Update(payerId string, id string, update func(*Document) error) (Document *Document, err error) {
	err = r.Session().Transaction(func(txn *cosmos.Transaction) error {
		p := &Document{}
		if err := txn.Get(payerId, id, p); err != nil {
			return err
		}
		if err := update(p); err != nil {
			return err
		}
		txn.Put(p)
		Document = p
		return nil
	})
	return
}
