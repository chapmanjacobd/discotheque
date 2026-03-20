CREATE TABLE IF NOT EXISTS captions (
    media_path TEXT NOT NULL,
    time REAL,
    text TEXT,
    FOREIGN KEY (media_path) REFERENCES media(path) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS idx_captions_path ON captions(media_path);
