package main

import (
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

func createTaxonomicLevelFromSelection(s *goquery.Selection) taxonomicLevel {
	taxLvlStrs := strings.Split(s.Text(), ":")
	for i := range taxLvlStrs {
		taxLvlStrs[i] = strings.TrimSpace(taxLvlStrs[i])
	}
	return taxonomicLevel{level: taxLvlStrs[0], value: taxLvlStrs[1]}
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
		Parallelism: 10,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting:", r.URL)
	})

	c.OnHTML("#bodyContent", func(e *colly.HTMLElement) {
		infoboxBiota := e.DOM.Find("table.infobox.biota")
		if infoboxBiota.Length() != 1 {
			return // No table.infobox.biota => this search path is a dead end.
		}

		taxLvlSel := infoboxBiota.Find("tr:contains('Kingdom')")
		taxLvls := []taxonomicLevel{createTaxonomicLevelFromSelection(taxLvlSel)}
		for !strings.Contains(taxLvlSel.Text(), "Species") {
			taxLvlSel = taxLvlSel.Next()
			taxLvls = append(taxLvls, createTaxonomicLevelFromSelection(taxLvlSel))
		}

		fmt.Println(taxLvls)

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
