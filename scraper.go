package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {
	// scraping logic goes here
	_ = colly.NewCollector(
		colly.AllowedDomains("www.scrapingcourse.com"),
	)
	fmt.Println("this works")
}
