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
	url := strings.Join([]string{sUrl.Scheme, sUrl.Host, href}, "")
	return Taxon{Rank: taxLvlStrs[0], Name: taxLvlStrs[1], Url: url}, nil
}

func processTaxon(taxLvls []Taxon, taxLvlColls map[string]arango.Collection) {
	// Store taxonomic data in graph db. Arango db?
	coll := taxLvlColls[strings.ToLower(taxLvls[len(taxLvls)-1].Rank)]
	metas, errs, err := coll.CreateDocuments(nil, taxLvls[len(taxLvls)-1:])

	if err != nil {
		log.Fatalf("Failed to create documents: %v", err)
	} else if err := errs.FirstNonNil(); err != nil {
		log.Fatalf("Failed to create documents: first error: %v", err)
	}

	fmt.Printf("Created document with key '%s' in collection '%s'\n", strings.Join(metas.Keys(), ","), coll.Name())
}

func main() {

	var err error

	// Load config.
	config, err := LoadConfig("./app.env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get ArangoDB collections.
	taxLvlColls, err := GetOrCreateCollections(config)
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
				taxLvls = append(taxLvls, t)
				if t.Rank == "Species" {
					taxLvls[len(taxLvls)-1].Url = e.Request.URL.String()
					fmt.Printf("Processing: %s\nGot: %v\n", e.Request.URL, taxLvls)
					processTaxon(taxLvls, taxLvlColls)
					// Species is a leaf in the tree. Terminate the search here.
					e.Request.Abort()
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
