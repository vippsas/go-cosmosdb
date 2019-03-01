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
	"crypto/x509"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/cosmos"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
	"github.com/vippsas/go-cosmosdb/logging"
	"net/http"
)

type Config struct {
	Uri                     string `yaml:"Uri"`
	MasterKey               string `yaml:"MasterKey"`
	MultiTenant             bool   `yaml:"MultiTenant"`
	TlsCertificate          string `yaml:"TlsCertificate"`
	TlsServerName           string `yaml:"TlsServerName"`
	TlsInsecureSkipVerify   bool   `yaml:"TlsInsecureSkipVerify"`
	DbName                  string `yaml:"DbName"`
	CollectionIdPrefix      string `yaml:"CollectionIdPrefix"`
	AllowExistingCollection bool   `yaml:"AllowExistingCollection"`
}

func check(err error, message string) {
	if err != nil {
		if message != "" {
			err = errors.New(message)
		}
		panic(err)
	}
}

// Factory for constructing the underlying, proper cosmosapi.Client given configuration.
// This is typically called by / wrapped by the test collection providers.
func RawClient(cfg Config) *cosmosapi.Client {
	if cfg.Uri == "" {
		panic("Missing requred parameter 'Uri'")
	}
	var caRoots *x509.CertPool
	if cfg.TlsCertificate != "" {
		caRoots = x509.NewCertPool()
		if !caRoots.AppendCertsFromPEM([]byte(cfg.TlsCertificate)) {
			panic("Failed to parse TLS certificate")
		}

	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caRoots,
				ServerName:         cfg.TlsServerName,
				InsecureSkipVerify: cfg.TlsInsecureSkipVerify,
			},
		},
	}

	return cosmosapi.New(cfg.Uri, cosmosapi.Config{
		MasterKey:  cfg.MasterKey,
		MaxRetries: 3,
	}, httpClient, nil)
}

func SetupUniqueCollectionWithExistingDatabaseAndDefaultThroughput(log logging.StdLogger, cfg Config, id, partitionKey string) cosmos.Collection {
	id = uuid.Must(uuid.NewV4()).String() + "-" + id
	log.Printf("Creating Cosmos collection %s/%s\n", cfg.DbName, id)
	client := RawClient(cfg)
	_, err := client.CreateCollection(context.Background(), cfg.DbName, cosmosapi.CollectionCreateOptions{
		Id: id,
		PartitionKey: &cosmosapi.PartitionKey{
			Paths: []string{"/" + partitionKey},
			Kind:  "Hash",
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create Cosmos collection %s in database %s\n: %+v", id, cfg.DbName, err))
	}
	return cosmos.Collection{
		Client:       client,
		DbName:       cfg.DbName,
		Name:         id,
		PartitionKey: partitionKey,
		Log:          log,
		Context:      context.Background(),
	}
}

func SetupCollection(log logging.StdLogger, cfg Config, collectionId, partitionKey string) cosmos.Collection {
	if cfg.CollectionIdPrefix == "" {
		cfg.CollectionIdPrefix = uuid.Must(uuid.NewV4()).String() + "-"
	}
	if cfg.DbName == "" {
		cfg.DbName = "default"
	}
	collectionId = cfg.CollectionIdPrefix + collectionId
	client := RawClient(cfg)
	if _, err := client.CreateDatabase(context.TODO(), cfg.DbName, nil); err != nil {
		if errors.Cause(err) != cosmosapi.ErrConflict {
			check(err, "Failed to create database")
		}
		// Database already existed, which is OK
	}
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

func TeardownCollection(collection cosmos.Collection) {
	collection.Client.DeleteCollection(collection.Context, collection.DbName, collection.Name)
}
