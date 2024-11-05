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
	ChapterName string
	MangaID     string
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

func (r *SQLite) GetTags(MangaID string) []Tag {
	query := `
	SELECT TagID, TagTitle
	FROM Tag
	JOIN ItemTag USING (TagID)
	JOIN Manga USING (MangaID)
	WHERE Manga.MangaID = ?`
	rows, err := r.db.Query(query, MangaID)
	if err != nil {
		log.Fatalf("%s: Failed to query db", err)
	}
	defer rows.Close()

	var all []Tag
	for rows.Next() {
		var t Tag
		err := rows.Scan(&t.TagID, &t.TagTitle)
		if err != nil {
			log.Fatalf("%s: Failed to query db", err)
		}
		all = append(all, t)
	}

	return all
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
	defer rows.Close()

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
		m.Tags = r.GetTags(m.MangaID)
		all = append(all, m)
	}

	return all, nil
}

func (r *SQLite) Initdb() error {
	create_query := `
	PRAGMA foreign_keys = ON;
	CREATE TABLE Manga(
	    MangaID VARCHAR(64) PRIMARY KEY,
	    SerTitle VARCHAR(32) NOT NULL UNIQUE,
	    FullTitle VARCHAR(128) NOT NULL,
	    Descr VARCHAR(1024),
	    TimeModified DATETIME
	);

	CREATE TABLE Tag (
	    TagID INTEGER PRIMARY KEY,
	    TagTitle VARCHAR(16)
	);

	CREATE TABLE ItemTag (
	    MangaID VARCHAR(64),
	    TagID INTEGER,

	    FOREIGN KEY (MangaID) REFERENCES Manga(MangaID)
	        ON UPDATE CASCADE
	        ON DELETE CASCADE,
	    FOREIGN KEY (TagID) REFERENCES Tag(TagID)
	        ON UPDATE CASCADE
	        ON DELETE CASCADE
	);

	CREATE TABLE Chapter (
	    ChapterHash VARCHAR(64) PRIMARY KEY,
	    ChapterNum REAL,
	    ChapterName VARCHAR(32),
	    MangaID VARCHAR(64),
	    Downloaded INTEGER NOT NULL,
	    IsRead INTEGER NOT NULL,

	    FOREIGN KEY (MangaID) REFERENCES Manga(MangaID)
	);`
	_, err := r.db.Exec(create_query)

	return err
}

func (r *SQLite) InsertChapters(cs []Chapter) {
	tx, err := r.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO Chapter values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, c := range cs {
		_, err = stmt.Exec(c.ChapterHash, c.ChapterNum, c.ChapterName, c.MangaID, c.Download, c.IsRead)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (r *SQLite) InsertManga(m Manga) {
	insertStmt := `INSERT INTO Manga values (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(insertStmt, m.MangaID, m.SerTitle, m.FullTitle, m.Descr, m.TimeModified)
	if err != nil {
		log.Fatal(err)
	}

	r.InsertChapters(m.Chapters)
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

	store.Initdb()

	test1 := Chapter{
		ChapterHash: "598c7824-5822-4ac0-90f5-5439f1f7015e",
		ChapterNum:  1.1,
		ChapterName: "Chapter 1.1",
		MangaID:     "ee51d8fb-ba27-46a5-b204-d565ea1b11aa",
		Download:    true,
		IsRead:      true,
	}
	test2 := Chapter{
		ChapterHash: "36c2be46-87a1-42f0-a0a6-51276706a7e9",
		ChapterNum:  1.2,
		ChapterName: "Chapter 1.2",
		MangaID:     "ee51d8fb-ba27-46a5-b204-d565ea1b11aa",
		Download:    true,
		IsRead:      false,
	}
	test := Manga{
		MangaID:   "ee51d8fb-ba27-46a5-b204-d565ea1b11aa",
		SerTitle:  "kokuhaku_sarete",
		FullTitle: "Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii",
		Descr: `The servant of the Tendou family, Eito, serves the perfect young lady, Hoshine.

			One day, Eito informs Hoshine that someone from another class confessed their feelings for him. Hoshine, who has hidden her feelings for Eito since childhood, begins to feel uneasy:

			"I’ve loved Eito for a much longer time!"

			As a result, Hoshine starts to approach Eito even more, making bolder advances than ever before! She gets close to him in crowded trains and begs him to sleep by her side… While Eito tries to maintain his role as a formal servant, Hoshine increasingly pursues him with her advances.

			It’s an adorable romantic comedy between a mistress and her servant, featuring a young lady who strives to win her servant’s love!
			`,
		TimeModified: time.Unix(0, 0),
		Tags:         []Tag{},
		Chapters:     []Chapter{test1, test2},
	}

	store.InsertManga(test)

	abb, err := store.GetAll()
	if err != nil {
		log.Fatalf("%s: Failed to query db", err)
	}
	fmt.Println(abb)
}
