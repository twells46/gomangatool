package backend

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

type MangaMeta struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Title struct {
				En string `json:"en"`
			} `json:"title"`
			AltTitles []struct {
				Ja   string `json:"ja,omitempty"`
				JaRo string `json:"ja-ro,omitempty"`
				En   string `json:"en,omitempty"`
			} `json:"altTitles"`
			Description struct {
				En string `json:"en"`
			} `json:"description"`
			PublicationDemographic string `json:"publicationDemographic"`
			Status                 string `json:"status"`
			Tags                   []struct {
				//ID string `json:"id"`
				//Type       string `json:"type"`
				Attributes struct {
					Name struct {
						En string `json:"en"`
					} `json:"name"`
				} `json:"attributes"`
			} `json:"tags"`
		} `json:"attributes"`
	} `json:"data"`
}

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
			log.Fatalf("%s: Failed to create directory %s", err, chapID)
		}
		f, err := os.Create(fname)
		if err != nil {
			log.Fatalf("%s: Failed to create file %s", err, fname)
		}

		<-limiter

		dlPage(pageURL, f)
	}
}

// Pull and decode chapter metadata
func getChapMetadata(chapID string) chapterMeta {
	chapURL := fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", chapID)

	// Get the image delivery metadata
	resp, err := http.Get(chapURL)
	if err != nil {
		log.Fatalf("%s: Failed to retrieve %s", err, chapURL)
	}
	defer resp.Body.Close()

	// Attempt to decode the response into the chapter struct
	dec := json.NewDecoder(resp.Body)
	var chap chapterMeta
	if err := dec.Decode(&chap); err != nil {
		log.Fatalf("%s: Failed to decode response from %s", err, chapURL)
	}

	return chap
}

// Download a single page to the file
func dlPage(pageURL string, f *os.File) {
	img, err := http.Get(pageURL)
	if err != nil {
		log.Fatalf("%s: Failed to retrieve %s", err, pageURL)
	}
	defer img.Body.Close()

	if _, err := io.Copy(f, img.Body); err != nil {
		log.Fatalf("%s: Failed to write to file %s", err, f.Name())
	}
}

// Pull the feed for a series
func GetFeed(seriesID string, offset int) SeriesFeed {
	feedURL := fmt.Sprintf("https://api.mangadex.org/manga/%s/feed", seriesID)
	params := url.Values{}
	params.Add("translatedLanguage[]", "en")
	params.Add("includeExternalUrl", "0")
	params.Add("offset", fmt.Sprint(offset))
	fullURL := fmt.Sprintf("%s?%s", feedURL, params.Encode())

	feedResp, err := http.Get(fullURL)
	if err != nil {
		log.Fatalf("%s: Failed to retrieve feed %s", err, fullURL)
	}
	defer feedResp.Body.Close()

	dec := json.NewDecoder(feedResp.Body)
	var m SeriesFeed
	if err := dec.Decode(&m); err != nil {
		log.Fatalf("%s: Failed to decode response from %s", err, fullURL)
	}

	return m
}

func PullMangaMeta(MangaID string) MangaMeta {
	url := fmt.Sprintf("https://api.mangadex.org/manga/%s", MangaID)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("%s: Failed to retrieve series from %s", err, url)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	var m MangaMeta
	if err := dec.Decode(&m); err != nil {
		log.Fatalf("%s: Failed to decode response from %s", err, url)
	}

	return m
}

func NewManga(MangaID string) {
	meta := PullMangaMeta(MangaID)
	title, abbrev := parseTitle(&meta)
	parseTags(&meta)

	m := Manga{
		MangaID:   MangaID,
		SerTitle:  abbrev,
		FullTitle: title,
		Descr:     meta.Data.Attributes.Description.En,
	}

	fmt.Println(m)
}

func parseTitle(meta *MangaMeta) (string, string) {
	titleOptions := []string{meta.Data.Attributes.Title.En}
	for _, v := range meta.Data.Attributes.AltTitles {
		if len(v.En) > 0 {
			titleOptions = append(titleOptions, v.En)
		} else if len(v.Ja) > 0 {
			titleOptions = append(titleOptions, v.Ja)
		} else if len(v.JaRo) > 0 {
			titleOptions = append(titleOptions, v.JaRo)
		}
	}
	var n int
	fmt.Println("Please choose a title:")
	for i, v := range titleOptions {
		fmt.Printf("[%d] %s\n", i, v)
	}

	fmt.Print("Your choice: ")
	fmt.Scanln(&n)
	fmt.Printf("You chose: %s\n", titleOptions[n])

	var abbrev string
	fmt.Print("What should the abbreviated title be? ")
	fmt.Scanln(&abbrev)
	return titleOptions[n], abbrev
}

func parseTags(meta *MangaMeta) []Tag {
	tagNames := make([]string, 2)
	for _, v := range meta.Data.Attributes.Tags {
		tagNames = append(tagNames, v.Attributes.Name.En)
	}

	return nil
}
