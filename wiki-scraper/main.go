package main

// From https://en.wikipedia.org/wiki/Animal ...
// As of 2022, 2.16 million living animal species have been described — of which around
// 1.05 million are insects, over 85,000 are molluscs, and around 65,000 are vertebrates
// — but it has been estimated there are around 7.77 million animal species in total.

import (
	"log"
)

func main() {

	var err error

	// Load config.
	config, err := LoadConfig("./app.env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get ArangoDB collections.
	_, taxLvlColls, err := GetOrCreateCollections(config)
	if err != nil {
		log.Fatalf("Failed to create collections: %v", err)
	}

	// Create Colly crawler.
	c := createCollyCrawler(config, taxLvlColls)

	// Start crawler at seed url.
	c.Visit(config.CrawlerSeedURL)

	// Wait for all threads to finish.
	c.Wait()
}
