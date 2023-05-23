package main

import (
	"fmt"
	"log"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// GetOrCreateCollections creates ArangoDB collections for all taxonomic levels.
func GetOrCreateCollections(config Config) (arango.Graph, map[string]arango.Collection, error) {
	// Create ArangoDB connection.
	var err error
	var client arango.Client
	var conn arango.Connection

	conn, err = http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{config.DatabaseUrl},
	})
	if err != nil {
		log.Fatalf("Failed to create HTTP connection: %v", err)
	}
	client, err = arango.NewClient(arango.ClientConfig{
		Connection:     conn,
		Authentication: arango.BasicAuthentication(config.DatabaseUser, config.DatabasePassword),
	})

	// Create ArangoDB database.
	var db arango.Database
	var db_exists bool

	db_exists, err = client.DatabaseExists(nil, config.DatabaseName)

	if db_exists {
		fmt.Println("That db exists already")
		db, err = client.Database(nil, config.DatabaseName)
		if err != nil {
			log.Fatalf("Failed to open existing database: %v", err)
		}
	} else {
		db, err = client.CreateDatabase(nil, config.DatabaseName, nil)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
	}

	// Get or create graph for taxonomic hierarchy.
	var graph arango.Graph
	graph, err = db.Graph(nil, GraphName)
	if arango.IsNotFound(err) {
		// Graph does not exist yet.
		graph, err = db.CreateGraph(nil, GraphName, &arango.CreateGraphOptions{
			EdgeDefinitions: []arango.EdgeDefinition{
				{Collection: PhylumCollName + "Members", From: []string{PhylumCollName}, To: []string{ClassCollName}},
				{Collection: ClassCollName + "Members", From: []string{ClassCollName}, To: []string{OrderCollName}},
				{Collection: OrderCollName + "Members", From: []string{OrderCollName}, To: []string{FamilyCollName}},
				{Collection: FamilyCollName + "Members", From: []string{FamilyCollName}, To: []string{GenusCollName}},
				{Collection: GenusCollName + "Members", From: []string{GenusCollName}, To: []string{SpeciesCollName}},
			},
		})
		if err != nil {
			log.Fatalf("Failed to create graph: %v", err)
		}
		fmt.Println("Created Graph with name: ", graph.Name())
	} else if err != nil {
		log.Fatalf("Failed to open graph: %v", err)
	} else {
		fmt.Println("Found Graph with name: ", graph.Name())
	}

	// Create collections for all taxonomic levels.
	var collExists bool
	var taxLvlCollNames []string = []string{PhylumCollName, ClassCollName, OrderCollName, FamilyCollName, GenusCollName, SpeciesCollName}
	var taxLvlColls map[string]arango.Collection = make(map[string]arango.Collection)

	for _, taxLvlCollName := range taxLvlCollNames {
		collExists, err = db.CollectionExists(nil, taxLvlCollName)

		var coll arango.Collection
		if !collExists {
			coll, err = db.CreateCollection(nil, taxLvlCollName, nil)
			if err != nil {
				log.Fatalf("Failed to create collection: %v", err)
			}
		} else {
			coll, _ = db.Collection(nil, taxLvlCollName)
		}
		taxLvlColls[taxLvlCollName] = coll
	}

	// Create edge collections for all taxonomic levels.
	var taxLvlEdgeCollNames []string = []string{
		PhylumCollName + "Members",
		ClassCollName + "Members",
		OrderCollName + "Members",
		FamilyCollName + "Members",
		GenusCollName + "Members",
	}

	for _, taxLvlEdgeCollName := range taxLvlEdgeCollNames {
		collExists, err = db.CollectionExists(nil, taxLvlEdgeCollName)

		var edgeColl arango.Collection
		if !collExists {
			edgeColl, err = db.CreateCollection(nil, taxLvlEdgeCollName, &arango.CreateCollectionOptions{
				Type: arango.CollectionTypeEdge,
			})
			if err != nil {
				log.Fatalf("Failed to create edge collection: %v", err)
			}
		} else {
			edgeColl, _ = db.Collection(nil, taxLvlEdgeCollName)
		}
		taxLvlColls[taxLvlEdgeCollName] = edgeColl
	}

	return graph, taxLvlColls, nil
}
