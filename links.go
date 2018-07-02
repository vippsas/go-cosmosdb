package cosmosdb

import (
	"strings"
)

func CreateCollLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName
}

func CreateDocsLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/docs"
}

func createDocLink(dbName, coll, doc string) string {
	return "dbs/" + dbName + "/colls/" + coll + "/docs/" + doc
}

// resourceTypeFromLink is used to extract the resource type link to use in the
// payload of the authorization header.
func resourceTypeFromLink(verb, link string) (rLink, rType string) {
	if strings.HasPrefix(link, "/") == false {
		link = "/" + link
	}
	if strings.HasSuffix(link, "/") == false {
		link = link + "/"
	}

	parts := strings.Split(link, "/")
	l := len(parts)

	// TODO: extend with missing http verbs.
	switch verb {
	case "GET", "DELETE":
		if l%2 == 0 {
			rLink = strings.Join(parts[1:l-1], "/")
			rType = parts[l-3]
		} else {
			rLink = strings.Join(parts[1:l-1], "/")
			rType = parts[l-2]
		}
	case "POST":
		if l%2 == 0 {
			rLink = strings.Join(parts[1:l-2], "/")
			rType = parts[l-3]
		} else {
			rLink = strings.Join(parts[1:l-2], "/")
			rType = parts[l-2]
		}

	default:
		if l%2 == 0 {
			rLink = strings.Join(parts[0:l-2], "/")
			rType = parts[l-3]
		} else {
			//rLink = strings.Join(parts[0:l-2], "/")
			rLink = link
			rType = parts[l-2]
		}
	}

	return
}
