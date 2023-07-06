package main

import (
	arango "github.com/arangodb/go-driver"
)

type TaxonSvc struct {
	db arango.Database
}

func NewTaxonSvc(db arango.Database) *TaxonSvc {
	return &TaxonSvc{db: db}
}

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
