-- SQLite schema for media library

CREATE TABLE IF NOT EXISTS playlists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE,
    title TEXT,
    extractor_key TEXT,
    extractor_config TEXT,
    time_deleted INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS playlist_items (
    playlist_id INTEGER NOT NULL,
    media_path TEXT NOT NULL,
    track_number INTEGER,
    PRIMARY KEY (playlist_id, media_path),
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    FOREIGN KEY (media_path) REFERENCES media(path) ON DELETE CASCADE
);

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
    subtitle_count INTEGER DEFAULT 0,

    -- Extra metadata
    album TEXT,
    artist TEXT,
    genre TEXT,
    mood TEXT,
    bpm INTEGER,
    key TEXT,
    decade TEXT,
    categories TEXT,
    city TEXT,
    country TEXT,
    description TEXT,
    language TEXT,

    -- Online / Social metadata
    webpath TEXT,
    uploader TEXT,
    time_uploaded INTEGER,
    time_downloaded INTEGER,
    view_count INTEGER,
    num_comments INTEGER,
    favorite_count INTEGER,
    score REAL,
    upvote_ratio REAL,

    -- Location
    latitude REAL,
    longitude REAL
);

CREATE TABLE IF NOT EXISTS captions (
    media_path TEXT NOT NULL,
    time REAL,
    text TEXT,
    FOREIGN KEY (media_path) REFERENCES media(path) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_captions_path ON captions(media_path);

-- FTS for captions
CREATE VIRTUAL TABLE IF NOT EXISTS captions_fts USING fts5(
    media_path UNINDEXED,
    text,
    content='captions'
);

-- Triggers for captions FTS
CREATE TRIGGER IF NOT EXISTS captions_ai AFTER INSERT ON captions BEGIN
    INSERT INTO captions_fts(rowid, media_path, text)
    VALUES (new.rowid, new.media_path, new.text);
END;

CREATE TRIGGER IF NOT EXISTS captions_ad AFTER DELETE ON captions BEGIN
    DELETE FROM captions_fts WHERE rowid = old.rowid;
END;

CREATE TABLE IF NOT EXISTS history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_path TEXT NOT NULL,
    time_played INTEGER DEFAULT (strftime('%s', 'now')),
    playhead INTEGER,
    done INTEGER,
    FOREIGN KEY (media_path) REFERENCES media(path) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_history_path ON history(media_path);
CREATE INDEX IF NOT EXISTS idx_history_time ON history(time_played);

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
    INSERT INTO media_fts(rowid, path, title)
    VALUES (new.rowid, new.path, new.title);
END;

CREATE TRIGGER IF NOT EXISTS media_ad AFTER DELETE ON media BEGIN
    DELETE FROM media_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS media_au AFTER UPDATE ON media BEGIN
    INSERT INTO media_fts(media_fts, rowid, path, title) VALUES('delete', old.rowid, old.path, old.title);
    INSERT INTO media_fts(rowid, path, title) VALUES (new.rowid, new.path, new.title);
END;
