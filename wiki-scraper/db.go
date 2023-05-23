package main

import (
	"fmt"
	"log"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// GetOrCreateCollections creates ArangoDB collections for all taxonomic levels.
func GetOrCreateCollections(config Config) (map[string]arango.Collection, error) {
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

	// Create collections for all taxonomic levels.
	var coll_exists bool
	var taxLvlCollNames []string = []string{PhylumCollName, ClassCollName, OrderCollName, FamilyCollName, GenusCollName, SpeciesCollName}
	var taxLvlColls map[string]arango.Collection = make(map[string]arango.Collection)

	for _, taxLvlCollName := range taxLvlCollNames {
		coll_exists, err = db.CollectionExists(nil, taxLvlCollName)

		var coll arango.Collection
		if !coll_exists {
			coll, err = db.CreateCollection(nil, taxLvlCollName, nil)
			if err != nil {
				log.Fatalf("Failed to create collection: %v", err)
			}
		} else {
			coll, _ = db.Collection(nil, taxLvlCollName)
		}
		taxLvlColls[taxLvlCollName] = coll
	}

	return taxLvlColls, nil
}
