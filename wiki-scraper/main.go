package main

// From https://en.wikipedia.org/wiki/Animal ...
// As of 2022, 2.16 million living animal species have been described — of which around
// 1.05 million are insects, over 85,000 are molluscs, and around 65,000 are vertebrates
// — but it has been estimated there are around 7.77 million animal species in total.

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	colly_ext "github.com/gocolly/colly/extensions"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

const (
	// seedURL = "https://en.wikipedia.org/wiki/Animal"
	seedURL             = "https://en.wikipedia.org/wiki/Eunice_aphroditois"
	allowedDomain       = "en.wikipedia.org"
	regexURLWikiNoFiles = "https://en.wikipedia.org/wiki/[^File:].+"
	maxTreeDepth        = 10 // TODO : Most optimal search for full list of species.
	async               = true
	parallelism         = 100 // TODO : Look into Wiki rate limits and mitigation strategies.

	DatabaseUrl      = "http://localhost:8529"
	DatabaseUser     = "root"
	DatabasePassword = "password"
	DatabaseName     = "animal_kingdom"
	PhylumCollName   = "pyhlum"
	ClassCollName    = "class"
	OrderCollName    = "order"
	FamilyCollName   = "family"
	GenusCollName    = "genus"
	SpeciesCollName  = "species"
)

type Taxon struct {
	Rank string `json:"rank"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func createTaxonomicLevelFromSelection(s *goquery.Selection) (Taxon, error) {
	taxLvlStrs := strings.Split(s.Text(), ":")
	if len(taxLvlStrs) != 2 {
		return Taxon{}, errors.New("Not a taxon")
	}
	for i := range taxLvlStrs {
		taxLvlStrs[i] = strings.TrimSpace(taxLvlStrs[i])
	}
	url := s.Children().Find("a[href]").First().AttrOr("href", "")
	url = strings.Join([]string{"https://", allowedDomain, url}, "")
	return Taxon{Rank: taxLvlStrs[0], Name: taxLvlStrs[1], Url: url}, nil
}

func processTaxon(taxLvls []Taxon, coll driver.Collection) {
	// Store taxonomic data in graph db. Arango db?
	metas, errs, err := coll.CreateDocuments(nil, taxLvls[len(taxLvls)-1:])

	if err != nil {
		log.Fatalf("Failed to create documents: %v", err)
	} else if err := errs.FirstNonNil(); err != nil {
		log.Fatalf("Failed to create documents: first error: %v", err)
	}

	fmt.Printf("Created document with key '%s' in collection '%s'\n", strings.Join(metas.Keys(), ","), coll.Name())
}

func main() {

	// Create ArangoDB connection.
	var err error
	var client driver.Client
	var conn driver.Connection

	conn, err = http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{DatabaseUrl},
		//Endpoints: []string{"https://5a812333269f.arangodb.cloud:8529/"},
	})
	if err != nil {
		log.Fatalf("Failed to create HTTP connection: %v", err)
	}
	client, err = driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(DatabaseUser, DatabasePassword),
		//Authentication: driver.BasicAuthentication("root", "wnbGnPpCXHwbP"),
	})

	// Create ArangoDB database.
	var db driver.Database
	var db_exists bool

	db_exists, err = client.DatabaseExists(nil, DatabaseName)

	if db_exists {
		fmt.Println("That db exists already")
		db, err = client.Database(nil, DatabaseName)
		if err != nil {
			log.Fatalf("Failed to open existing database: %v", err)
		}
	} else {
		db, err = client.CreateDatabase(nil, DatabaseName, nil)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
	}

	// Create collections for all taxonomic levels.
	var coll_exists bool
	var taxLevelCollNames []string = []string{PhylumCollName, ClassCollName, OrderCollName, FamilyCollName, GenusCollName, SpeciesCollName}
	var taxLevelColls map[string]driver.Collection = make(map[string]driver.Collection)

	for _, taxLevelCollName := range taxLevelCollNames {
		coll_exists, err = db.CollectionExists(nil, taxLevelCollName)

		var coll driver.Collection
		if !coll_exists {
			coll, err = db.CreateCollection(nil, taxLevelCollName, nil)
			if err != nil {
				log.Fatalf("Failed to create collection: %v", err)
			}
		} else {
			coll, _ = db.Collection(nil, taxLevelCollName)
		}
		taxLevelColls[taxLevelCollName] = coll
	}

	// Create Colly crawler.
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.URLFilters(
			regexp.MustCompile(regexURLWikiNoFiles),
		),
		colly.Async(async),
		colly.MaxDepth(maxTreeDepth),
	)

	colly_ext.RandomUserAgent(c)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: parallelism,
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

		taxLvlSel := infoboxBiota.Find("tr:contains('Kingdom')")
		t, err := createTaxonomicLevelFromSelection(taxLvlSel)
		taxLvls := []Taxon{t}
		for {
			taxLvlSel = taxLvlSel.Next()
			t, err = createTaxonomicLevelFromSelection(taxLvlSel)
			if err != nil {
				break
			}
			taxLvls = append(taxLvls, t)
			if t.Rank == "Species" {
				// TODO : Add URLs to all taxonomic levels.
				taxLvls[len(taxLvls)-1].Url = e.Request.URL.String()
				fmt.Printf("Processing: %s\nGot: %v\n", e.Request.URL, taxLvls)
				processTaxon(taxLvls, taxLevelColls[strings.ToLower(t.Rank)])
				// Species is a leaf in the tree. Terminate the search here.
				e.Request.Abort()
			}
		}

		e.DOM.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if exists {
				c.Visit(e.Request.AbsoluteURL(link))
			}
		})
	})

	// Start crawler at seed url.
	c.Visit(seedURL)

	// Wait for all threads to finish.
	c.Wait()
}
