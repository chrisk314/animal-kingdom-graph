package main

import (
	"errors"
	"fmt"
	"strings"

	arango "github.com/arangodb/go-driver"
)

func checkTaxonSequence(taxLvls []Taxon) error {
	var rankSequence = []string{"Kingdom", "Phylum", "Class", "Order", "Family", "Genus", "Species"}
	var rankMap = make(map[string]bool)
	for _, taxon := range taxLvls {
		rankMap[taxon.Rank] = true
	}
	for _, rank := range rankSequence {
		if _, ok := rankMap[rank]; !ok {
			fmt.Printf("Missing taxonomic level '%s'\n", rank)
			return errors.New("Missing taxonomic level")
		}
	}
	return nil
}

func processTaxon(taxLvls []Taxon, taxLvlColls map[string]arango.Collection) {
	// Check all required taxonomic levels are present.
	err := checkTaxonSequence(taxLvls)
	if err != nil {
		return
	}

	var idParent arango.DocumentID = ""
	var rankParent string = ""

	// Store taxonomic data for all taxonomic levels in ArangoDB.
	for _, taxon := range taxLvls {
		id := addTaxonToCollection(taxon, taxLvlColls)
		if id != "" {
			if idParent != "" {
				addParentTaxonLinkToEdgeCollection(taxLvlColls, id, idParent, rankParent)
			}
			idParent = id
			rankParent = strings.ToLower(taxon.Rank)
		}
	}
}
