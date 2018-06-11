package cosmosdb

func CreateDatabaseLink(dbName string) string {
	return "dbs/" + dbName
}

func CreateCollLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName
}

func CreateDocsLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/docs"
}

func CreateDocLink(dbName, coll, doc string) string {
	return "dbs/" + dbName + "/colls/" + coll + "/docs/" + doc
}
