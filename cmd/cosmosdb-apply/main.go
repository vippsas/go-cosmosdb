package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vippsas/go-cosmosdb"
	"log"
)

const (
	CosmosDbKeyEnvVarName = "COSMOSDB_KEY"
)

var options struct {
	instanceName string
	filePaths    string

	verbose bool
}

// This tools allows the user to imperatively set up and configure collections in a pre-existing database
func main() {
	flag.StringVar(&options.instanceName, "instanceName", "", "Name of the CosmosDB account/instance")
	flag.StringVar(&options.filePaths, "filePaths", "", "Comma-separated list of files to import. Supports globbing.")
	flag.BoolVar(&options.verbose, "verbose", false, "Enable to get log statements sent to Stdout")

	flag.Parse()

	if options.verbose == true {
		log.SetOutput(os.Stdout)
	}

	validateParameters()

	// Parse file paths
	paths := getPaths(options.filePaths)
	fmt.Printf("The following %d definition file(s) will be processed: %s\n", len(paths), paths)

	masterKey := getCosmosDbMasterKey()

	// --- Input is validated and we're ready to to more expensive tasks

	// Parse all definition files
	collectionDefinitions := getCollectionDefinitions(paths)

	client := newCosmosDbClient(masterKey)

	for i, def := range collectionDefinitions {
		fmt.Printf("[%d/%d] Processing collection '%s'\n", i+1, len(collectionDefinitions), def.CollectionID)
		handleCollectionDefinition(def, client)
		fmt.Printf("[%d/%d] Finished processing collection '%s'\n", i+1, len(collectionDefinitions), def.CollectionID)
	}
}

// --- General helper functions

// Panic with format
func panicf(f string, a ...interface{}) {
	panic(fmt.Sprintf(f, a))
}

// Panic with error and format
func panicef(f string, e error, a ...interface{}) {
	panic(fmt.Sprintf(f+" -> "+e.Error(), a))
}

func getCosmosDbMasterKey() string {
	// Get key from env. vars.
	masterKey, dbKeySet := os.LookupEnv(CosmosDbKeyEnvVarName)

	if !dbKeySet {
		panicf("Environment var. '%s' is not set", CosmosDbKeyEnvVarName)
	}

	return masterKey
}

// Will exit with code 1 if it doesn't validate
func validateParameters() {
	if options.instanceName == "" || options.filePaths == "" {
		fmt.Println("Missing parameters. Use -h to see usage")
		os.Exit(1)
	}
}

// Param 'filePaths' is a comma-separated string
func getPaths(filePaths string) []string {
	var paths []string
	for _, path := range strings.Split(filePaths, ",") {
		foundPaths, err := filepath.Glob(path)
		if err != nil {
			panicef("Error parsing file paths", err)
		}

		paths = append(paths, foundPaths...)
	}
	return paths
}

func newCosmosDbClient(masterKey string) *cosmosdb.Client {
	// Create CosmosDB client
	cosmosCfg := cosmosdb.Config{
		MasterKey: masterKey,
	}
	client := cosmosdb.New(fmt.Sprintf("https://%s.documents.azure.com:443", options.instanceName), cosmosCfg, nil)
	return client
}

// --- Database related

func ensureDatabaseExists(client *cosmosdb.Client, def collectionDefinition) {
	_, err := client.GetDatabase(context.Background(), def.DatabaseID, nil)
	if err != nil {
		log.Printf("Could not get database. Assuming database does not exist.\n")
		_, err = client.CreateDatabase(context.Background(), def.DatabaseID, nil)

		if err != nil {
			panicef("Could not create a new database named '%s'", err, def.DatabaseID)
		}

		log.Printf("Database '%s' was created.\n", def.DatabaseID)
	} else {
		log.Printf("Database '%s' exists.\n", def.DatabaseID)
	}
}

// --- Collection related

func getCollectionDefinitions(filePaths []string) []collectionDefinition {
	colDefs := make([]collectionDefinition, 0, len(filePaths))

	for i := 0; i < len(filePaths); i++ {
		path := filePaths[i]

		// Read file from FS
		content, fileErr := ioutil.ReadFile(path)
		if fileErr != nil {
			panicf("Could not read file '%s' -> %s\n", path, fileErr.Error())
		}

		// Unmarshal JSON to struct
		var colDef []collectionDefinition
		if err := json.Unmarshal(content, &colDef); err != nil {
			panic(err)
		}

		// Set file path for all definitions. Used when looking up source.
		for i := 0; i < len(colDef); i++ {
			colDef[i].FilePath = filepath.Dir(path)
		}

		colDefs = append(colDefs, colDef...)
	}

	return colDefs
}

func handleCollectionDefinition(def collectionDefinition, client *cosmosdb.Client) {
	// We need to check three cases.
	// 1: Added. In definition and not among existing collections.
	// 2: Updated. In both places, but need to be replaced.
	// (3. Removed. Not in definition, but among existing collections.)

	ensureDatabaseExists(client, def)
	dbCol, colFound := getCollection(client, def)

	if !colFound {
		// Collection does not exist
		// NOTE: Offers are created as a part of the collection
		createCollection(def, client)
	} else {
		// Collection exists
		replaceCollection(def, dbCol, client)
		replaceOffers(def, dbCol, client)
	}

	// Check triggers

	collectionTriggers, ltErr := client.ListTriggers(context.Background(), def.DatabaseID, def.CollectionID)
	if ltErr != nil {
		panicef("Could not list triggers for collection '%s' in DB '%s'", ltErr, def.CollectionID, def.DatabaseID)
	}

	for _, trigDef := range def.Triggers {
		_, trigFound := triggerExists(collectionTriggers.Triggers, trigDef.ID)

		if !trigFound {
			createTrigger(trigDef, client, def)
		} else {
			replaceTrigger(trigDef, client, def)
		}
	}

	// TODO: Do the same for UDF as for trigger
	// TODO: Do the same for SPROC as for trigger
}

func getCollection(client *cosmosdb.Client, def collectionDefinition) (*cosmosdb.Collection, bool) {
	dbName := def.DatabaseID
	colName := def.CollectionID

	dbCollection, err := client.GetCollection(context.Background(), dbName, colName)

	if err != nil {
		log.Printf("Could not get collection '%s' in database '%s' -> %s", colName, dbName, err.Error())
		return nil, false
	}

	return dbCollection, true
}

func createCollection(def collectionDefinition, client *cosmosdb.Client) {
	colCreateOpts := cosmosdb.CollectionCreateOptions{
		Id:              def.CollectionID,
		IndexingPolicy:  &def.IndexingPolicy,
		PartitionKey:    &def.PartitionKey,
		OfferType:       cosmosdb.OfferType(def.Offer.Type),
		OfferThroughput: cosmosdb.OfferThroughput(def.Offer.Throughput),
	}

	_, err := client.CreateCollection(context.Background(), def.DatabaseID, colCreateOpts)
	if err != nil {
		panicef("Create collection '%s' failed", err, colCreateOpts.Id)
	}

	log.Printf("Collection created\n")
}

func replaceCollection(def collectionDefinition, existingCol *cosmosdb.Collection, client *cosmosdb.Client) {
	colReplaceOpts := cosmosdb.CollectionReplaceOptions{
		Id:             def.CollectionID,
		IndexingPolicy: &def.IndexingPolicy,
		PartitionKey:   existingCol.PartitionKey,
	}

	updatedCol, err := client.ReplaceCollection(context.Background(), def.DatabaseID, colReplaceOpts)
	if err != nil {
		panicef("Could not replace collection '%s'", err, def.CollectionID)
	}

	log.Printf("Sucsefully updated collection '%s'\n", updatedCol.Id)
}

// --- Triggers related

func triggerExists(triggers []cosmosdb.Trigger, triggerName string) (*cosmosdb.Trigger, bool) {
	for _, c := range triggers {
		if c.Id == triggerName {
			return &c, true
		}
	}

	return nil, false
}

func replaceTrigger(trigDef trigger, client *cosmosdb.Client, def collectionDefinition) {
	opts := cosmosdb.TriggerReplaceOptions{
		Id:        trigDef.ID,
		Type:      cosmosdb.TriggerType(trigDef.TriggerType),
		Operation: cosmosdb.TriggerOperation(trigDef.TriggerOperation),
		Body:      getJavaScriptBody(trigDef.Body, def.FilePath),
	}

	_, trigErr := client.ReplaceTrigger(context.Background(), def.DatabaseID, def.CollectionID, opts)
	if trigErr != nil {
		panicef("Updating trigger '%s' on collection '%s' failed", trigErr, trigDef.ID, def.CollectionID)

	}

	log.Printf("Trigger '%s' was updated\n", trigDef.ID)
}

func createTrigger(trigDef trigger, client *cosmosdb.Client, def collectionDefinition) {
	opts := cosmosdb.TriggerCreateOptions{
		Id:        trigDef.ID,
		Type:      cosmosdb.TriggerType(trigDef.TriggerType),
		Operation: cosmosdb.TriggerOperation(trigDef.TriggerOperation),
		Body:      getJavaScriptBody(trigDef.Body, def.FilePath),
	}

	_, trigErr := client.CreateTrigger(context.Background(), def.DatabaseID, def.CollectionID, opts)

	if trigErr != nil {
		panicef("Creating trigger '%s' on collection '%s' failed", trigErr, trigDef.ID, def.CollectionID)
	}

	log.Printf("Trigger '%s' was created\n", trigDef.ID)
}

func getJavaScriptBody(body triggerBody, directory string) string {
	switch body.SourceLocation {

	case "inline":
		return body.InlineSource

	case "file":

		absFilePath, _ := filepath.Abs(directory)
		filePath := filepath.Join(absFilePath, body.FileName)
		source, err := ioutil.ReadFile(filePath)

		if err != nil {
			panicef("Could not read source file from '%s'", err, filePath)
		}

		return string(source)

	default:
		panicf("Unknown source location '%s' found in trigger definition", body.SourceLocation)
		return ""
	}
}

// --- Offers related

func replaceOffers(def collectionDefinition, dbCol *cosmosdb.Collection, client *cosmosdb.Client) {
	dbOffers, err := client.ListOffers(context.Background(), nil)
	if err != nil {
		panicef("Could not list offers in DB", err)
	}
	for _, off := range dbOffers.Offers {
		if off.OfferResourceId == dbCol.Rid {
			// offer applies to this resource

			offReplOpts := cosmosdb.OfferReplaceOptions{
				Rid:              off.Rid,
				OfferResourceId:  off.OfferResourceId,
				Id:               off.Id,
				OfferVersion:     off.OfferVersion,
				ResourceSelfLink: off.Self,
				OfferType:        cosmosdb.OfferType(def.Offer.Type),
				Content: cosmosdb.OfferThroughputContent{
					Throughput: cosmosdb.OfferThroughput(def.Offer.Throughput),
				},
			}
			_, err := client.ReplaceOffer(context.Background(), offReplOpts, nil)
			if err != nil {
				panicef("Could not update offer '%s'", err, off.Id)
			}

			fmt.Printf("Updated offer '%s'. Throughput=%d, Type=%v", off.Id, def.Offer.Throughput, def.Offer.Type)
		}
	}
}

// --- Inline types used to deserialize the input

type collectionDefinition struct {
	FilePath     string
	DatabaseID   string `json:"databaseId"`
	CollectionID string `json:"collectionId"`
	Offer        struct {
		Throughput int    `json:"throughput"`
		Type       string `json:"type"`
	} `json:"offer"`
	IndexingPolicy cosmosdb.IndexingPolicy `json:"indexingPolicy"`
	PartitionKey   cosmosdb.PartitionKey   `json:"partitionKey"`
	Triggers       []trigger               `json:"triggers"`
	Udfs           []interface{}           `json:"udfs"`
	Sprocs         []interface{}           `json:"sprocs"`
}

type trigger struct {
	ID               string      `json:"id"`
	TriggerType      string      `json:"triggerType"`
	TriggerOperation string      `json:"triggerOperation"`
	Body             triggerBody `json:"body"`
}

type triggerBody struct {
	SourceLocation string `json:"sourceLocation"`
	InlineSource   string `json:"inlineSource,omitempty"`
	FileName       string `json:"fileName,omitempty"`
}
