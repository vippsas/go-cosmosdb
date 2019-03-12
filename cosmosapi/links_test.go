package cosmosapi

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

func TestCreateLinksUrlSafe(t *testing.T) {
	assert.Equal(t, "dbs/mydb/colls/mycoll/triggers/%0D", createTriggerLink("mydb", "mycoll", "\r"))
	assert.Equal(t, "dbs/mydb/colls/%7B%27collection%27:%27mycoll%27%7D", createCollLink("mydb", "{'collection':'mycoll'}"))
	assert.Equal(t, "dbs/mydb/colls/%3Cdiv%3Emycoll%3C%2Fdiv%3E/docs", createDocsLink("mydb", "<div>mycoll</div>"))
	assert.Equal(t, "dbs/mydb/colls/mycoll/docs/id%3FadditionalParameters=true", createDocLink("mydb", "mycoll", "id?additionalParameters=true"))
	assert.Equal(t, "dbs/mydb/colls/%3Cdiv%3Emycoll%3C%2Fdiv%3E/sprocs", createSprocsLink("mydb", "<div>mycoll</div>"))
	assert.Equal(t, "dbs/mydb/colls/%3Cdiv%3Emycoll%3C%2Fdiv%3E/sprocs/thesproc", createSprocLink("mydb", "<div>mycoll</div>", "thesproc"))
}
