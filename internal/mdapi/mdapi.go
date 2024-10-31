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

type chapter struct {
	Result  string `json:"result"`
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
		//DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}

func DlChapter(chapID string) {
	chapURL := fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", chapID)

	resp, err := http.Get(chapURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve %s", chapURL)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var chap chapter
	if err := dec.Decode(&chap); err != nil {
		log.Fatalf("ERROR: Failed to decode respoonse from %s", chapURL)
	}

	limiter := time.Tick(350 * time.Millisecond)

	for _, pageName := range chap.Chapter.Data {
		pageURL := fmt.Sprintf("%s/data/%s/%s", chap.BaseURL, chap.Chapter.Hash, pageName)
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

func dlPage(pageURL string, f *os.File) {
	img, err := http.Get(pageURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve %s", pageURL)
	}
	defer img.Body.Close()

	// TODO: Track num bytes for error reporting
	if _, err := io.Copy(f, img.Body); err != nil {
		log.Fatalf("ERROR: Failed to write to file %s", f.Name())
	}
}
