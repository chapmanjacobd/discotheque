-- SQLite schema for media library
CREATE TABLE IF NOT EXISTS media (
    path TEXT PRIMARY KEY,
    title TEXT,
    duration INTEGER,
    size INTEGER,
    time_created INTEGER,
    time_modified INTEGER,
    time_deleted INTEGER DEFAULT 0,
    time_first_played INTEGER DEFAULT 0,
    time_last_played INTEGER DEFAULT 0,
    play_count INTEGER DEFAULT 0,
    playhead INTEGER DEFAULT 0,

    -- Media type info
    type TEXT,
    width INTEGER,
    height INTEGER,
    fps REAL,

    -- Codec info
    video_codecs TEXT,
    audio_codecs TEXT,
    subtitle_codecs TEXT,
    video_count INTEGER DEFAULT 0,
    audio_count INTEGER DEFAULT 0,
    subtitle_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_time_deleted ON media(time_deleted);
CREATE INDEX IF NOT EXISTS idx_time_last_played ON media(time_last_played);
CREATE INDEX IF NOT EXISTS idx_duration ON media(duration);
CREATE INDEX IF NOT EXISTS idx_size ON media(size);

-- Optional FTS table
CREATE VIRTUAL TABLE IF NOT EXISTS media_fts USING fts5(
    path,
    title,
    content='media',
    content_rowid='rowid'
);

-- Trigger to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS media_ai AFTER INSERT ON media BEGIN
    INSERT INTO media_fts(rowid, path, title, captions)
    VALUES (new.rowid, new.path, new.title, new.captions);
END;

CREATE TRIGGER IF NOT EXISTS media_ad AFTER DELETE ON media BEGIN
    DELETE FROM media_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS media_au AFTER UPDATE ON media BEGIN
    UPDATE media_fts SET path = new.path, title = new.title, captions = new.captions
    WHERE rowid = new.rowid;
END;
