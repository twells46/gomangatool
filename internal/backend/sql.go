package backend

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
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
	Demographic  string
	PubStatus    string
}

// Store the sql connection.
// Need a custom struct to define custom methods
type SQLite struct {
	db *sql.DB
}

// Return a new SQLite connection
func NewDb(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

// Given a slice of tag names, retrieve the ID and return a slice of Tag structs.
func (r *SQLite) tagNamesToTags(names []string) []Tag {
	tags := make([]Tag, 0)
	stmt, _ := r.db.Prepare("SELECT TagID, TagTitle FROM Tag WHERE TagTitle = ?")

	for _, v := range names {
		row := stmt.QueryRow(v)
		//row := r.db.QueryRow("SELECT TagID, TagTitle FROM Tag WHERE TagTitle = ?", v)
		var t Tag
		if err := row.Scan(&t.TagID, &t.TagTitle); err != nil {
			log.Fatalf("%s: Failed to get tags", err)
		}
		tags = append(tags, t)
	}

	return tags
}

// Get all the tags for a given manga.
// Works on an ID because it is part of the process of
// constructing a new Manga struct, so it makes no sense
// for it to take the struct.
func (r *SQLite) getTags(MangaID string) []Tag {
	query := `
	SELECT TagID, TagTitle
	FROM Tag
	JOIN ItemTag USING (TagID)
	JOIN Manga USING (MangaID)
	WHERE Manga.MangaID = ?`
	rows, err := r.db.Query(query, MangaID)
	if err != nil {
		log.Fatalf("%s: Failed to query get tags", err)
	}
	defer rows.Close()

	var all []Tag
	for rows.Next() {
		var t Tag
		err := rows.Scan(&t.TagID, &t.TagTitle)
		if err != nil {
			log.Fatalf("%s: Failed to parse tags", err)
		}
		all = append(all, t)
	}

	return all
}

// Get all the chapters for a given manga.
// Takes ID instead of the full struct for the same reason as getTags.
func (r *SQLite) getChapters(MangaID string) []Chapter {
	query := `
	SELECT *
	FROM Chapter
	WHERE MangaID = ?`
	rows, err := r.db.Query(query, MangaID)
	if err != nil {
		log.Fatalf("%s: Failed to query db for chapters", err)
	}
	defer rows.Close()

	var all []Chapter
	for rows.Next() {
		var c Chapter
		err := rows.Scan(&c.ChapterHash, &c.ChapterNum, &c.ChapterName, &c.MangaID, &c.Download, &c.IsRead)
		if err != nil {
			log.Fatalf("%s: Failed to parse chapter", err)
		}
		all = append(all, c)
	}

	return all
}

// Get all the Manga from the DB, complete with tags and chapters.
func (r *SQLite) GetAll() []Manga {
	rows, err := r.db.Query("SELECT * FROM Manga")
	if err != nil {
		log.Fatalf("%s: Failed to query db for manga", err)
	}
	defer rows.Close()

	var all []Manga

	for rows.Next() {
		var m Manga
		err := rows.Scan(&m.MangaID, &m.SerTitle, &m.FullTitle, &m.Descr, &m.TimeModified, &m.Demographic, &m.PubStatus)
		if err != nil {
			log.Fatalf("%s: Failed to parse manga", err)
		}
		m.Chapters = r.getChapters(m.MangaID)
		m.Tags = r.getTags(m.MangaID)
		all = append(all, m)
	}

	return all
}

// Initialize the database
func (r *SQLite) initdb() {
	create_stmt := `
	PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS Manga(
	    MangaID VARCHAR(64) PRIMARY KEY,
	    SerTitle VARCHAR(32) NOT NULL UNIQUE,
	    FullTitle VARCHAR(128) NOT NULL,
	    Descr VARCHAR(1024),
	    TimeModified DATETIME,
	    Demographic VARCHAR(7),
	    PubStatus VARCHAR(9),

	    CHECK (Demographic IN ('Shounen', 'Shoujo', 'Seinen', 'Jousei')),
	    CHECK (PubStatus IN ('Ongoing', 'Completed', 'Hiatus', 'Cancelled'))
    );

	CREATE TABLE IF NOT EXISTS Tag (
	    TagID INTEGER PRIMARY KEY,
	    TagTitle VARCHAR(16) UNIQUE
	);
	CREATE UNIQUE INDEX IF NOT EXISTS TagTitle_idx on Tag(TagTitle);

	CREATE TABLE IF NOT EXISTS ItemTag (
	    MangaID VARCHAR(64),
	    TagID INTEGER,

	    FOREIGN KEY (MangaID) REFERENCES Manga(MangaID)
	        ON UPDATE CASCADE
	        ON DELETE CASCADE,
	    FOREIGN KEY (TagID) REFERENCES Tag(TagID)
	        ON UPDATE CASCADE
	        ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS Chapter (
	    ChapterHash VARCHAR(64) PRIMARY KEY,
	    ChapterNum REAL,
	    ChapterName VARCHAR(32),
	    MangaID VARCHAR(64),
	    Downloaded INTEGER NOT NULL,
	    IsRead INTEGER NOT NULL,

	    FOREIGN KEY (MangaID) REFERENCES Manga(MangaID)
	);
	CREATE INDEX IF NOT EXISTS ChapterMid_idx on Chapter(MangaID);`

	_, err := r.db.Exec(create_stmt)
	if err != nil {
		log.Fatalf("%s: Failed to initialize DB", err)
	}
}

// Given a slice of tag names, add the tags to the DB if they don't already exist.
func (r *SQLite) insertTags(names []string) {
	tx, err := r.db.Begin()
	if err != nil {
		log.Fatalf("%s: Failed to begin transaction", err)
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO Tag (TagTitle) values (?)")
	if err != nil {
		log.Fatalf("%s: Failed to prepare transaction", err)
	}
	defer stmt.Close()

	for _, name := range names {
		_, err = stmt.Exec(name)
		if err != nil {
			log.Fatalf("%s: Failed to execute transaction on %s", err, name)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("%s: Failed to commit transaction", err)
	}
}

// Link the specified Tags with their Manga in the DB.
func (r *SQLite) linkTags(MangaID string, tags []Tag) {
	tx, err := r.db.Begin()
	if err != nil {
		log.Fatalf("%s: Failed to begin transaction", err)
	}
	stmt, err := tx.Prepare("INSERT INTO ItemTag values (?, ?)")
	if err != nil {
		log.Fatalf("%s: Failed to prepare transaction", err)
	}
	defer stmt.Close()

	for _, t := range tags {
		_, err = stmt.Exec(MangaID, t.TagID)
		if err != nil {
			log.Fatalf("%s: Failed to execute transaction on %v", err, t)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("%s: Failed to commit transaction", err)
	}
}

// Insert the given chapters to the DB.
func (r *SQLite) insertChapters(chapters []Chapter) {
	tx, err := r.db.Begin()
	if err != nil {
		log.Fatalf("%s: Failed to begin transaction", err)
	}
	stmt, err := tx.Prepare("INSERT INTO Chapter values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("%s: Failed to prepare transaction", err)
	}
	defer stmt.Close()

	for _, c := range chapters {
		_, err = stmt.Exec(c.ChapterHash, c.ChapterNum, c.ChapterName, c.MangaID, c.Download, c.IsRead)
		if err != nil {
			log.Fatalf("%s: Failed to execute transaction on %v", err, c)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("%s: Failed to commit transaction", err)
	}
}

// Insert the given Manga into the DB
func (r *SQLite) insertManga(m Manga) {
	// The Manga table in the DB only has 7 fields, so this is correct.
	// See directly below for where we insert the tags and chapters.
	insertStmt := `INSERT INTO Manga values (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(insertStmt, m.MangaID, m.SerTitle, m.FullTitle, m.Descr, m.TimeModified, m.Demographic, m.PubStatus)
	if err != nil {
		log.Fatalf("%s: Failed to insert %v", err, m)
	}

	r.insertChapters(m.Chapters)
	r.linkTags(m.MangaID, m.Tags)
}

// Get a new DB connection.
// Guarantees that the file you specify will be created
// the database will be initialized.
func Opendb(name string) *SQLite {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		log.Fatalf("%s: Failed to open %s", err, name)
	}

	store := NewDb(db)
	store.initdb()

	return store
}

func SqlTester() {
	os.Remove("test.db")
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := NewDb(db)

	store.initdb()

	testt1 := Tag{1, "Romance"}
	testt2 := Tag{2, "Harem"}

	store.insertTags([]string{"Romance", "Harem", "Romance"})

	testc1 := Chapter{
		ChapterHash: "598c7824-5822-4ac0-90f5-5439f1f7015e",
		ChapterNum:  1.1,
		ChapterName: "Chapter 1.1",
		MangaID:     "ee51d8fb-ba27-46a5-b204-d565ea1b11aa",
		Download:    true,
		IsRead:      true,
	}
	testc2 := Chapter{
		ChapterHash: "36c2be46-87a1-42f0-a0a6-51276706a7e9",
		ChapterNum:  1.2,
		ChapterName: "Chapter 1.2",
		MangaID:     "ee51d8fb-ba27-46a5-b204-d565ea1b11aa",
		Download:    true,
		IsRead:      false,
	}
	testm1 := Manga{
		MangaID:   "ee51d8fb-ba27-46a5-b204-d565ea1b11aa",
		SerTitle:  "kokuhaku_sarete",
		FullTitle: "Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii",
		Descr: `The servant of the Tendou family, Eito, serves the perfect young lady, Hoshine.

One day, Eito informs Hoshine that someone from another class confessed their feelings for him. Hoshine, who has hidden her feelings for Eito since childhood, begins to feel uneasy:

"I’ve loved Eito for a much longer time!"

As a result, Hoshine starts to approach Eito even more, making bolder advances than ever before! She gets close to him in crowded trains and begs him to sleep by her side… While Eito tries to maintain his role as a formal servant, Hoshine increasingly pursues him with her advances.

It’s an adorable romantic comedy between a mistress and her servant, featuring a young lady who strives to win her servant’s love!`,
		TimeModified: time.Unix(0, 0),
		Tags:         []Tag{testt1},
		Chapters:     []Chapter{testc1, testc2},
		Demographic:  "Shounen",
		PubStatus:    "Ongoing",
	}

	store.insertManga(testm1)

	testc3 := Chapter{
		ChapterHash: "362936f9-2456-4120-9bea-b247df21d0bc",
		ChapterNum:  1,
		ChapterName: "Detective-chan's Assistant",
		MangaID:     "6941f16b-b56e-404a-b4ba-2fc7e009d38f",
		Download:    true,
		IsRead:      true,
	}
	testc4 := Chapter{
		ChapterHash: "ac58155f-0045-41cc-a726-79b5e049e51d",
		ChapterNum:  2,
		ChapterName: "We have to promote the detective club!",
		MangaID:     "6941f16b-b56e-404a-b4ba-2fc7e009d38f",
		Download:    true,
		IsRead:      false,
	}

	testm2 := Manga{
		MangaID:   "6941f16b-b56e-404a-b4ba-2fc7e009d38f",
		SerTitle:  "joushu_de_ite",
		FullTitle: "Isshou Watashi no Joshu de ite!",
		Descr: `"The relationship between a detective and their assistant is not unlike romance!"

Due to his natural tendency to assist others (a "helper type" personality that makes him constantly want to lend a hand), Mitomo Tasuku had always lived in the shadows. To break out of this, he made it his life's mission to "become number one." However, at the high school he transfers to, he encounters three detective girls, who happen to be his rivals from elementary school with whom he formed a detective club together. Thus begins Tasuku's fierce battle for the top detective spot! However... these girls seem to harbor passionate feelings of a different kind...!?`,
		TimeModified: time.Unix(0, 0),
		Tags:         []Tag{testt1, testt2},
		Chapters:     []Chapter{testc3, testc4},
		Demographic:  "Shounen",
		PubStatus:    "Ongoing",
	}

	store.insertManga(testm2)

	abb := store.GetAll()
	fmt.Println(abb)
}
