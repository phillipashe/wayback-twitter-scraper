package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gocolly/colly"
)

type WaybackData struct {
	Id  string
	Url string
}

var waybackData []WaybackData

// tweetCollector gets all of the data necessary to query the tweets from the Wayback Machine.
func tweetCollector() {
	c := colly.NewCollector(
		colly.AllowedDomains("archive.org", "web.archive.org"),
	)
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

	var initial_url string
	file, err := os.ReadFile("settings.json")
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		panic(err)
	}
	initial_url = data["initial_url"].(string)

	if initial_url == "" {
		panic(errors.New("initial_url is required"))
	}

	// // called before an HTTP request is triggered
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Requesting initial JSON of URLs")
	})

	// // triggered when the scraper encounters an error
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	// fired when the server responds
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Page visited: ", r.Request.URL)
		var resp [][]string

		if err := json.Unmarshal(r.Body, &resp); err != nil {
			panic(err)
		}
		for _, ov := range resp {
			if ov[1] == "text/html" {
				waybackData = append(waybackData, WaybackData{
					Id:  ov[0],
					Url: ov[2],
				})
			}
		}

		// use urls
		for _, url := range waybackData {
			fmt.Println(url)
		}
	})

	err = c.Visit(initial_url)
	if err != nil {
		fmt.Println(err.Error())
	}

	c.Wait()

}

func handleTweets(*colly.Collector) {
	fmt.Println("Tweet")
}

func main() {
	tweetCollector()
}
