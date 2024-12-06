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
	"slices"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

// This stores the response from a `manga/%s` API query
// to be parsed into more useful forms.
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
			LastVolume             string `json:"lastVolume"`
			LastChapter            string `json:"lastChapter"`
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

// Stores the response from an `at-home/server/%s` API query.
// This is what we use to download a single chapter -
// the final URL we use to download the pages is built
// out of the results from this struct.
type chapterMeta struct {
	Result  string `json:"result"`
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
	} `json:"chapter"`
}

// Stores the list of chapters returned as part of a
// `manga/%s/feed` API query. It exists to allow easier organization
// when parsing the entire feed response.
type feedChData struct {
	ID         string `json:"id"`
	Attributes struct {
		Title   string `json:"title"`
		Volume  string `json:"volume"`
		Chapter string `json:"chapter"`
	} `json:"attributes"`
}

// Stores the response from a `manga/%s/feed` API query.
type SeriesFeed struct {
	Result   string       `json:"result"`
	Response string       `json:"response"`
	Data     []feedChData `json:"data"`
	Limit    int          `json:"limit"`
	Offset   int          `json:"offset"`
	Total    int          `json:"total"`
}

// Download a chapter given the Chapter struct and return the
// updated Chapter.
// NOTE: This also updates the Downloaded status in the DB.
func dlChapter(c Chapter, store *SQLite) Chapter {
	chap := getChapMetadata(c.ChapterHash)

	// The regex takes a page name from the API like this:
	// x6-23b96047cdd7217e5f493894de6d536afa046e7a33695e539a6960e2a7304d35.jpg
	// and uses match groups to create this:
	// 6.jpg
	pageNameCleaner := regexp.MustCompile(`^[A-z]?([0-9]+)-.*(\.[a-z]*)`)

	if err := os.MkdirAll(c.ChapterPath, 0770); err != nil {
		log.Fatalf("%s: Failed to create directory %s", err, c.ChapterHash)
	}

	// Respect API rate limit
	limiter := time.Tick(350 * time.Millisecond)

	for _, pageName := range chap.Chapter.Data {
		pageURL := fmt.Sprintf("%s/data/%s/%s", chap.BaseURL, chap.Chapter.Hash, pageName)

		// Clean and 0-pad each page
		fname := fmt.Sprintf("%s/%07s", c.ChapterPath, pageNameCleaner.ReplaceAllString(pageName, "${1}${2}"))

		f, err := os.Create(fname)
		if err != nil {
			log.Fatalf("%s: Failed to create file %s", err, fname)
		}

		<-limiter

		dlPage(pageURL, f)
	}

	return store.UpdateChapterDownloaded(c)
}

// Pull and decode a single chapter's metadata.
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

// Download a single page.
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

// Create a new Manga, store it in the DB, and return it.
// This does not do anything with feeds or getting the chapters,
// it only gets the series info.
func NewManga(meta MangaMeta, title string, abbrev string, store *SQLite) Manga {
	tags := parseTags(&meta, store)
	var demo string
	if meta.Data.Attributes.PublicationDemographic == "" {
		demo = "Unknown"
	} else {
		demo = meta.Data.Attributes.PublicationDemographic
	}

	var finV int
	if meta.Data.Attributes.LastVolume == "" {
		finV = 0
	} else {
		n, _ := strconv.ParseInt(meta.Data.Attributes.LastVolume, 10, 32)
		finV = int(n)
	}

	var finC float64
	if meta.Data.Attributes.LastChapter == "" {
		finC = 0
	} else {
		i, _ := strconv.ParseFloat(meta.Data.Attributes.LastChapter, 64)
		finC = i
	}

	m := Manga{
		MangaID:      meta.Data.ID,
		SerTitle:     abbrev,
		FullTitle:    title,
		Descr:        meta.Data.Attributes.Description.En,
		TimeModified: time.Unix(0, 0),
		Tags:         tags,
		Chapters:     []Chapter{},
		lastVolume:   finV,
		lastChapter:  finC,
		Demographic:  goodUpper(demo),
		PubStatus:    goodUpper(meta.Data.Attributes.Status),
	}

	store.insertManga(m)
	return m
}

// Parse the given tags, guarantee they are in the DB,
// then return them in the Tag struct.
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

// Helper to uppercase the first letter of a string
func goodUpper(text string) string {
	r, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToUpper(r)) + text[size:]
}

// Pull the MD feed and add the chapters to the DB.
// Returns the updated Manga.
func RefreshFeed(manga Manga, store *SQLite) Manga {
	// Implementation note: Right now, this function only gets new chapters.
	// However, it may be useful later to rework it to get everything every time, which would
	// automatically update when MD sorts or updates old chapters.
	offset := 0
	feed := pullFeedMeta(manga.MangaID, offset, manga.TimeModified)

	chapters := manga.Chapters

	for ok := true; ok; ok = feed.Offset < feed.Total {
		pageChapters := parseChData(feed.Data, manga.MangaID, manga.SerTitle)
		chapters = append(chapters, pageChapters...)
		offset += 50
		feed = pullFeedMeta(manga.MangaID, offset, manga.TimeModified)
	}

	slices.SortFunc(chapters, chapterCmp)
	manga.Chapters = chapters
	store.insertChapters(chapters)
	//manga = store.UpdateTimeModified(manga)
	return manga
}

// Handle all the ugly stuff of parsing the chapters from the API response.
func parseChData(data []feedChData, mangaID string, abbrev string) []Chapter {
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
				n, err := strconv.ParseFloat(d.Attributes.Volume, 64)
				if err != nil {
					v = 0
				} else {
					v = int64(n)
				}
				//log.Fatalf("%s: Failed to parse int from %s", err, d.Attributes.Volume)
			}
			vol = int(v)
		}

		path := fmt.Sprintf("/home/twells/media/manga/%s/%02d/%05.1f-%s", abbrev, vol, chNum, d.ID)

		c := Chapter{
			ChapterHash: d.ID,
			ChapterNum:  chNum,
			ChapterName: title,
			VolumeNum:   vol,
			MangaID:     mangaID,
			Downloaded:  false,
			IsRead:      false,
			ChapterPath: path,
		}
		chapters = append(chapters, c)
	}

	return chapters
}

// Pull and decode the feed for a series.
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

// Downloads the given chapters, returning the updated entries.
// Any chapters with Chapter.Downloaded == true are ignored.
func DownloadChapters(store *SQLite, chapters ...Chapter) []Chapter {
	for i, c := range chapters {
		if !c.Downloaded {
			chapters[i] = dlChapter(c, store)
		}
	}

	return chapters
}
