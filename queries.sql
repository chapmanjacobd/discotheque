-- name: GetMedia :many
SELECT * FROM media
WHERE time_deleted = 0
ORDER BY path
LIMIT ?;

-- name: GetMediaByType :many
SELECT * FROM media
WHERE time_deleted = 0
  AND (
    (? AND type LIKE 'video/%')
    OR (? AND type LIKE 'audio/%' AND video_count = 0)
    OR (? AND type LIKE 'image/%')
  )
ORDER BY path
LIMIT ?;
-- name: GetMediaBySize :many
SELECT * FROM media
WHERE time_deleted = 0
  AND size >= ?
  AND size <= ?
ORDER BY size DESC
LIMIT ?;

-- name: GetMediaByDuration :many
SELECT * FROM media
WHERE time_deleted = 0
  AND duration >= ?
  AND duration <= ?
ORDER BY duration DESC
LIMIT ?;

-- name: GetMediaByPath :many
SELECT * FROM media
WHERE time_deleted = 0
  AND path LIKE ?
ORDER BY path
LIMIT ?;

-- name: GetWatchedMedia :many
SELECT * FROM media
WHERE time_deleted = 0
  AND COALESCE(time_last_played, 0) > 0
ORDER BY time_last_played DESC
LIMIT ?;

-- name: GetUnwatchedMedia :many
SELECT * FROM media
WHERE time_deleted = 0
  AND COALESCE(time_last_played, 0) = 0
ORDER BY path
LIMIT ?;

-- name: GetUnfinishedMedia :many
SELECT * FROM media
WHERE time_deleted = 0
  AND playhead > 0
  AND playhead < duration * 0.95
ORDER BY time_last_played DESC
LIMIT ?;

-- name: GetMediaByPlayCount :many
SELECT * FROM media
WHERE time_deleted = 0
  AND play_count >= ?
  AND play_count <= ?
ORDER BY play_count DESC
LIMIT ?;

-- name: GetRandomMedia :many
SELECT * FROM media
WHERE time_deleted = 0
ORDER BY RANDOM()
LIMIT ?;

-- name: GetSiblingMedia :many
SELECT * FROM media
WHERE time_deleted = 0
  AND path LIKE ?
  AND path != ?
ORDER BY path
LIMIT ?;

-- name: SearchMediaFTS :many
SELECT m.* FROM media m
JOIN media_fts f ON m.rowid = f.rowid
WHERE f.path MATCH sqlc.arg('query')
  AND m.time_deleted = 0
ORDER BY f.rank
LIMIT sqlc.arg('limit');

-- name: UpdatePlayHistory :exec
UPDATE media
SET time_last_played = ?,
    time_first_played = COALESCE(time_first_played, ?),
    play_count = COALESCE(play_count, 0) + 1,
    playhead = ?
WHERE path = ?;

-- name: MarkDeleted :exec
UPDATE media
SET time_deleted = ?
WHERE path = ?;

-- name: UpdatePath :exec
UPDATE media
SET path = ?
WHERE path = ?;

-- name: UpsertMedia :exec
INSERT INTO media (
    path, title, duration, size, time_created, time_modified,
    type, width, height, fps,
    video_codecs, audio_codecs, subtitle_codecs,
    video_count, audio_count, subtitle_count,
    album, artist, genre, description, language
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?, ?, ?
)
ON CONFLICT(path) DO UPDATE SET
    title = excluded.title,
    duration = excluded.duration,
    size = excluded.size,
    time_modified = excluded.time_modified,
    type = excluded.type,
    width = excluded.width,
    height = excluded.height,
    fps = excluded.fps,
    video_codecs = excluded.video_codecs,
    audio_codecs = excluded.audio_codecs,
    subtitle_codecs = excluded.subtitle_codecs,
    video_count = excluded.video_count,
    audio_count = excluded.audio_count,
    subtitle_count = excluded.subtitle_count,
    album = excluded.album,
    artist = excluded.artist,
    genre = excluded.genre,
    description = excluded.description,
    language = excluded.language;

-- name: InsertPlaylist :one
INSERT INTO playlists (path, extractor_key, extractor_config)
VALUES (?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
    extractor_key = excluded.extractor_key,
    extractor_config = excluded.extractor_config
RETURNING id;

-- name: GetStats :one
SELECT
    COUNT(*) as total_count,
    SUM(size) as total_size,
    SUM(duration) as total_duration,
    COUNT(CASE WHEN COALESCE(time_last_played, 0) > 0 THEN 1 END) as watched_count,
    COUNT(CASE WHEN COALESCE(time_last_played, 0) = 0 THEN 1 END) as unwatched_count
FROM media
WHERE time_deleted = 0;

-- name: GetStatsByType :many
SELECT
    type,
    COUNT(*) as count,
    SUM(size) as total_size,
    SUM(duration) as total_duration
FROM media
WHERE time_deleted = 0
GROUP BY type;
