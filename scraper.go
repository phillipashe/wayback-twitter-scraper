package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type WaybackData struct {
	Id  string
	Url string
}

var waybackData []WaybackData
var waybackUrl string

func extractFilename(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}

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
	fmt.Println("Starting tweet collector")
	c := colly.NewCollector(
		colly.AllowedDomains("archive.org", "web.archive.org"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
	)

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		// Redirects mean it's a retweet
		if r.StatusCode >= 300 && r.StatusCode < 400 {
			fmt.Println("Skipping page because it is a redirect or retweet: ", r.Request.URL)
			return
		}
	})

	c.OnHTML(".js-tweet-text-container", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
	})

	c.OnHTML(".AdaptiveMedia-container", func(e *colly.HTMLElement) {
		e.ForEach("img", func(_ int, el *colly.HTMLElement) {
			// get the URL for each image
			imgURL := el.Attr("src")
			resp, err := http.Get(imgURL)
			// ugly if/else here because colly doesn't have a continue statement in the foreach
			if err == nil || resp.StatusCode == 404 {
				defer resp.Body.Close()
				file, err := os.Create(fmt.Sprintf("./images/%s", extractFilename(imgURL)))
				if err == nil {
					defer file.Close()
					file.ReadFrom(resp.Body)

					fmt.Printf("image found at: %s\n", waybackUrl)
				} else {
					fmt.Printf("failed to write file: %s\n", imgURL)
				}

			} else {
				fmt.Printf("failed to get image: %s\n", imgURL)
			}
		})
	})

	for _, data := range waybackData {
		waybackUrl = fmt.Sprintf("https://web.archive.org/web/%s/%s", data.Id, data.Url)
		if err := c.Visit(waybackUrl); err != nil {
			fmt.Println(err.Error())
		}
	}
}

func main() {
	tweetCollector()
	handleTweets()
}
