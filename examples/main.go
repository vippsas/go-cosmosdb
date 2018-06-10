package main

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/starsheriff/go-cosmosdb"
)

type config struct {
	DbUrl  string
	DbKey  string
	DbName string
}

func fromEnv() config {
	cfg := config{}
	if err := envconfig.Process("", &cfg); err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	return cfg
}

type ExampleDoc struct {
	Id                    string `json:"id"`
	Value                 string
	RecipientPartitionKey string
}

type ExampleGetDoc struct {
	cosmosdb.Document
	Id                    string `json:"id"`
	RecipientPartitionKey string
}

func main() {
	fmt.Printf("Starting with examples...\n")

	cfg := fromEnv()
	cosmosCfg := cosmosdb.Config{
		MasterKey: cfg.DbKey,
	}

	client := cosmosdb.New(cfg.DbUrl, cosmosCfg, nil)

	// Get a database
	dbLink := cosmosdb.CreateDatabaseLink(cfg.DbName)
	db, err := client.GetDatabase(context.Background(), dbLink, nil)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	fmt.Println(db)

	// Create a document without partition key
	link := cosmosdb.CreateDocsLink(cfg.DbName, "batchstatuses")
	doc := ExampleDoc{Id: "aaa", Value: "666", RecipientPartitionKey: "asdf"}
	err = client.CreateDocument(context.Background(), link, doc, nil)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	// Create a document with partition key
	link = cosmosdb.CreateDocsLink(cfg.DbName, "invoices")
	doc = ExampleDoc{Id: "aaa", Value: "666", RecipientPartitionKey: "asdf"}
	ro := cosmosdb.RequestOptions{
		cosmosdb.ReqOpPartitionKey: "[\"asdf\"]",
	}
	err = client.CreateDocument(context.Background(), link, doc, &ro)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	// Get a document with partitionkey
	link = cosmosdb.CreateDocLink(cfg.DbName, "invoices", "aaa")
	doc = ExampleDoc{Id: "aaa"}
	ro = cosmosdb.RequestOptions{
		cosmosdb.ReqOpPartitionKey: "[\"asdf\"]",
	}
	err = client.GetDocument(context.Background(), link, &ro, &doc)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	fmt.Printf("Received document: %+v\n", doc)

	return
}
