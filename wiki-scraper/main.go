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

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.URLFilters(
			regexp.MustCompile(regexURLWikiNoFiles),
		),
		colly.Async(true),
		colly.MaxDepth(1),
	)

	extensions.RandomUserAgent(c)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
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
		}

		fmt.Printf("Processing: %s\nGot: %v\n", e.Request.URL, taxLvls)

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
