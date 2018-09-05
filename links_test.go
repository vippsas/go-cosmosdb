package cosmosdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceTypeFromLink(t *testing.T) {
	cases := []struct {
		in    string
		rLink string
		rType string
	}{
		{"/dbs", "", "dbs"},
		{"dbs", "", "dbs"},
		{"/dbs/myDb", "dbs/myDb", "dbs"},
		{"/dbs/myDb/", "dbs/myDb", "dbs"},
		{"/dbs/myDb/colls", "dbs/myDb", "colls"},
		{"/dbs/myDb/colls/", "dbs/myDb", "colls"},
		{"/dbs/myDb/colls/someCol", "dbs/myDb/colls/someCol", "colls"},
		{"/dbs/myDb/colls/someCol/", "dbs/myDb/colls/someCol", "colls"},
		{"/dbs/myDb/colls/myColl/docs/", "dbs/myDb/colls/myColl", "docs"},
		{"/dbs/db/colls/col/docs/doc", "dbs/db/colls/col/docs/doc", "docs"},
		{"/dbs/db/colls/col/docs/doc", "dbs/db/colls/col/docs/doc", "docs"},
		{"/offers/myOffer", "myoffer", "offers"},
		{"/offers/CASING", "casing", "offers"},
	}
	for _, c := range cases {
		t.Run("case: "+c.in, func(t *testing.T) {
			rLink, rType := resourceTypeFromLink(c.in)
			assert.Equal(t, c.rType, rType, "Type")
			assert.Equal(t, c.rLink, rLink, "Link")
		})
	}
}
