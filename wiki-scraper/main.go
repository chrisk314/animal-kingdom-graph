package main

import (
	"fmt"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

const seedURL = "https://en.wikipedia.org/wiki/Animal"
const allowedDomain = "en.wikipedia.org"

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
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

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// fmt.Println(link)
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Start crawler at seed url.
	c.Visit(seedURL)

	// Wait for all threads to finish.
	c.Wait()
}
