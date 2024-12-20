package backend

import (
	"cmp"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ------- TYPES -------

// A review of a Manga.
type Review struct {
	MangaID string
	Rating  int    // Rating can be whatever you want, but I intended it to be some n/100
	Rev     string // The full text of your review
}

// A Tag can represent a genre or prominent element of a Manga
type Tag struct {
	TagID    int
	TagTitle string
}

func (t Tag) String() string {
	return t.TagTitle
}

// Hold a single chapter of a manga.
type Chapter struct {
	ChapterHash string
	ChapterNum  float64
	ChapterName string
	VolumeNum   int
	MangaID     string
	Downloaded  bool
	IsRead      bool
	ChapterPath string
}

// Implement list.DefaultItem
func (c Chapter) FilterValue() string { return fmt.Sprintf("%.1f %s", c.ChapterNum, c.ChapterName) }

// Implement list.Item
func (c Chapter) Title() string { return fmt.Sprintf("%.1f: %s", c.ChapterNum, c.ChapterName) }
func (c Chapter) Description() string {
	var dl, r string
	if c.Downloaded {
		dl = "Downloaded: ◯︎"
	} else {
		dl = "Downloaded: X"
	}
	if c.IsRead {
		r = "Read: ◯︎"
	} else {
		r = "Read: X"
	}
	return dl + "\t" + r
}

// Function to use with slice.SortFunc.
// Returns a negative number when a < b, a positive number
// when a > b, and 0 when a == b.
func chapterCmp(a, b Chapter) int {
	if volDiff := cmp.Compare(a.VolumeNum, b.VolumeNum); (a.VolumeNum != 0 && b.VolumeNum != 0) && volDiff != 0 {
		return volDiff
	}
	return cmp.Compare(a.ChapterNum, b.ChapterNum)
}

type Manga struct {
	MangaID      string
	SerTitle     string
	FullTitle    string
	Descr        string
	TimeModified time.Time
	Tags         []Tag
	Chapters     []Chapter
	lastVolume   int
	lastChapter  float64
	Demographic  string
	PubStatus    string
	Review       Review
}

// Implement list.DefaultItem
func (m Manga) FilterValue() string { return fmt.Sprintf("%s %s %v", m.FullTitle, m.SerTitle, m.Tags) }

// Implement list.Item
func (m Manga) Title() string       { return fmt.Sprintf("%s (%s)", m.FullTitle, m.SerTitle) }
func (m Manga) Description() string { return m.Descr }

// Store the SQL connection.
type SQLite struct {
	db *sql.DB
}

// Return a new SQLite connection
func newDb(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

// ------- CREATE FUNCTIONS -------

// Given a slice of tag names, add the tags to the DB if they don't already exist.
// Because of the database structure, to associate a Tag with a Manga, use linkTags
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

// Link the specified Tags with a Manga.
// Both the Tags and the Manga referenced must already be in the DB.
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
		log.Fatalf("%s: Failed to begin chapter add transaction", err)
	}
	// Sometimes the API return duplicates
	// Don't know why it does, but just ignore them
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO Chapter values (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("%s: Failed to prepare transaction", err)
	}
	defer stmt.Close()

	for _, c := range chapters {
		_, err = stmt.Exec(
			c.ChapterHash,
			c.ChapterNum,
			c.ChapterName,
			c.VolumeNum,
			c.MangaID,
			c.Downloaded,
			c.IsRead,
			c.ChapterPath)
		if err != nil {
			log.Fatalf("%s: Failed to execute transaction on %v", err, c)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("%s: Failed to commit chapter add transaction", err)
	}
}

// Insert the given Manga into the DB
func (r *SQLite) insertManga(m Manga) {
	insertStmt := `INSERT INTO Manga values (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(insertStmt,
		m.MangaID,
		m.SerTitle,
		m.FullTitle,
		m.Descr,
		m.TimeModified,
		m.lastVolume,
		m.lastChapter,
		m.Demographic,
		m.PubStatus)
	if err != nil {
		log.Fatalf("%s: Failed to insert %v", err, m)
	}

	r.insertChapters(m.Chapters)
	r.linkTags(m.MangaID, m.Tags)

	//log.Printf("Successfully inserted %v", m)
}

// Insert a new review into the DB
func (r *SQLite) insertReview(rev Review) {
	insertStmt := "INSERT INTO Review VALUES (?, ?, ?)"
	_, err := r.db.Exec(insertStmt, rev.MangaID, rev.Rating, rev.Rev)
	if err != nil {
		log.Fatalf("%s: Failed to insert %v", err, rev)
	}
}

// ------- READ FUNCTIONS -------

// Given a slice of tag names, retrieve the ID and return a slice of Tag structs.
// Intended solely for use in parseTags.
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
// This is only indended for use in GetByID and GetAll.
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

	all := make([]Tag, 0)
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
func (r *SQLite) GetChapters(MangaID string) []Chapter {
	query := `
	SELECT *
	FROM Chapter
	WHERE MangaID = ?`
	rows, err := r.db.Query(query, MangaID)
	if err != nil {
		log.Fatalf("%s: Failed to query db for chapters", err)
	}
	defer rows.Close()

	all := make([]Chapter, 0)
	for rows.Next() {
		var c Chapter
		err := rows.Scan(&c.ChapterHash, &c.ChapterNum, &c.ChapterName, &c.VolumeNum, &c.MangaID, &c.Downloaded, &c.IsRead, &c.ChapterPath)
		if err != nil {
			log.Fatalf("%s: Failed to parse chapter", err)
		}
		all = append(all, c)
	}

	return all
}

// Get the review for a given manga
func (r *SQLite) GetReview(MangaID string) Review {
	row := r.db.QueryRow("SELECT * FROM Review WHERE MangaID = ?", MangaID)

	var rev Review
	err := row.Scan(&rev.MangaID, &rev.Rating, &rev.Rev)
	if err == sql.ErrNoRows {
		return Review{}
	} else if err != nil {
		log.Fatalf("%s: Failed to query db for revew for id %s", err, MangaID)
	}

	return rev
}

// Get a single Manga from the DB
func (r *SQLite) GetByID(mangaID string) Manga {
	row := r.db.QueryRow("SELECT * FROM Manga WHERE MangaID = ?", mangaID)
	var m Manga
	err := row.Scan(&m.MangaID, &m.SerTitle, &m.FullTitle, &m.Descr,
		&m.TimeModified, &m.lastVolume, &m.lastChapter, &m.Demographic, &m.PubStatus)
	if err != nil {
		log.Fatalf("%s: Failed to query db for manga", err)
	}
	m.Chapters = r.GetChapters(m.MangaID)
	m.Tags = r.getTags(m.MangaID)
	m.Review = r.GetReview(m.MangaID)

	return m
}

// Get all the Manga from the DB, complete with tags, chapters, and review
func (r *SQLite) GetAll() []Manga {
	rows, err := r.db.Query("SELECT * FROM Manga")
	if err != nil {
		log.Fatalf("%s: Failed to query db for manga", err)
	}
	defer rows.Close()

	all := make([]Manga, 0)

	for rows.Next() {
		var m Manga
		err := rows.Scan(
			&m.MangaID,
			&m.SerTitle,
			&m.FullTitle,
			&m.Descr,
			&m.TimeModified,
			&m.lastVolume,
			&m.lastChapter,
			&m.Demographic,
			&m.PubStatus)
		if err != nil {
			log.Fatalf("%s: Failed to parse manga", err)
		}

		m.Chapters = r.GetChapters(m.MangaID)
		m.Tags = r.getTags(m.MangaID)
		m.Review = r.GetReview(m.MangaID)

		all = append(all, m)
	}

	return all
}

// ------- UPDATE FUNCTIONS -------

// Update the TimeModified for the given Manga in the DB
// and return the updated Manga.
func (r *SQLite) UpdateTimeModified(m Manga) Manga {
	m.TimeModified = time.Now()
	updateStmt := "UPDATE Manga SET TimeModified = ? WHERE MangaID = ?"
	res, err := r.db.Exec(updateStmt, m.TimeModified, m.MangaID)
	if err != nil {
		log.Fatalf("%s: Failed to update access time %v", err, m)
	} else if n, _ := res.RowsAffected(); n > 1 {
		log.Fatalf("Bad UpdateAtime: Updated %d rows", n)
	}

	return m
}

// Update Downloaded for the given Chapter in the DB
// and return the updated Chapter.
func (r *SQLite) UpdateChapterDownloaded(c Chapter) Chapter {
	stmt := "UPDATE Chapter SET Downloaded = 1 WHERE ChapterHash = ?"
	res, err := r.db.Exec(stmt, c.ChapterHash)
	if err != nil {
		log.Fatalf("%s: Failed to update downloaded status %v", err, c)
	} else if n, _ := res.RowsAffected(); n > 1 {
		log.Fatalf("Bad UpdateChapterDownloaded: Updated %d rows", n)
	}

	c.Downloaded = true
	return c
}

// Update IsRead for the given Chapter in the DB.
func (r *SQLite) UpdateChapterRead(c Chapter) {
	stmt := "UPDATE Chapter SET IsRead = 1 WHERE ChapterHash = ?"
	res, err := r.db.Exec(stmt, c.ChapterHash)
	if err != nil {
		log.Fatalf("%s: Failed to update read status %v", err, c)
	} else if n, _ := res.RowsAffected(); n > 1 {
		log.Fatalf("Bad UpdateChapterRead: Updated %d rows", n)
	}
}

// Initialization functions

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
		LastVolume INTEGER,
		LastChapter REAL,
	    Demographic VARCHAR(7),
	    PubStatus VARCHAR(9),

	    CHECK (Demographic IN ('Shounen', 'Shoujo', 'Seinen', 'Josei', 'Unknown')),
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
		VolumeNum INTEGER,
	    MangaID VARCHAR(64),
	    Downloaded INTEGER NOT NULL,
	    IsRead INTEGER NOT NULL,
		ChapterPath VARCHAR(64),

	    FOREIGN KEY (MangaID) REFERENCES Manga(MangaID)
	);
	CREATE INDEX IF NOT EXISTS ChapterMid_idx on Chapter(MangaID);

	CREATE TABLE IF NOT EXISTS Review (
		MangaID VARCHAR(64) PRIMARY KEY,
		Rating INTEGER,
		Rev VARCHAR(5120),

		FOREIGN KEY (MangaID) REFERENCES Manga(MangaID),
		CHECK (
			Rating BETWEEN 0 AND 100
		)
	);`

	_, err := r.db.Exec(create_stmt)
	if err != nil {
		log.Fatalf("%s: Failed to initialize DB", err)
	}
}

// Get a new DB connection.
// Guarantees that the file you specify will be created
// and the tables will be initialized.
func Opendb(name string) *SQLite {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		log.Fatalf("%s: Failed to open %s", err, name)
	}

	store := newDb(db)
	store.initdb()

	return store
}
