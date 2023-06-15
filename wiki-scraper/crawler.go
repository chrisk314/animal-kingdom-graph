package main

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

func buildCrawlerOnHTML(c *colly.Collector, config Config, taxLvlColls map[string]arango.Collection) func(*colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
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
	}
}

// CreateCollyCrawler creates a Colly crawler for extracting taxonomic data
// from Wikipedia animal species pages.
func CreateCollyCrawler(config Config, taxLvlColls map[string]arango.Collection) *colly.Collector {
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
	c.OnHTML("#bodyContent", buildCrawlerOnHTML(c, config, taxLvlColls))

	return c
}
