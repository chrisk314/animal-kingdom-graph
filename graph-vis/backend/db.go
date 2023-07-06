package main

import (
	"fmt"
	"log"

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

func GetArangoDB(config Config) (arango.Database, error) {
	var err error
	var client arango.Client
	var db arango.Database
	client, err = createArangoDBClient(config)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB client: %v", err)
	}
	db, err = createArangoDB(config, client)
	if err != nil {
		log.Fatalf("Failed to create ArangoDB: %v", err)
	}
	return db, nil
}
