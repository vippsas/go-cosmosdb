// The cosmostest package contains utilities for writing tests with cosmos, using a real database
// or the emulator as a backend, and with the option of multiple tests running side by side
// in multiple namespaces in a single collection to save costs.
//
//  Configuration
//
//  The standard configuration is to have a special file
//  "testconfig.yaml" in the currenty directory when running the
//  test. The config struct is expected inside a key "cosmostest", like this:
//
//  cosmostest:
//    Uri: "https://foo.documents.azure.com:443/"
//    MasterKey: "yourkeyhere=="
//    <... other fields from Config ...>
//
package cosmostest

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/vippsas/go-cosmosdb/cosmos"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Uri                     string `yaml:"Uri"`
	MasterKey               string `yaml:"MasterKey"`
	MultiTenant             bool   `yaml:"MultiTenant"`
	TlsInsecureSkipVerify   bool   `yaml:"TlsInsecureSkipVerify"`
	DbName                  string `yaml:"DbName"`
	CollectionIdPrefix      string `yaml:"CollectionIdPrefix"`
	AllowExistingCollection bool   `yaml:"AllowExistingCollection"`
}

var globalConfig Config

func check(err error, message string) {
	if err != nil {
		if message != "" {
			err = errors.New(message)
		}
		panic(err)
	}
}

func loadGlobalConfig() {
	var configDoc struct {
		CosmosTest Config `yaml:"cosmostest"`
	}

	configfile, err := os.Open("testconfig.yaml")
	check(err, "Problems opening testconfig.yaml")
	defer configfile.Close()
	d := yaml.NewDecoder(configfile)
	err = d.Decode(&configDoc)
	check(err, "")
	if configDoc.CosmosTest.Uri == "" {
		panic(errors.New("load from localconfig.yaml failed, expected info not in file"))
	}
	globalConfig = configDoc.CosmosTest
}

// Factory for constructing the underlying, proper cosmosapi.Client given configuration.
// This is typically called by / wrapped by the test collection providers.
func RawClient(cfg Config) *cosmosapi.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.TlsInsecureSkipVerify,
			},
		},
	}

	return cosmosapi.New(cfg.Uri, cosmosapi.Config{
		MasterKey:  cfg.MasterKey,
		MaxRetries: 3,
	}, httpClient)
}

// Setup will initialize a fresh collection using the DB / emulator pointed to in localconfig.yaml.
// Use New() to provide config in another manner.
func Setup(log cosmos.Logger, collectionId, partitionKey string) cosmos.Collection {
	if globalConfig.Uri == "" {
		loadGlobalConfig()
	}
	return SetupCollection(log, globalConfig, collectionId, partitionKey)
}

func Teardown(c cosmos.Collection) {
	if globalConfig.Uri == "" {
		panic(errors.New("Teardown called before Setup..."))
	}

	// Not implemented in driver yet..
}

func SetupCollection(log cosmos.Logger, cfg Config, collectionId, partitionKey string) cosmos.Collection {
	prefix := cfg.CollectionIdPrefix
	if prefix == "" {
		prefix = uuid.Must(uuid.NewV4()).String() + "-"
	}

	collectionId = prefix + collectionId

	client := RawClient(cfg)
	_, err := client.CreateCollection(context.Background(), cfg.DbName, cosmosapi.CollectionCreateOptions{
		Id: collectionId,
		PartitionKey: &cosmosapi.PartitionKey{
			Paths: []string{"/" + partitionKey},
			Kind:  "Hash",
		},
		OfferThroughput: 400,
	})
	if cfg.AllowExistingCollection && errors.Cause(err) == cosmosapi.ErrConflict {
		err = nil
	}
	check(err, "")

	return cosmos.Collection{
		Client:       client,
		DbName:       cfg.DbName,
		Name:         collectionId,
		PartitionKey: partitionKey,
		Log:          log,
	}

}
