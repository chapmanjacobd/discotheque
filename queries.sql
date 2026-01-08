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
  AND time_last_played > 0
ORDER BY time_last_played DESC
LIMIT ?;

-- name: GetUnwatchedMedia :many
SELECT * FROM media
WHERE time_deleted = 0
  AND time_last_played = 0
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
