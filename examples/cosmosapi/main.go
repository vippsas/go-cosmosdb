package main

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
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
	cosmosapi.Document
	Id                    string `json:"id"`
	RecipientPartitionKey string
}

func main() {
	fmt.Printf("Starting with examples...\n")

	cfg := fromEnv()
	cosmosCfg := cosmosapi.Config{
		MasterKey: cfg.DbKey,
	}

	client := cosmosapi.New(cfg.DbUrl, cosmosCfg, nil)

	// Get a database
	db, err := client.GetDatabase(context.Background(), cfg.DbName, nil)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	fmt.Println(db)

	// Create a document without partition key
	doc := ExampleDoc{Id: "aaa", Value: "666"}
	ops := cosmosapi.CreateDocumentOptions{}
	resource, _, err := client.CreateDocument(context.Background(), cfg.DbName, "batchstatuses", doc, ops)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}
	fmt.Println(resource)

	// Create a document with partition key
	fmt.Printf("\n CreateDocument with partition key.\n")
	doc = ExampleDoc{Id: "aaa", Value: "666", RecipientPartitionKey: "asdf"}
	ops = cosmosapi.CreateDocumentOptions{
		PartitionKeyValue: "asdf",
		IsUpsert:          true,
	}
	resource, _, err = client.CreateDocument(context.Background(), cfg.DbName, "invoices", doc, ops)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", resource)

	// Create a document with partition key
	fmt.Printf("\n CreateDocument with partition key.\n")
	resource, _, err = client.CreateDocument(context.Background(), cfg.DbName, "invoices", doc, ops)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", resource)

	// Get a document with partitionkey
	fmt.Printf("\nGet document with partition key.\n")
	doc = ExampleDoc{Id: "aaa"}
	ro := cosmosapi.GetDocumentOptions{
		PartitionKeyValue: "asdf",
	}
	err = client.GetDocument(context.Background(), cfg.DbName, "invoices", "aaa", ro, &doc)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	fmt.Printf("Received document: %+v\n", doc)

	// Replace a document with partitionkey
	fmt.Printf("\nReplace document with partition key.\n")
	doc = ExampleDoc{Id: "aaa", Value: "new value", RecipientPartitionKey: "asdf"}
	replaceOps := cosmosapi.ReplaceDocumentOptions{
		PartitionKeyValue: "asdf",
	}
	response, _, err := client.ReplaceDocument(context.Background(), cfg.DbName, "invoices", "aaa", &doc, replaceOps)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}
	fmt.Printf("Replaced document: %+v\n", response)

	// Replace a document with partitionkey
	fmt.Printf("\nReplace document with partition key.\n")
	doc = ExampleDoc{Id: "aaa", Value: "yet another new value", RecipientPartitionKey: "asdf"}
	replaceOps.IfMatch = response.Etag

	response, _, err = client.ReplaceDocument(context.Background(), cfg.DbName, "invoices", "aaa", &doc, replaceOps)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}
	fmt.Printf("Replaced document: %+v\n", response)

	// Get a document with partitionkey
	fmt.Printf("\nGet document with partition key.\n")
	doc = ExampleDoc{Id: "aaa"}
	ro = cosmosapi.GetDocumentOptions{
		PartitionKeyValue: "asdf",
	}
	err = client.GetDocument(context.Background(), cfg.DbName, "invoices", "aaa", ro, &doc)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	fmt.Printf("Received document: %+v\n", doc)

	// Query Documents
	fmt.Println("Query Documents")
	qops := cosmosapi.DefaultQueryDocumentOptions()
	qops.PartitionKeyValue = "asdf"

	qry := cosmosapi.Query{
		Query: "SELECT * FROM c WHERE c.id = @id",
		Params: []cosmosapi.QueryParam{
			{
				Name:  "@id",
				Value: "aaa",
			},
		},
	}

	var docs []ExampleDoc
	fmt.Printf("docs: %+v\n", docs)
	res, err := client.QueryDocuments(context.Background(), cfg.DbName, "invoices", qry, &docs, qops)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}
	//fmt.Printf("type of Documents: kind: %s", reflect.TypeOf

	fmt.Printf("Query results: %+v\n", res)
	fmt.Printf("Query results: %+v\n", res.Documents)
	fmt.Printf("Docs after: %+v\n", docs)

	// Delete a document with partition key
	fmt.Printf("\nDelete document with partition key.\n")
	do := cosmosapi.DeleteDocumentOptions{
		PartitionKeyValue: "asdf",
	}
	err = client.DeleteDocument(context.Background(), cfg.DbName, "invoices", "aaa", do)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	return
}
