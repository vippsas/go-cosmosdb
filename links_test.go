package cosmosdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceTypeFromLink(t *testing.T) {
	cases := []struct {
		verb  string
		in    string
		rLink string
		rType string
	}{
		{"GET", "/dbs", "dbs", "dbs"},
		{"GET", "dbs", "dbs", "dbs"},
		{"GET", "/dbs/myDb", "dbs/myDb", "dbs"},
		{"GET", "/dbs/myDb/", "dbs/myDb", "dbs"},
		{"GET", "/dbs/myDb/colls", "dbs/myDb", "colls"},
		{"GET", "/dbs/myDb/colls/", "dbs/myDb", "colls"},
		{"GET", "/dbs/myDb/colls/someCol", "dbs/myDb/colls/someCol", "colls"},
		{"GET", "/dbs/myDb/colls/someCol/", "dbs/myDb/colls/someCol", "colls"},
		{"POST", "/dbs/myDb/colls/myColl/docs/", "dbs/myDb/colls/myColl", "docs"},
		{"POST", "/dbs/vippsinvoice/colls", "dbs/vippsinvoice", "colls"},
		{"DELETE", "/dbs/db/colls/col/docs/doc", "dbs/db/colls/col/docs/doc", "docs"},
		{"PUT", "/dbs/db/colls/col/docs/doc", "dbs/db/colls/col/docs/doc", "docs"},
	}
	for _, c := range cases {
		t.Run("case: "+c.verb+": "+c.in, func(t *testing.T) {
			rLink, rType := resourceTypeFromLink(c.verb, c.in)
			assert.Equal(t, c.rType, rType, "Type")
			assert.Equal(t, c.rLink, rLink, "Link")
		})
	}
}
