package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Manga struct {
	MangaID      string
	SerTitle     string
	FullTitle    string
	Descr        string
	ChaptersRead float64
	TimeModified time.Time
}

// TODO: CRUD

type SQLite struct {
	db *sql.DB
}

func NewDb(db *sql.DB) *SQLite {
	return &SQLite{db}
}

type Database interface {
	Init() error
	CreateRecord(manga Manga) error
	GetAll() ([]Manga, error)
	Get() (Manga, error)
}

func (s *SQLite) Init() error {
	createQuery := `
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS Manga(
    MangaID VARCHAR(64) PRIMARY KEY,
    SerTitle VARCHAR(32) NOT NULL UNIQUE,
    FullTitle VARCHAR(128) NOT NULL,
    Descr VARCHAR(1024),
    ChaptersRead REAL,
    TimeModified INTEGER
);

CREATE TABLE IF NOT EXISTS Tag (
    TagID INTEGER PRIMARY KEY,
    TagTitle VARCHAR(16)
);

CREATE TABLE IF NOT EXISTS ItemTag (
    MangaID VARCHAR(64),
    TagID INTEGER,

    FOREIGN KEY (MangaID) 
        REFERENCES Manga(MangaID)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    FOREIGN KEY (TagID)
        REFERENCES Tag(TagID)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);`
	_, err := s.db.Exec(createQuery)

	return err
}

func (s *SQLite) CreateRecord(m Manga) error {
	insQuery := "INSERT INTO Manga(MangaID, SerTitle, FullTitle, Descr, ChaptersRead, TimeModified) values (?, ?, ?, ?, ?, ?)"
	_, err := s.db.Exec(insQuery,
		m.MangaID, m.SerTitle, m.FullTitle, m.Descr, m.ChaptersRead, m.TimeModified.Unix())

	return err
}

func main() {
	//md.DlChapter(`362936f9-2456-4120-9bea-b247df21d0bc`)
	//feed := md.GetFeed(`6941f16b-b56e-404a-b4ba-2fc7e009d38f`, 0)
	//fmt.Println(feed)
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatalf("%s: Failed to open SQL database", err)
	}
	defer db.Close()

	test := Manga{"ee51d8fb-ba27-46a5-b204-d565ea1b11aa", "kokuhaku_sarete", "asdf", "The servant of the Tendou family, Eito, serves the perfect young lady, Hoshine. One day, Eito informs Hoshine that someone from another class confessed their feelings for him. Hoshine, who has hidden her feelings for Eito since childhood, begins to feel uneasy: «I''ve loved Eito for a much longer time!» As a result, Hoshine starts to approach Eito even more, making bolder advances than ever before! She gets close to him in crowded trains and begs him to sleep by her side… While Eito tries to maintain his role as a formal servant, Hoshine increasingly pursues him with her advances. It''s an adorable romantic comedy between a mistress and her servant, featuring a young lady who strives to win her servant''s love!", 0, time.Unix(0, 0)}
	store := NewDb(db)
	if err = store.Init(); err != nil {
		log.Fatalf("%s: Failed to init db", err)
	}
	if err = store.CreateRecord(test); err != nil {
		log.Fatalf("%s: Failed to create record", err)
	}
	fmt.Println("asdf")
}
