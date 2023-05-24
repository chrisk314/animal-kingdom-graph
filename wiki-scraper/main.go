package main

// From https://en.wikipedia.org/wiki/Animal ...
// As of 2022, 2.16 million living animal species have been described — of which around
// 1.05 million are insects, over 85,000 are molluscs, and around 65,000 are vertebrates
// — but it has been estimated there are around 7.77 million animal species in total.

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	colly_ext "github.com/gocolly/colly/extensions"

	arango "github.com/arangodb/go-driver"
)

type Taxon struct {
	Rank string `json:"rank"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func createTaxonomicLevelFromSelection(s *goquery.Selection, sUrl url.URL) (Taxon, error) {
	taxLvlStrs := strings.Split(s.Text(), ":")
	if len(taxLvlStrs) != 2 {
		return Taxon{}, errors.New("Not a taxon")
	}
	for i := range taxLvlStrs {
		taxLvlStrs[i] = strings.TrimSpace(taxLvlStrs[i])
	}
	href := s.Children().Find("a[href]").First().AttrOr("href", "")
	hrefUrl, err := url.Parse(href)
	if err != nil {
		log.Fatalf("Failed to parse url: %v", err)
	}
	url := sUrl.ResolveReference(hrefUrl).String()
	return Taxon{Rank: taxLvlStrs[0], Name: taxLvlStrs[1], Url: url}, nil
}

func processTaxon(taxLvls []Taxon, taxLvlColls map[string]arango.Collection) {
	// Store taxonomic data for all taxonomic levels in ArangoDB.
	var idParent arango.DocumentID = ""
	var rankParent string = ""

	// Check all required taxonomic levels are present.
	var rankSequence = []string{"Kingdom", "Phylum", "Class", "Order", "Family", "Genus", "Species"}
	var rankMap = make(map[string]bool)
	for _, taxon := range taxLvls {
		rankMap[taxon.Rank] = true
	}
	for _, rank := range rankSequence {
		if _, ok := rankMap[rank]; !ok {
			fmt.Printf("Missing taxonomic level '%s'\n", rank)
			return
		}
	}

	// Store taxonomic data for all taxonomic levels in ArangoDB.
	for _, taxon := range taxLvls {
		coll, ok := taxLvlColls[strings.ToLower(taxon.Rank)]
		if !ok {
			// Taxonomic heirerchy level not tracked in collections.
			fmt.Printf("Skipping taxonomic level '%s'\n", taxon.Rank)
			continue
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
		if idParent != "" {
			// Create edge from parent taxon to current taxon.
			edgeCollName := fmt.Sprintf("%sMembers", rankParent)
			edgeColl, ok := taxLvlColls[edgeCollName]
			if !ok {
				// Taxonomic heirerchy level not tracked in collections.
				continue
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
		idParent = id
		rankParent = strings.ToLower(taxon.Rank)
	}
}

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
	c := colly.NewCollector(
		colly.AllowedDomains(config.CrawlerAllowedDomain),
		colly.URLFilters(
			regexp.MustCompile(config.CrawlerRegexURLWikiNoFiles),
		),
		colly.Async(config.CrawlerAsync),
		colly.MaxDepth(config.CrawlerMaxTreeDepth),
	)

	colly_ext.RandomUserAgent(c)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.CrawlerParallelism,
	})

	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting:", r.URL)
	// })

	// HTML handler function.
	c.OnHTML("#bodyContent", func(e *colly.HTMLElement) {
		infoboxBiota := e.DOM.Find("table.infobox.biota")
		if infoboxBiota.Length() != 1 {
			return // No table.infobox.biota => this search path is a dead end.
		}

		species := infoboxBiota.Find("tr:contains('Species')")
		if species.Length() != 0 {
			taxLvlSel := infoboxBiota.Find("tr:contains('Kingdom')")
			taxLvls := []Taxon{}
			for {
				t, err := createTaxonomicLevelFromSelection(taxLvlSel, *e.Request.URL)
				if err != nil {
					break
				}
				if t.Rank == "Kingdom" && t.Name != config.KingdomName {
					// Not an animal.
					return
				}
				taxLvls = append(taxLvls, t)
				if t.Rank == "Species" {
					taxLvls[len(taxLvls)-1].Url = e.Request.URL.String()
					fmt.Printf("Processing: %s\nGot: %v\n", e.Request.URL, taxLvls)
					processTaxon(taxLvls, taxLvlColls)
					// Species is a leaf in the tree. Terminate the search here.
					return
				}
				taxLvlSel = taxLvlSel.Next()
			}
		}

		// Limit visited links to those in table.infobox.biota.
		infoboxBiota.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if exists {
				c.Visit(e.Request.AbsoluteURL(link))
			}
		})
	})

	// Start crawler at seed url.
	c.Visit(config.CrawlerSeedURL)

	// Wait for all threads to finish.
	c.Wait()
}
