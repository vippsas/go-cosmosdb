package readfeed

import (
	"fmt"
	"github.com/vippsas/go-cosmosdb/cosmos"
)

type Document struct {
	cosmos.BaseModel
	Model        string `json:"model" cosmosmodel:"Document/0"`
	PartitionKey string `json:"partitionkey"`
	Text         string `json:"text"`
}

func (d Document) String() string {
	return fmt.Sprintf("Id=%s PartitionKey=%s Text=%s", d.Id, d.PartitionKey, d.Text)
}

func (*Document) PostGet(txn *cosmos.Transaction) error {
	return nil
}

func (*Document) PrePut(txn *cosmos.Transaction) error {
	return nil
}

type DocumentRepo interface {
	GetOrCreate(toCreate *Document) (Document *Document, created bool, err error)
	Get(partitionKey string, id string) (*Document, error)
	Update(partitionKey string, id string, update func(Document *Document) error) (Document *Document, err error)
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
		err = txn.Get(toCreate.PartitionKey, toCreate.Id, ret)
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

func (r DocumentCosmosRepo) Get(partitionKey string, id string) (*Document, error) {
	Document := &Document{}
	err := r.Session().Transaction(func(txn *cosmos.Transaction) error {
		err := txn.Get(partitionKey, id, Document)
		return err
	})
	return Document, err
}

func (r DocumentCosmosRepo) Update(partitionKey string, id string, update func(*Document) error) (document *Document, err error) {
	err = r.Session().Transaction(func(txn *cosmos.Transaction) error {
		p := &Document{}
		if err := txn.Get(partitionKey, id, p); err != nil {
			return err
		}
		if err := update(p); err != nil {
			return err
		}
		txn.Put(p)
		document = p
		return nil
	})
	return
}
