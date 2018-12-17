// +build !offline

package readfeed

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vippsas/go-cosmosdb/cosmos"
	"github.com/vippsas/go-cosmosdb/cosmostest"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var fixture cosmos.Collection

func TestMain(m *testing.M) {
	config := LoadCosmosConfiguration()
	fixture = cosmos.Collection{
		Client:       cosmostest.RawClient(config),
		DbName:       "feedtest",
		Name:         "feedtest-1",
		PartitionKey: "partitionkey",
		Log:          log.New(os.Stdout, "", 0),
	}
	retCode := m.Run()
	os.Exit(retCode)
}

func Test_WhenDocumentIsInsertedThenChangeAppearsOnFeed(t *testing.T) {
	err := fixture.RacingPut(aDocument())
	assert.NoError(t, err)
}

func aDocument() *Document {
	toCreate := &Document{}
	ret := &Document{}
	err := fixture.Session().Transaction(func(txn *cosmos.Transaction) error {
		var err error
		err = txn.Get(toCreate.PartitionKey, toCreate.Id, ret)
		if err != nil {
			return err
		}
		if !ret.IsNew() {
			return nil
		}
		ret = toCreate
		txn.Put(ret)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return ret
}

type Document struct {
	cosmos.BaseModel
	Model        string `json:"model" cosmosmodel:"Document/0"`
	PartitionKey string `json:"partitionkey"`
}

func (*Document) PostGet(txn *cosmos.Transaction) error {
	return nil
}

func (*Document) PrePut(txn *cosmos.Transaction) error {
	return nil
}

func LoadCosmosConfiguration() cosmostest.Config {
	var configDoc struct {
		CosmosTest cosmostest.Config `yaml:"CosmosTest"`
	}
	if configfile, err := OpenConfigurationFile(); err != nil {
		panic(fmt.Sprintf("Failed to read test configuration: %v", err))
	} else if err = yaml.NewDecoder(configfile).Decode(&configDoc); err != nil {
		panic(fmt.Sprintf("Failed to parse test configuration: %v", err))
	} else {
		return configDoc.CosmosTest
	}
}

func OpenConfigurationFile() (*os.File, error) {
	return doOpenConfigurationFile(".")
}

func doOpenConfigurationFile(path string) (*os.File, error) {
	if path, err := filepath.Abs(filepath.Join(path, "testconfig.yaml")); err != nil {
		return nil, err // Fail
	} else if file, err := os.Open(path); err == nil {
		return file, nil // Eureka!
	} else if filepath.Dir(path) == filepath.Dir(filepath.Dir(path)) {
		return nil, err // Fail -- searched up to root directory without finding file
	} else {
		return doOpenConfigurationFile(filepath.Dir(filepath.Dir(path))) // Check parent directory
	}
}
