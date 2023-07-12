package main

import (
	"fmt"

	arango "github.com/arangodb/go-driver"
)

type TaxonSvc struct {
	db arango.Database
}

func NewTaxonSvc(db arango.Database) *TaxonSvc {
	return &TaxonSvc{db: db}
}

// Get returns a single taxon by ID.
func (svc *TaxonSvc) Get(rank string, id string) (Taxon, error) {
	taxon := Taxon{}
	col, err := svc.db.Collection(nil, rank)
	if err != nil {
		return taxon, err
	}
	_, err = col.ReadDocument(nil, id, &taxon)
	if err != nil {
		return taxon, err
	}
	return taxon, nil
}

// GetChildren returns a list of taxon children.
func (svc *TaxonSvc) GetChildren(rank string, id string) ([]Taxon, error) {
	taxa := []Taxon{}
	query := "FOR v IN 1..1 INBOUND @start GRAPH 'animal_kingdom' RETURN v"
	bindVars := map[string]interface{}{
		"start": fmt.Sprintf("%s/%s", rank, id),
	}
	cursor, err := svc.db.Query(nil, query, bindVars)
	if err != nil {
		return taxa, err
	}
	defer cursor.Close()
	for {
		var taxon Taxon
		_, err := cursor.ReadDocument(nil, &taxon)
		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return taxa, err
		}
		taxa = append(taxa, taxon)
	}
	return taxa, nil
}
