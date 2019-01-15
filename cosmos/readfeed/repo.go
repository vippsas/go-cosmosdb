package readfeed

import (
	"fmt"
	"github.com/vippsas/go-cosmosdb/cosmos"
)

type testDocument struct {
	cosmos.BaseModel
	Model        string `json:"model" cosmosmodel:"Document/0"`
	PartitionKey string `json:"partitionkey"`
	Text         string `json:"text"`
}

func (d testDocument) String() string {
	return fmt.Sprintf("Id=%s PartitionKey=%s Text=%s", d.Id, d.PartitionKey, d.Text)
}

func (*testDocument) PostGet(txn *cosmos.Transaction) error {
	return nil
}

func (*testDocument) PrePut(txn *cosmos.Transaction) error {
	return nil
}

type testDocumentRepo struct {
	Collection cosmos.Collection
	session    cosmos.Session
	hasSession bool
}

func (r testDocumentRepo) Session() *cosmos.Session {
	if !r.hasSession {
		r.session = r.Collection.Session()
		r.hasSession = true
	}
	return &r.session
}

func (r testDocumentRepo) GetOrCreate(toCreate *testDocument) (ret *testDocument, created bool, err error) {
	ret = &testDocument{}
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

func (r testDocumentRepo) Update(partitionKey string, id string, update func(*testDocument) error) (document *testDocument, err error) {
	err = r.Session().Transaction(func(txn *cosmos.Transaction) error {
		p := &testDocument{}
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
