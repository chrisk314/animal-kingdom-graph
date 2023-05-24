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
	var dbExists bool

	dbExists, err = client.DatabaseExists(nil, config.DatabaseName)

	if dbExists {
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
	graph, err = db.Graph(nil, config.GraphName)
	if arango.IsNotFound(err) {
		// Graph does not exist yet.
		graph, err = db.CreateGraph(nil, config.GraphName, &arango.CreateGraphOptions{
			EdgeDefinitions: []arango.EdgeDefinition{
				{Collection: KingdomCollName + "Members", To: []string{KingdomCollName}, From: []string{PhylumCollName}},
				{Collection: PhylumCollName + "Members", To: []string{PhylumCollName}, From: []string{ClassCollName}},
				{Collection: ClassCollName + "Members", To: []string{ClassCollName}, From: []string{OrderCollName}},
				{Collection: OrderCollName + "Members", To: []string{OrderCollName}, From: []string{FamilyCollName}},
				{Collection: FamilyCollName + "Members", To: []string{FamilyCollName}, From: []string{GenusCollName}},
				{Collection: GenusCollName + "Members", To: []string{GenusCollName}, From: []string{SpeciesCollName}},
			},
		})
		if err != nil {
			log.Fatalf("Failed to create Graph: %v", err)
		}
		fmt.Println("Created Graph with name: ", graph.Name())
	} else if err != nil {
		log.Fatalf("Failed to open Graph: %v", err)
	} else {
		fmt.Println("Found Graph with name: ", graph.Name())
	}

	// Create collections for all taxonomic levels.
	var collExists bool
	var coll arango.Collection
	var taxLvlCollNames []string = []string{KingdomCollName, PhylumCollName, ClassCollName, OrderCollName, FamilyCollName, GenusCollName, SpeciesCollName}
	var taxLvlColls map[string]arango.Collection = make(map[string]arango.Collection)

	for _, taxLvlCollName := range taxLvlCollNames {
		collExists, err = graph.VertexCollectionExists(nil, taxLvlCollName)

		if !collExists {
			coll, err = graph.CreateVertexCollection(nil, taxLvlCollName)
			if err != nil {
				log.Fatalf("Failed to create collection: %v", err)
			}
			fmt.Printf("Created collection '%s'\n", coll.Name())
		} else {
			coll, _ = graph.VertexCollection(nil, taxLvlCollName)
			fmt.Printf("Using existing collection '%s'\n", coll.Name())
		}
		taxLvlColls[taxLvlCollName] = coll
	}

	// Create edge collections for all taxonomic levels.
	var taxLvlEdgeCollNames []string = []string{
		KingdomCollName + "Members",
		PhylumCollName + "Members",
		ClassCollName + "Members",
		OrderCollName + "Members",
		FamilyCollName + "Members",
		GenusCollName + "Members",
	}

	for _, taxLvlEdgeCollName := range taxLvlEdgeCollNames {
		coll, _, err := graph.EdgeCollection(nil, taxLvlEdgeCollName)
		if err != nil {
			log.Fatalf("Failed to select edge collection: %v", err)
		}
		fmt.Printf("Using existing edge collection '%s'\n", coll.Name())
		taxLvlColls[taxLvlEdgeCollName] = coll
	}

	return graph, taxLvlColls, nil
}
