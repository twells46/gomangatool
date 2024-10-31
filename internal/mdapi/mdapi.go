package mdapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type chapter_meta struct {
	Result  string `json:"result"`
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
		//DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
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

func getChapMetadata(chapID string) chapter_meta {
	chapURL := fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", chapID)

	// Get the image delivery metadata
	resp, err := http.Get(chapURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve %s", chapURL)
	}
	defer resp.Body.Close()

	// Attempt to decode the response into the chapter struct
	dec := json.NewDecoder(resp.Body)
	var chap chapter_meta
	if err := dec.Decode(&chap); err != nil {
		log.Fatalf("ERROR: Failed to decode respoonse from %s", chapURL)
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
