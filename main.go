package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Tag struct {
	TagID    int
	TagTitle string
}

type Chapter struct {
	ChapterHash string
	ChapterNum  float64
	MangaID     string
	ChapterName string
	Download    bool
	IsRead      bool
}

type Manga struct {
	MangaID      string
	SerTitle     string
	FullTitle    string
	Descr        string
	TimeModified time.Time
	Tags         []Tag
	Chapters     []Chapter
}

type SQLite struct {
	db *sql.DB
}

type Database interface {
	Init() error
	Create(manga Manga) error
	GetAll() ([]Manga, error)
	Get(id string) (*Manga, error)
}

func NewDb(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

func (r *SQLite) GetChapters(MangaID string) []Chapter {
	query := `
	SELECT *
	FROM Chapter
	WHERE MangaID = ?`
	rows, err := r.db.Query(query, MangaID)
	if err != nil {
		log.Fatalf("%s: Failed to query db", err)
	}
	rows.Close()

	var all []Chapter
	for rows.Next() {
		var c Chapter
		err := rows.Scan(&c.ChapterHash, &c.ChapterNum, &c.MangaID, &c.ChapterName, &c.Download, &c.IsRead)
		if err != nil {
			log.Fatalf("%s: Failed to query db", err)
		}
		all = append(all, c)
	}

	return all
}

func (r *SQLite) GetAll() ([]Manga, error) {
	rows, err := r.db.Query("SELECT * FROM Manga")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Manga

	for rows.Next() {
		var m Manga
		err := rows.Scan(&m.MangaID, &m.SerTitle, &m.FullTitle, &m.Descr, &m.TimeModified)
		if err != nil {
			return nil, err
		}
		m.Chapters = r.GetChapters(m.MangaID)
		all = append(all, m)
	}

	return all, nil
}

func main() {
	/*
		md.DlChapter(`362936f9-2456-4120-9bea-b247df21d0bc`)
		feed := md.GetFeed(`6941f16b-b56e-404a-b4ba-2fc7e009d38f`, 0)
		fmt.Println(feed)
	*/
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := NewDb(db)

	test, err := store.GetAll()
	if err != nil {
		log.Fatalf("%s: Failed to query db", err)
	}

	fmt.Println(test)
}
