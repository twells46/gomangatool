package mdapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type chapterMeta struct {
	Result  string `json:"result"`
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
	} `json:"chapter"`
}

type SeriesFeed struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     []struct {
		ID         string `json:"id"`
		Attributes struct {
			Volume  string `json:"volume"`
			Chapter string `json:"chapter"`
		} `json:"attributes"`
	} `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// Download a chapter given the chapter's ID
func DlChapter(chapID string) {
	chap := getChapMetadata(chapID)

	// Respect API rate limit
	limiter := time.Tick(350 * time.Millisecond)

	for _, pageName := range chap.Chapter.Data {
		pageURL := fmt.Sprintf("%s/data/%s/%s", chap.BaseURL, chap.Chapter.Hash, pageName)
		// TODO: Include chapter number in folder name
		fname := fmt.Sprintf("%s/%s", chapID, pageName)

		if err := os.MkdirAll(chapID, 0770); err != nil {
			log.Fatalf("ERROR: Failed to create directory %s", chapID)
		}
		f, err := os.Create(fname)
		if err != nil {
			log.Fatalf("ERROR: Failed to create file %s", fname)
		}

		<-limiter

		dlPage(pageURL, f)
	}
}

func getChapMetadata(chapID string) chapterMeta {
	chapURL := fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", chapID)

	// Get the image delivery metadata
	resp, err := http.Get(chapURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve %s", chapURL)
	}
	defer resp.Body.Close()

	// Attempt to decode the response into the chapter struct
	dec := json.NewDecoder(resp.Body)
	var chap chapterMeta
	if err := dec.Decode(&chap); err != nil {
		log.Fatalf("ERROR: Failed to decode response from %s", chapURL)
	}

	return chap
}

func dlPage(pageURL string, f *os.File) {
	img, err := http.Get(pageURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve %s", pageURL)
	}
	defer img.Body.Close()

	if _, err := io.Copy(f, img.Body); err != nil {
		log.Fatalf("ERROR: Failed to write to file %s", f.Name())
	}
}

func GetFeed(seriesID string, offset int) SeriesFeed {
	feedURL := fmt.Sprintf("https://api.mangadex.org/manga/%s/feed", seriesID)
	params := url.Values{}
	params.Add("translatedLanguage[]", "en")
	params.Add("includeExternalUrl", "0")
	params.Add("offset", fmt.Sprint(offset))
	fullURL := fmt.Sprintf("%s?%s", feedURL, params.Encode())

	feedResp, err := http.Get(fullURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve feed %s", fullURL)
	}
	defer feedResp.Body.Close()

	dec := json.NewDecoder(feedResp.Body)
	var m SeriesFeed
	if err := dec.Decode(&m); err != nil {
		log.Fatalf("ERROR: Failed to decode response from %s", fullURL)
	}

	return m
}
