package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

const (
	// seedURL = "https://en.wikipedia.org/wiki/Animal"
	seedURL             = "https://en.wikipedia.org/wiki/Eunice_aphroditois"
	allowedDomain       = "en.wikipedia.org"
	regexURLWikiNoFiles = "https://en.wikipedia.org/wiki/[^File:].+"
	maxTreeDepth        = 10
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

type Species struct {
	Rank string `json:"rank"` // TODO : Should Rank field be included in the stored data?
	Name string `json:"name"`
	Url  string `json:"url"`
}

type taxon struct {
	rank string
	name string
}

func createTaxonomicLevelFromSelection(s *goquery.Selection) (taxon, error) {
	taxLvlStrs := strings.Split(s.Text(), ":")
	if len(taxLvlStrs) != 2 {
		return taxon{}, errors.New("Not a taxon")
	}
	for i := range taxLvlStrs {
		taxLvlStrs[i] = strings.TrimSpace(taxLvlStrs[i])
	}
	return taxon{rank: taxLvlStrs[0], name: taxLvlStrs[1]}, nil
}

func processSpecies(taxLvls []taxon) {
	// TODO : Implement this.
	// Store taxonomic data in graph db. Arango db?
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

	for _, taxLevelCollName := range taxLevelCollNames {
		coll_exists, err = db.CollectionExists(nil, taxLevelCollName)

		if !coll_exists {
			_, err = db.CreateCollection(nil, taxLevelCollName, nil)
			if err != nil {
				log.Fatalf("Failed to create collection: %v", err)
			}
		}
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

	extensions.RandomUserAgent(c)

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
		taxLvls := []taxon{t}
		for {
			taxLvlSel = taxLvlSel.Next()
			t, err = createTaxonomicLevelFromSelection(taxLvlSel)
			if err != nil {
				break
			}
			taxLvls = append(taxLvls, t)
			if t.rank == "Species" {
				fmt.Printf("Processing: %s\nGot: %v\n", e.Request.URL, taxLvls)
				processSpecies(taxLvls)
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
