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
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
	)

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

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Requesting initial JSON of URLs")
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		var resp [][]string

		if err := json.Unmarshal(r.Body, &resp); err != nil {
			panic(err)
		}
		for _, ov := range resp {
			// only get text pages, because these contain the full tweet data
			if ov[1] == "text/html" {
				waybackData = append(waybackData, WaybackData{
					Id:  ov[2],
					Url: ov[0],
				})
			}
		}
	})

	err = c.Visit(initial_url)
	if err != nil {
		fmt.Println(err.Error())
	}

	c.Wait()
}

func handleTweets() {
	c := colly.NewCollector(
		colly.AllowedDomains("archive.org", "web.archive.org"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s\n", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Page visited: ", r.Request.URL)
		// Redirects mean it's a retweet
		if r.StatusCode >= 300 && r.StatusCode < 400 {
			return
		}
	})

	c.OnHTML(".AdaptiveMedia-container", func(e *colly.HTMLElement) {
		e.ForEach("img", func(_ int, el *colly.HTMLElement) {
			imgURL := el.Attr("src")
			fmt.Println("Printing image")
			fmt.Println(imgURL)
		})
	})

	for _, data := range waybackData {
		if err := c.Visit(fmt.Sprintf("https://web.archive.org/web/%s/%s", data.Id, data.Url)); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(data.Url)
	}
}

func main() {
	tweetCollector()
	handleTweets()
}
