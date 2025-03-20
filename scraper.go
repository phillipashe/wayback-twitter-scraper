package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/gocolly/colly"
)

func HandleTweet(*colly.Collector) {
	fmt.Println("Tweet")
}

func main() {
	// scraping logic goes here
	c := colly.NewCollector(
		colly.AllowedDomains("www.changethislater.com"),
	)

	var initial_url string
	flag.StringVar(&initial_url, "initial_url", "", "Initial URL to scrape")
	flag.Parse()

	if initial_url == "" {
		panic(errors.New("initial_url is required"))
	}

	fmt.Printf("initial_url: %s", initial_url)
}
