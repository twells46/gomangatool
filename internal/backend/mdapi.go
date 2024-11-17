package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
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

type FeedChData struct {
	ID string `json:"id"`
	//Type       string `json:"type"`
	Attributes struct {
		Title   string `json:"title"`
		Volume  string `json:"volume"`
		Chapter string `json:"chapter"`
	} `json:"attributes"`
}

type SeriesFeed struct {
	Result   string       `json:"result"`
	Response string       `json:"response"`
	Data     []FeedChData `json:"data"`
	Limit    int          `json:"limit"`
	Offset   int          `json:"offset"`
	Total    int          `json:"total"`
}

// Download a chapter given the chapter's ID
func dlChapter(c Chapter, store *SQLite) Chapter {
	chap := getChapMetadata(c.ChapterHash)

	// The regex takes a page name from the API like this:
	// x6-23b96047cdd7217e5f493894de6d536afa046e7a33695e539a6960e2a7304d35.jpg
	// and turns it into this:
	// 6.jpg
	pageNameCleaner := regexp.MustCompile(`^[A-z]?([0-9]+)-.*(\.[a-z]*)`)

	dirName := fmt.Sprintf("%02d/%05.1f-%s", c.VolumeNum, c.ChapterNum, c.ChapterHash)
	if err := os.MkdirAll(dirName, 0770); err != nil {
		log.Fatalf("%s: Failed to create directory %s", err, c.ChapterHash)
	}

	// Respect API rate limit
	limiter := time.Tick(350 * time.Millisecond)

	for _, pageName := range chap.Chapter.Data {
		pageURL := fmt.Sprintf("%s/data/%s/%s", chap.BaseURL, chap.Chapter.Hash, pageName)

		// Clean and 0-pad each page
		fname := fmt.Sprintf("%s/%07s", dirName, pageNameCleaner.ReplaceAllString(pageName, "${1}${2}"))

		f, err := os.Create(fname)
		if err != nil {
			log.Fatalf("%s: Failed to create file %s", err, fname)
		}

		<-limiter

		dlPage(pageURL, f)
	}

	return store.UpdateChapterDownloaded(c)
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

// Retrieve and parse the metadata for this given series from the series' ID.
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

// Create and store a new manga.
// This does not do anything with feeds or getting the chapters,
// it only gets the series info.
func NewManga(meta MangaMeta, title string, abbrev string, store *SQLite) {
	tags := parseTags(&meta, store)
	var demo string
	if meta.Data.Attributes.PublicationDemographic == "" {
		demo = "Unknown"
	} else {
		demo = meta.Data.Attributes.PublicationDemographic
	}

	m := Manga{
		MangaID:      meta.Data.ID,
		SerTitle:     abbrev,
		FullTitle:    title,
		Descr:        meta.Data.Attributes.Description.En,
		TimeModified: time.Unix(0, 0),
		Tags:         tags,
		Chapters:     []Chapter{},
		Demographic:  goodUpper(demo),
		PubStatus:    goodUpper(meta.Data.Attributes.Status),
	}

	store.insertManga(m)
}

func goodUpper(text string) string {
	r, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToUpper(r)) + text[size:]
}

// Parse the given tags, guarantee they are in the DB,
// then return them in the Tag struct
func parseTags(meta *MangaMeta, store *SQLite) []Tag {
	tagNames := make([]string, 0)

	for _, v := range meta.Data.Attributes.Tags {
		t := v.Attributes.Name.En

		// This is probably not necessary, but I'm paranoid now
		if utf8.RuneCountInString(t) > 0 {
			tagNames = append(tagNames, t)
		}
	}

	store.insertTags(tagNames)
	return store.tagNamesToTags(tagNames)
}

// Pull the feed, add the chapters to the DB
// Returns the updated Manga
func RefreshFeed(manga Manga, store *SQLite) Manga {
	// Implementation note: Right now, this function only gets new chapters.
	// However, it may be useful later to rework it to get everything every time, which would
	// automatically update when MD sorts or updates old chapters.
	offset := 0
	feed := pullFeedMeta(manga.MangaID, offset, time.Unix(0, 0))

	chapters := make([]Chapter, 0)

	for ok := true; ok; ok = feed.Offset < feed.Total {
		pageChapters := parseChData(feed.Data, manga.MangaID)
		chapters = append(chapters, pageChapters...)
		offset += 50
		fmt.Println(pageChapters)
		feed = pullFeedMeta(manga.MangaID, offset, time.Unix(0, 0))
	}

	manga.Chapters = chapters
	store.insertChapters(chapters)
	manga = store.UpdateAtime(manga)
	return manga
}

// Handles all the ugly stuff of parsing the chapters from the API response
func parseChData(data []FeedChData, mangaID string) []Chapter {
	chapters := make([]Chapter, 0)
	var err error
	for _, d := range data {
		var title string
		if d.Attributes.Title == "" {
			title = fmt.Sprintf("Ch. %s", d.Attributes.Chapter)
		} else {
			title = d.Attributes.Title
		}

		var chNum float64
		if d.Attributes.Chapter == "" {
			chNum = 0
		} else {
			chNum, err = strconv.ParseFloat(d.Attributes.Chapter, 64)
			if err != nil {
				log.Fatalf("%s: Failed to parse float from %s", err, d.Attributes.Chapter)
			}
		}

		var vol int
		if d.Attributes.Volume == "" {
			vol = 0
		} else {
			v, err := strconv.ParseInt(d.Attributes.Volume, 10, 32)
			if err != nil {
				log.Fatalf("%s: Failed to parse int from %s", err, d.Attributes.Volume)
			}
			vol = int(v)
		}

		c := Chapter{
			ChapterHash: d.ID,
			ChapterNum:  chNum,
			ChapterName: title,
			VolumeNum:   vol,
			MangaID:     mangaID,
			Downloaded:  false,
			IsRead:      false,
		}
		chapters = append(chapters, c)
	}

	return chapters
}

// Pull and decode the feed for a series
func pullFeedMeta(mangaID string, offset int, lastUpdated time.Time) SeriesFeed {
	feedURL := fmt.Sprintf("https://api.mangadex.org/manga/%s/feed", mangaID)
	params := url.Values{}
	params.Add("translatedLanguage[]", "en")
	params.Add("includeExternalUrl", "0")
	params.Add("offset", fmt.Sprint(offset))
	params.Add("publishAtSince", lastUpdated.Format("2006-01-02T15:04:05"))
	params.Add("limit", "50")
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

func DownloadAll(chapters []Chapter, store *SQLite) []Chapter {
	for i, c := range chapters {
		if !c.Downloaded {
			chapters[i] = dlChapter(c, store)
		}
	}

	return chapters
}
