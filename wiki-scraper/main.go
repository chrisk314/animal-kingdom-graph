package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// const seedURL = "https://en.wikipedia.org/wiki/Animal"
const seedURL = "https://en.wikipedia.org/wiki/Eunice_aphroditois"
const allowedDomain = "en.wikipedia.org"
const regexURLWikiNoFiles = "https://en.wikipedia.org/wiki/[^File:].+"
const maxTreeDepth = 10
const async = true
const parallelism = 100 // TODO : Look into Wiki rate limits and mitigation strategies.

type taxonomicLevel struct {
	level string
	value string
}

func createTaxonomicLevelFromSelection(s *goquery.Selection) (taxonomicLevel, error) {
	taxLvlStrs := strings.Split(s.Text(), ":")
	if len(taxLvlStrs) != 2 {
		return taxonomicLevel{}, errors.New("Not a taxonomic level")
	}
	for i := range taxLvlStrs {
		taxLvlStrs[i] = strings.TrimSpace(taxLvlStrs[i])
	}
	return taxonomicLevel{level: taxLvlStrs[0], value: taxLvlStrs[1]}, nil
}

func processSpecies(taxLvls []taxonomicLevel) {
	// TODO : Implement this.
	// Store taxonomic data in graph db. Arango db?
}

func main() {
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

	c.OnHTML("#bodyContent", func(e *colly.HTMLElement) {
		infoboxBiota := e.DOM.Find("table.infobox.biota")
		if infoboxBiota.Length() != 1 {
			return // No table.infobox.biota => this search path is a dead end.
		}

		taxLvlSel := infoboxBiota.Find("tr:contains('Kingdom')")
		tl, err := createTaxonomicLevelFromSelection(taxLvlSel)
		taxLvls := []taxonomicLevel{tl}
		for {
			taxLvlSel = taxLvlSel.Next()
			tl, err = createTaxonomicLevelFromSelection(taxLvlSel)
			if err != nil {
				break
			}
			taxLvls = append(taxLvls, tl)
			if tl.level == "Species" {
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
