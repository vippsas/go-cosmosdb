// +build !offline

package readfeed

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vippsas/go-cosmosdb/cosmos"
	"github.com/vippsas/go-cosmosdb/cosmostest"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var collection cosmos.Collection

func TestMain(m *testing.M) {
	config := LoadCosmosConfiguration()
	collection = cosmostest.SetupUniqueCollectionWithExistingDatabaseAndDefaultThroughput(log.New(os.Stdout, "", 0), config, "feedtest", "partitionkey")
	retCode := m.Run()
	cosmostest.TeardownCollection(collection)
	os.Exit(retCode)
}

func Test_WhenDocumentsAreInsertedOrUpdatedThenChangeAppearsOnFeed(t *testing.T) {
	givenScenario(t).
		whenDocumentIsInserted(aDocument("1", "a", "abc")).
		thenFeedHasCorrespondingChanges().
		whenDocumentsAreInserted(aDocument("2", "SAasdasd", "sdf"), aDocument("3", "23csfdzadf", "dsf")).
		thenFeedHasCorrespondingChanges().
		whenDocumentIsUpdated("1", "a", "cba").
		thenFeedHasCorrespondingChanges()
}

type scenario struct {
	t         *testing.T
	etags     map[string]string
	documents []Document
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

func (s *scenario) readFeed(maxItemsPerPartionRange int) []Document {
	var allChanges []Document
	for partitionKeyRangeId, etag := range s.etags {
		var changesInPartitionRange []Document
		newEtag, err := collection.ReadFeed(etag, partitionKeyRangeId, maxItemsPerPartionRange, &changesInPartitionRange)
		assert.NoError(s.t, err)
		fmt.Printf("Found %d document(s) in partition range <%s>:\n", len(changesInPartitionRange), partitionKeyRangeId)
		for _, doc := range changesInPartitionRange {
			fmt.Println(" ", doc)
		}
		s.etags[partitionKeyRangeId] = newEtag
		allChanges = append(allChanges, changesInPartitionRange...)
	}
	return allChanges
}

func (s *scenario) thenFeedHasCorrespondingChanges() *scenario {
	s.refreshPartitionKeyRanges()
	changes := s.readFeed(len(s.documents))
	assert.Equal(s.t, len(s.documents), len(changes), "Expected %d documents on feed but found %d", len(s.documents), len(changes))
	for i, insertedDocument := range s.documents {
		s.assertEqualDocuments(insertedDocument, changes[i])
	}
	s.documents = nil
	return s
}

func (s *scenario) assertEqualDocuments(document1, document2 Document) {
	assert.Equal(s.t, document1.Id, document2.Id)
	assert.Equal(s.t, document1.PartitionKey, document2.PartitionKey)
}

func (s *scenario) whenDocumentIsInserted(document Document) *scenario {
	_, fresh, err := DocumentCosmosRepo{Collection: collection}.GetOrCreate(&document)
	assert.NoError(s.t, err)
	assert.True(s.t, fresh)
	fmt.Printf("Inserted document %s\n", document.String())
	s.documents = append(s.documents, document)
	return s
}

func (s *scenario) whenDocumentsAreInserted(documents ...Document) *scenario {
	for _, document := range documents {
		s.whenDocumentIsInserted(document)
	}
	return s
}

func (s *scenario) whenDocumentIsUpdated(id string, partitionKey string, text string) *scenario {
	document, err := DocumentCosmosRepo{Collection: collection}.Update(partitionKey, id, func(d *Document) error {
		d.Text = text
		return nil
	})
	assert.NoError(s.t, err)
	assert.Equal(s.t, text, document.Text)
	fmt.Printf("Updated document %s\n", document.String())
	s.documents = append(s.documents, *document)
	return s
}

func aDocument(id, partitionKey string, text string) Document {
	return Document{BaseModel: cosmos.BaseModel{Id: id}, PartitionKey: partitionKey, Text: text}
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
