package main

import (
	"fmt"
	"log"
	"strings"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func createArangoDBClient(config Config) (arango.Client, error) {
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
	return client, nil
}

func createArangoDB(config Config, client arango.Client) (arango.Database, error) {
	var err error
	var db arango.Database
	var exists bool
	exists, err = client.DatabaseExists(nil, config.DatabaseName)
	if exists {
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
	return db, nil
}

func createArangoDBGraph(config Config, db arango.Database) (arango.Graph, error) {
	var err error
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
	return graph, nil
}

func createArangoDBCollections(config Config, graph arango.Graph) (map[string]arango.Collection, error) {
	var err error
	var exists bool
	var coll arango.Collection
	var taxLvlCollNames []string = []string{KingdomCollName, PhylumCollName, ClassCollName, OrderCollName, FamilyCollName, GenusCollName, SpeciesCollName}
	var taxLvlColls map[string]arango.Collection = make(map[string]arango.Collection)

	for _, taxLvlCollName := range taxLvlCollNames {
		exists, err = graph.VertexCollectionExists(nil, taxLvlCollName)

		if !exists {
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
	return taxLvlColls, nil
}

func createArangoDBEdgeCollections(config Config, graph arango.Graph, taxLvlColls map[string]arango.Collection) (map[string]arango.Collection, error) {
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
	return taxLvlColls, nil
}

// GetOrCreateCollections creates ArangoDB collections for all taxonomic levels.
func GetOrCreateCollections(config Config) (arango.Graph, map[string]arango.Collection, error) {
	// Create ArangoDB connection.
	client, err := createArangoDBClient(config)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB client: %v", err)
	}

	// Create ArangoDB database.
	db, err := createArangoDB(config, client)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB database: %v", err)
	}

	// Get or create graph for taxonomic hierarchy.
	graph, err := createArangoDBGraph(config, db)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB graph: %v", err)
	}

	// Create collections for all taxonomic levels.
	taxLvlColls, err := createArangoDBCollections(config, graph)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB collections: %v", err)
	}

	// Create edge collections for all taxonomic levels.
	createArangoDBEdgeCollections(config, graph, taxLvlColls)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB edge collections: %v", err)
	}

	return graph, taxLvlColls, nil
}

func addTaxonToCollection(taxon Taxon, taxLvlColls map[string]arango.Collection) arango.DocumentID {
	coll, ok := taxLvlColls[strings.ToLower(taxon.Rank)]
	if !ok {
		// Taxonomic heirerchy level not tracked in collections.
		fmt.Printf("Skipping taxonomic level '%s'\n", taxon.Rank)
		return ""
	}
	// Check if taxon already exists in collection.
	query := fmt.Sprintf("FOR t IN %s FILTER t.name == '%s' RETURN t", coll.Name(), taxon.Name)
	cursor, err := coll.Database().Query(nil, query, nil)
	if err != nil {
		log.Fatalf("Failed to query collection: %v", err)
	}
	defer cursor.Close()
	var qTaxon Taxon
	var id arango.DocumentID = ""
	for {
		qMeta, err := cursor.ReadDocument(nil, &qTaxon)
		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Failed to read document: %v", err)
		}
		if qTaxon.Name == taxon.Name {
			// Taxon already exists in collection.
			id = qMeta.ID
			fmt.Printf("Found document with id '%s' in collection '%s'\n", id, coll.Name())
			break
		}
	}
	if id == "" {
		// Taxon does not exist in collection. Create it.
		meta, err := coll.CreateDocument(nil, taxon)
		if err != nil {
			log.Fatalf("Failed to create document: %v", err)
		}
		id = meta.ID
		fmt.Printf("Created document with id '%s' in collection '%s'\n", id, coll.Name())
	}
	return id
}

func addParentTaxonLinkToEdgeCollection(taxLvlColls map[string]arango.Collection, id, idParent arango.DocumentID, rankParent string) {
	// Create edge from parent taxon to current taxon.
	edgeCollName := fmt.Sprintf("%sMembers", rankParent)
	edgeColl, ok := taxLvlColls[edgeCollName]
	if !ok {
		// Taxonomic heirerchy level not tracked in collections.
		log.Fatalf("Failed to retreive collection: %v", edgeCollName)
	}
	// Check if edge already exists in collection.
	query := fmt.Sprintf("FOR e IN %s FILTER e._from == '%s' AND e._to == '%s' RETURN e", edgeColl.Name(), id, idParent)
	cursor, err := edgeColl.Database().Query(nil, query, nil)
	if err != nil {
		log.Fatalf("Failed to query collection: %v", err)
	}
	defer cursor.Close()
	var qEdge arango.EdgeDocument
	var idEdge arango.DocumentID = ""
	for {
		qMeta, err := cursor.ReadDocument(nil, &qEdge)
		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Failed to read document: %v", err)
		}
		if qEdge.From == id && qEdge.To == idParent {
			// Edge already exists in collection.
			idEdge = qMeta.ID
			break
		}
	}
	if idEdge == "" {
		// Edge does not exist in collection. Create it.
		_, err := edgeColl.CreateDocument(nil, arango.EdgeDocument{From: id, To: idParent})
		if err != nil {
			fmt.Printf("Failed edge from '%s' to '%s' in collection '%s'\n", id, idParent, edgeColl.Name())
			log.Fatalf("Failed to create document: %v", err)
		}
		fmt.Printf("Created edge from '%s' to '%s' in collection '%s'\n", id, idParent, edgeColl.Name())
	}
}
