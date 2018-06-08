package main

import (
	"context"
	"fmt"

	"github.com/starsheriff/go-cosmosdb"
)

func main() {
	db := cosmosdb.Database{}

	coll, err := db.Collection(context.Background(), "myCollection")
	if err != nil {
		fmt.Println(err)
	}

	coll.Document()
	return
}
