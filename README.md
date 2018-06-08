# go-cosmosdb
go sdk for Azure Cosmos DB

* no `_self` links
* full support for partitioned collections
* simple interface
* supports all operations with user defined ids
* naming conventions follow the RestAPI [https://docs.microsoft.com/en-us/rest/api/cosmos-db/](https://docs.microsoft.com/en-us/rest/api/cosmos-db/)

* it more closely follows the api of the official SDKs
    * [https://docs.microsoft.com/python/api/pydocumentdb?view=azure-python](python)
    * [https://docs.microsoft.com/javascript/api/documentdb/?view=azure-node-latest](node.js)

# Usage

* instantiate a `config` struct. Set the keys, url and some other parameters.
* call the constructor `New(cfg config)`

* `cosmosdb` follows the hierachy of Cosmos DB. This means that you can operate
  on the resource the current type represents. The database struct can work with
  resources that belong to a cosmos database, the `Collection` type can work with
  resources that belong to a collection.
* `doc interface{}` may seem weird in some contexts, e.g. `DeleteDocument`, why
  not use a signature like `DeleteDocument(ctx context.Context, id string)`. The
  reason is that there are several ways to address the document. Either by self
  link, with or without `_etag` or by the `id`. All on collections with or without
  a partition key.
    * use `_self` if possible
    * if `_etag` is present, use it
    * otherwise use id
    * if neither exists -> error


# Examples

## Create Document

```
type Document struct {
    id string
}

newDoc, err := coll.CreateDocument(context.Background(), doc)
```


#FAQ

