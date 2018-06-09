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
	id    string
	value string
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

	// Create a document
	link := cosmosdb.CreateCollLink(cfg.DbName, "invoices")
	doc := ExampleDoc{id: "aaa", value: "666"}
	err = client.CreateDocument(context.Background(), link, doc, nil)
	if err != nil {
		err = errors.WithStack(err)
		fmt.Println(err)
	}

	return
}
