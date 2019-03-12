package cosmosapi

import (
	"strings"
)

func CreateTriggerLink(dbName, collName, triggerName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/triggers/" + triggerName
}

func CreateCollLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName
}

func createDocsLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/docs"
}

func createDocLink(dbName, collName, doc string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/docs/" + doc
}

func createSprocsLink(dbName, collName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/sprocs"
}

func createSprocLink(dbName, collName, sprocName string) string {
	return "dbs/" + dbName + "/colls/" + collName + "/sprocs/" + sprocName
}

// resourceTypeFromLink is used to extract the resource type link to use in the
// payload of the authorization header.
func resourceTypeFromLink(link string) (rLink, rType string) {
	if link == "" {
		return "", ""
	}

	// Ensure link has leading '/'
	if strings.HasPrefix(link, "/") == false {
		link = "/" + link
	}

	// Ensure link ends with '/'
	if strings.HasSuffix(link, "/") == false {
		link = link + "/"
	}

	parts := strings.Split(link, "/")
	l := len(parts)

	// Offer is inconsistent from the rest of the API
	// For details see "Headers" block on https://docs.microsoft.com/en-us/rest/api/cosmos-db/get-an-offer
	if parts[1] == "offers" {
		rType = parts[1]
		rLink = strings.ToLower(parts[2])
		return
	}

	if l%2 == 0 {
		rType = parts[l-3]
		rLink = strings.Join(parts[1:l-1], "/")
	} else {
		// E.g. /dbs/myDb/colls/myColl/docs/
		// In this scenario the link is incomplete.
		// I.e. it does not not point to a specific resource

		rType = parts[l-2]
		rLink = strings.Join(parts[1:l-2], "/")
	}

	return
}
