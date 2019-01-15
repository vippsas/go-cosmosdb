// +build !offline

package readfeed

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vippsas/go-cosmosdb/cosmos"
	"github.com/vippsas/go-cosmosdb/cosmostest"
	"gopkg.in/yaml.v2"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

var collection cosmos.Collection
var currentId int

func TestMain(m *testing.M) {
	config := LoadCosmosConfiguration()
	collection = cosmostest.SetupUniqueCollectionWithExistingDatabaseAndDefaultThroughput(log.New(os.Stdout, "", 0), config, "feedtest", "partitionkey")
	retCode := m.Run()
	cosmostest.TeardownCollection(collection)
	os.Exit(retCode)
}

func Test_WhenDocumentsAreInsertedOrUpdatedThenChangeAppearsOnFeed(t *testing.T) {
	givenScenario(t).
		whenNDocumentsAreInsertedOnSamePartition(1).
		thenFeedHasCorrespondingChanges(1000).
		whenNDocumentsAreInsertedOnDifferentPartitions(2).
		thenFeedHasCorrespondingChanges(2).
		whenNDocumentsAreInsertedOnSamePartition(1).
		thenFeedHasCorrespondingChanges(3).
		whenNDocumentsAreInsertedOnSamePartition(50).
		thenFeedHasCorrespondingChanges(1000)
}

type scenario struct {
	t         *testing.T
	etags     map[string]string
	documents []testDocument
}

func givenScenario(t *testing.T) *scenario {
	scenario := scenario{t: t}
	return &scenario
}

func (s *scenario) getPartitionKeyRangeIds() (ids []string) {
	currentRanges, err := collection.GetPartitionKeyRanges()
	assert.NoError(s.t, err)
	for _, r := range currentRanges {
		ids = append(ids, r.Id)
	}
	return ids
}

func (s *scenario) refreshPartitionKeyRanges() *scenario {
	currentRangeIds := s.getPartitionKeyRangeIds()
	refreshedEtags := make(map[string]string)
	for _, currentRangeId := range currentRangeIds {
		refreshedEtags[currentRangeId] = s.etags[currentRangeId]
		delete(s.etags, currentRangeId)
	}
	s.etags = refreshedEtags
	return s
}

func (s *scenario) readFeed(pageSize int) []testDocument {
	var allChanges []testDocument
	for partitionKeyRangeId, etag := range s.etags {
		var changesInPartitionRange []testDocument
		response, err := collection.ReadFeed(etag, partitionKeyRangeId, pageSize, &changesInPartitionRange)
		assert.NoError(s.t, err)
		assert.Empty(s.t, response.Continuation)
		fmt.Printf("Found %d document(s) in partition range <%s> from etag %s (next etag: %s):\n", len(changesInPartitionRange), etag, partitionKeyRangeId, response.Etag)
		if len(changesInPartitionRange) > 0 {
			for _, doc := range changesInPartitionRange {
				fmt.Println(" ", doc)
			}
			s.etags[partitionKeyRangeId] = response.Etag
			allChanges = append(allChanges, changesInPartitionRange...)
		}
	}
	return allChanges
}

func (s *scenario) thenFeedHasCorrespondingChanges(pageSize int) *scenario {
	s.refreshPartitionKeyRanges()
	// First full pages
	numFullPages := len(s.documents) / pageSize
	for page := 0; page < numFullPages; page++ {
		fmt.Printf("Reading full page %d with size %d\n", page, pageSize)
		changes := s.readFeed(pageSize)
		assert.Equal(s.t, pageSize, len(changes), "Expected %d documents on feed but found %d", pageSize, len(changes))
		for i, insertedDocument := range s.documents[page*pageSize : page*pageSize+pageSize] {
			s.assertEqualDocuments(insertedDocument, changes[i])
		}
	}
	// Eventual last non-full page
	lastPageSize := len(s.documents) % pageSize
	if lastPageSize > 0 {
		page := len(s.documents) / pageSize
		fmt.Printf("Reading last page %d with size %d\n", page, lastPageSize)
		changes := s.readFeed(lastPageSize)
		assert.Equal(s.t, lastPageSize, len(changes), "Expected %d documents on feed but found %d", lastPageSize, len(changes))
		for i, insertedDocument := range s.documents[page*pageSize:] {
			assert.Equal(s.t, len(s.documents)%pageSize, len(changes), "Expected %d documents on feed but found %d", len(s.documents)%pageSize, len(changes))
			s.assertEqualDocuments(insertedDocument, changes[i])
		}
	}
	s.documents = nil
	return s
}

func (s *scenario) assertEqualDocuments(document1, document2 testDocument) {
	assert.Equal(s.t, document1.Id, document2.Id)
	assert.Equal(s.t, document1.PartitionKey, document2.PartitionKey)
}

func (s *scenario) whenDocumentIsInserted(document testDocument) *scenario {
	_, fresh, err := testDocumentRepo{Collection: collection}.GetOrCreate(&document)
	assert.NoError(s.t, err)
	assert.True(s.t, fresh)
	fmt.Printf("Inserted document %s\n", document.String())
	s.documents = append(s.documents, document)
	return s
}

func (s *scenario) whenNDocumentsAreInsertedOnSamePartition(n int) *scenario {
	partitionKey := strconv.Itoa(rand.Intn(100000000))
	for i := 0; i < n; i++ {
		currentId += 1
		s.whenDocumentIsInserted(aDocument(strconv.Itoa(currentId), partitionKey, "a text"))
	}
	return s
}

func (s *scenario) whenNDocumentsAreInsertedOnDifferentPartitions(n int) *scenario {
	for i := 0; i < n; i++ {
		currentId += 1
		partitionKey := strconv.Itoa(rand.Intn(100000000))
		s.whenDocumentIsInserted(aDocument(string(currentId), partitionKey, "a text"))
	}
	return s
}

func (s *scenario) whenDocumentsAreInserted(documents ...testDocument) *scenario {
	for _, document := range documents {
		s.whenDocumentIsInserted(document)
	}
	return s
}

func (s *scenario) whenDocumentIsUpdated(id string, partitionKey string, text string) *scenario {
	document, err := testDocumentRepo{Collection: collection}.Update(partitionKey, id, func(d *testDocument) error {
		d.Text = text
		return nil
	})
	assert.NoError(s.t, err)
	assert.Equal(s.t, text, document.Text)
	fmt.Printf("Updated document %s\n", document.String())
	s.documents = append(s.documents, *document)
	return s
}

func aDocument(id, partitionKey string, text string) testDocument {
	return testDocument{BaseModel: cosmos.BaseModel{Id: id}, PartitionKey: partitionKey, Text: text}
}

func LoadCosmosConfiguration() cosmostest.Config {
	var configDoc struct {
		CosmosTest cosmostest.Config `yaml:"CosmosTest"`
	}
	if configfile, err := OpenConfigurationFile(); err != nil {
		panic(fmt.Sprintf("Failed to read test configuration: %v", err))
	} else if err = yaml.NewDecoder(configfile).Decode(&configDoc); err != nil {
		panic(fmt.Sprintf("Failed to parse test configuration: %v", err))
	} else {
		if configDoc.CosmosTest.DbName == "" {
			configDoc.CosmosTest.DbName = "default"
		}
		return configDoc.CosmosTest
	}
}

func OpenConfigurationFile() (*os.File, error) {
	return doOpenConfigurationFile(".")
}

func doOpenConfigurationFile(path string) (*os.File, error) {
	if path, err := filepath.Abs(filepath.Join(path, "testconfig.yaml")); err != nil {
		return nil, err // Fail
	} else if file, err := os.Open(path); err == nil {
		return file, nil // Eureka!
	} else if filepath.Dir(path) == filepath.Dir(filepath.Dir(path)) {
		return nil, err // Fail -- searched up to root directory without finding file
	} else {
		return doOpenConfigurationFile(filepath.Dir(filepath.Dir(path))) // Check parent directory
	}
}
