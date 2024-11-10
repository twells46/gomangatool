PRAGMA foreign_keys = ON;
CREATE TABLE Manga(
    MangaID VARCHAR(64) PRIMARY KEY,
    SerTitle VARCHAR(32) NOT NULL UNIQUE,
    FullTitle VARCHAR(128) NOT NULL,
    Descr VARCHAR(1024),
    TimeModified DATETIME,
    Demographic VARCHAR(7),
    PubStatus VARCHAR(9),

    CHECK (Demographic IN ('Shounen', 'Shoujo', 'Seinen', 'Josei')),
    CHECK (PubStatus IN ('Ongoing', 'Completed', 'Hiatus', 'Cancelled'))
);

CREATE TABLE Tag (
    TagID INTEGER PRIMARY KEY,
    TagTitle VARCHAR(16) UNIQUE
);

CREATE INDEX TagTitle_idx on Tag(TagTitle);

-- Using the extra table for tagging doesn't save enough space to be worthwhile
-- However, it does enable me to easily correlate a tag name with a title in the db,
-- which I assume is better optimized than it would be in code
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

--CREATE INDEX TagManga_idx on ItemTag(MangaID);
--CREATE INDEX TagID_idx on ItemTag(TagID);

CREATE TABLE Chapter (
    ChapterHash VARCHAR(64) PRIMARY KEY,
    ChapterNum REAL,
    ChapterName VARCHAR(32),
    MangaID VARCHAR(64),
    Downloaded INTEGER NOT NULL,
    IsRead INTEGER NOT NULL,

    FOREIGN KEY (MangaID) REFERENCES Manga(MangaID)
);

CREATE INDEX ChapterMid_idx on Chapter(MangaID);

INSERT INTO Tag (TagTitle) VALUES ('Romance');
INSERT INTO Tag (TagTitle) VALUES ('Harem');
INSERT INTO Manga VALUES('ee51d8fb-ba27-46a5-b204-d565ea1b11aa', 'kokuhaku_sarete', 'Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii', 'The servant of the Tendou family, Eito, serves the perfect young lady, Hoshine. One day, Eito informs Hoshine that someone from another class confessed their feelings for him. Hoshine, who has hidden her feelings for Eito since childhood, begins to feel uneasy: «I''ve loved Eito for a much longer time!» As a result, Hoshine starts to approach Eito even more, making bolder advances than ever before! She gets close to him in crowded trains and begs him to sleep by her side… While Eito tries to maintain his role as a formal servant, Hoshine increasingly pursues him with her advances. It''s an adorable romantic comedy between a mistress and her servant, featuring a young lady who strives to win her servant''s love!', 0, 'Shounen', 'Ongoing');
INSERT INTO Manga VALUES('6941f16b-b56e-404a-b4ba-2fc7e009d38f', 'assistant_for_life', 'full_title', 'desc', 12345, 'Shounen', 'Ongoing');

INSERT INTO Chapter VALUES('598c7824-5822-4ac0-90f5-5439f1f7015e', 1.1, 'Ch. 1.1', 'ee51d8fb-ba27-46a5-b204-d565ea1b11aa', 1, 1);
INSERT INTO Chapter VALUES('36c2be46-87a1-42f0-a0a6-51276706a7e9', 1.2, 'Ch. 1.2', 'ee51d8fb-ba27-46a5-b204-d565ea1b11aa', 1, 0);

INSERT INTO ItemTag VALUES ('ee51d8fb-ba27-46a5-b204-d565ea1b11aa', 1);
INSERT INTO ItemTag VALUES ('6941f16b-b56e-404a-b4ba-2fc7e009d38f', 1);
INSERT INTO ItemTag VALUES ('6941f16b-b56e-404a-b4ba-2fc7e009d38f', 2);

--SELECT * FROM Tag;
--SELECT * FROM Manga;
--SELECT * FROM ItemTag;
--SELECT datetime(123456789, 'unixepoch');
--SELECT datetime((select TimeModified from Manga), 'unixepoch');

-- Get titles matching tags
--SELECT *
--    FROM Manga
--    JOIN ItemTag USING (MangaID)
--    JOIN Tag USING (TagID)
--    WHERE Tag.TagTitle = 'Romance';

-- Get tags by title
--SELECT TagTitle
--    FROM Tag
--    JOIN ItemTag USING (TagID)
--    JOIN Manga USING (MangaID)
--    WHERE Manga.SerTitle = 'assistant_for_life';

--Get unread chapters
--SELECT *
--    FROM Manga
--    JOIN Chapter USING (MangaID)
--    WHERE Chapter.IsRead = 0;
