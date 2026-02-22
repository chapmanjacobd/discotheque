package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	database "github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	"github.com/chapmanjacobd/discotheque/web"
)

type ServeCmd struct {
	models.GlobalFlags
	Databases            []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	Port                 int      `short:"p" default:"5555" help:"Port to listen on"`
	PublicDir            string   `help:"Override embedded web assets with local directory"`
	Dev                  bool     `help:"Enable development mode (auto-reload)"`
	Trashcan             bool     `help:"Enable trash/recycle page and empty bin functionality"`
	GlobalProgress       bool     `help:"Enable server-side playback progress tracking"`
	ApplicationStartTime int64    `kong:"-"`
	thumbnailCache       sync.Map `kong:"-"`
	hasFfmpeg            bool     `kong:"-"`
}

func (c *ServeCmd) IsQueryTrait()    {}
func (c *ServeCmd) IsFilterTrait()   {}
func (c *ServeCmd) IsSortTrait()     {}
func (c *ServeCmd) IsPlaybackTrait() {}

func (c *ServeCmd) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/databases", c.handleDatabases)
	mux.HandleFunc("/api/categories", c.handleCategories)
	mux.HandleFunc("/api/genres", c.handleGenres)
	mux.HandleFunc("/api/ratings", c.handleRatings)
	mux.HandleFunc("/api/query", c.handleQuery)
	mux.HandleFunc("/api/play", c.handlePlay)
	mux.HandleFunc("/api/delete", c.handleDelete)
	mux.HandleFunc("/api/progress", c.handleProgress)
	mux.HandleFunc("/api/rate", c.handleRate)
	mux.HandleFunc("/api/playlists", c.handlePlaylists)
	mux.HandleFunc("/api/playlists/items", c.handlePlaylistItems)
	mux.HandleFunc("/api/events", c.handleEvents)
	mux.HandleFunc("/api/ls", c.handleLs)
	mux.HandleFunc("/api/raw", c.handleRaw)
	mux.HandleFunc("/api/hls/playlist", c.handleHLSPlaylist)
	mux.HandleFunc("/api/hls/segment", c.handleHLSSegment)
	mux.HandleFunc("/api/subtitles", c.handleSubtitles)
	mux.HandleFunc("/api/thumbnail", c.handleThumbnail)
	mux.HandleFunc("/opds", c.handleOPDS)

	if c.Trashcan {
		mux.HandleFunc("/api/trash", c.handleTrash)
		mux.HandleFunc("/api/empty-bin", c.handleEmptyBin)
	}

	// Serve static files
	var handler http.Handler
	if c.PublicDir != "" {
		slog.Info("Serving static files from directory", "dir", c.PublicDir)
		handler = http.FileServer(http.Dir(c.PublicDir))
	} else {
		slog.Info("Serving embedded static files")
		handler = http.FileServer(http.FS(web.FS))
	}
	mux.Handle("/", handler)
	return mux
}

// execDB connects to the database and executes fn. If a corruption error occurs,
// it attempts to repair the database and retries the operation once.
func (c *ServeCmd) execDB(ctx context.Context, dbPath string, fn func(*sql.DB) error) error {
	const maxRetries = 1
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			// Connect error might be corruption too (e.g. invalid header)
			if database.IsCorruptionError(err) && i < maxRetries {
				slog.Warn("Database corruption detected on connect, attempting repair", "db", dbPath)
				if repErr := database.Repair(dbPath); repErr != nil {
					return fmt.Errorf("repair failed: %w (original error: %v)", repErr, err)
				}
				slog.Info("Database repaired, retrying connect", "db", dbPath)
				continue
			}
			return err
		}

		err = fn(sqlDB)
		sqlDB.Close() // Close immediately after use to allow repair tools to lock file if needed

		if err != nil {
			if database.IsCorruptionError(err) && i < maxRetries {
				slog.Warn("Database corruption detected on query, attempting repair", "db", dbPath)
				if repErr := database.Repair(dbPath); repErr != nil {
					slog.Error("Database repair failed", "db", dbPath, "error", repErr)
					return err // Return original error if repair fails
				}
				slog.Info("Database repaired, retrying operation", "db", dbPath)
				continue
			}
			return err
		}
		return nil
	}
	return lastErr
}

func (c *ServeCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	c.ApplicationStartTime = time.Now().UnixNano()

	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			slog.Error("Failed to connect to database on startup", "db", dbPath, "error", err)
			continue
		}
		if err := InitDB(sqlDB); err != nil {
			slog.Error("Failed to initialize database", "db", dbPath, "error", err)
		}
		sqlDB.Close()
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		slog.Warn("ffmpeg not found in PATH, on-the-fly transcoding will be unavailable")
		c.hasFfmpeg = false
	} else {
		c.hasFfmpeg = true
	}

	handler := c.Mux()
	slog.Info("Server starting", "port", c.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), handler)
}

func (c *ServeCmd) handleDatabases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Databases      []string `json:"databases"`
		Trashcan       bool     `json:"trashcan"`
		GlobalProgress bool     `json:"global_progress"`
		Dev            bool     `json:"dev"`
	}{
		Databases:      c.Databases,
		Trashcan:       c.Trashcan,
		GlobalProgress: c.GlobalProgress,
		Dev:            c.Dev,
	}
	json.NewEncoder(w).Encode(resp)
}

func (c *ServeCmd) handleCategories(w http.ResponseWriter, r *http.Request) {
	counts := make(map[string]int64)

	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		stats, err := queries.GetCategoryStats(r.Context())
		sqlDB.Close()
		if err != nil {
			continue
		}

		for _, s := range stats {
			counts[s.Category] = counts[s.Category] + s.Count
		}
	}

	type catStat struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var res []catStat
	res = make([]catStat, 0)
	for k, v := range counts {
		if v > 0 {
			res = append(res, catStat{Category: k, Count: v})
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Count > res[j].Count
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (c *ServeCmd) handleRatings(w http.ResponseWriter, r *http.Request) {
	counts := make(map[int64]int64)

	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		stats, err := queries.GetRatingStats(r.Context())
		sqlDB.Close()
		if err != nil {
			continue
		}

		for _, s := range stats {
			counts[s.Rating] = counts[s.Rating] + s.Count
		}
	}

	type ratStat struct {
		Rating int64 `json:"rating"`
		Count  int64 `json:"count"`
	}
	var res []ratStat
	res = make([]ratStat, 0)
	for k, v := range counts {
		res = append(res, ratStat{Rating: k, Count: v})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Rating > res[j].Rating
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (c *ServeCmd) handleQuery(w http.ResponseWriter, r *http.Request) {
	flags := c.GlobalFlags

	// Override flags from URL params
	q := r.URL.Query()
	if search := q.Get("search"); search != "" {
		flags.Search = strings.Fields(search)
	}
	if category := q.Get("category"); category != "" {
		flags.Category = category
	}
	if genre := q.Get("genre"); genre != "" {
		flags.Genre = genre
	}
	if rating := q.Get("rating"); rating != "" {
		if r, err := strconv.Atoi(rating); err == nil {
			if r == 0 {
				flags.Where = append(flags.Where, "(score IS NULL OR score = 0)")
			} else {
				flags.Where = append(flags.Where, fmt.Sprintf("score = %d", r))
			}
		}
	}
	if sortBy := q.Get("sort"); sortBy != "" {
		flags.SortBy = sortBy
		if sortBy == "random" {
			flags.Random = true
		}
	}
	if reverse := q.Get("reverse"); reverse == "true" {
		flags.Reverse = true
	}
	if limit := q.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			flags.Limit = l
		}
	}
	if offset := q.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			flags.Offset = o
		}
	}
	if minSize := q.Get("min_size"); minSize != "" {
		flags.Size = append(flags.Size, ">"+minSize+"MB")
	}
	if maxSize := q.Get("max_size"); maxSize != "" {
		flags.Size = append(flags.Size, "<"+maxSize+"MB")
	}
	if minDuration := q.Get("min_duration"); minDuration != "" {
		flags.Duration = append(flags.Duration, ">"+minDuration+"min")
	}
	if maxDuration := q.Get("max_duration"); maxDuration != "" {
		flags.Duration = append(flags.Duration, "<"+maxDuration+"min")
	}
	if minScore := q.Get("min_score"); minScore != "" {
		flags.Where = append(flags.Where, "score >= "+minScore)
	}
	if maxScore := q.Get("max_score"); maxScore != "" {
		flags.Where = append(flags.Where, "score <= "+maxScore)
	}
	if all := q.Get("all"); all == "true" {
		flags.All = true
	}
	if video := q.Get("video"); video == "true" {
		flags.VideoOnly = true
	}
	if audio := q.Get("audio"); audio == "true" {
		flags.AudioOnly = true
	}
	if image := q.Get("image"); image == "true" {
		flags.ImageOnly = true
	}
	if text := q.Get("text"); text == "true" {
		flags.TextOnly = true
	}
	if watched := q.Get("watched"); watched == "true" {
		w := true
		flags.Watched = &w
	}
	if unfinished := q.Get("unfinished"); unfinished == "true" {
		flags.Unfinished = true
	}
	if completed := q.Get("completed"); completed == "true" {
		flags.Completed = true
	}

	media, err := query.MediaQuery(context.Background(), c.Databases, flags)
	if err != nil {
		slog.Error("Query failed", "dbs", c.Databases, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.hasFfmpeg {
		for i := range media {
			media[i].Transcode = utils.GetTranscodeStrategy(media[i].Media).NeedsTranscode
		}
	}

	query.SortMedia(media, flags)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(media)
}

func (c *ServeCmd) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.Path, "http") && !utils.FileExists(req.Path) {
		slog.Warn("File not found, marking as deleted in databases", "path", req.Path)
		c.markDeletedInAllDBs(r.Context(), req.Path, true)
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Trigger local playback
	slog.Info("Playing", "path", req.Path)
	cmd := exec.Command("mpv", req.Path)
	// We run it in background and don't wait for it
	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ServeCmd) markDeletedInAllDBs(ctx context.Context, path string, deleted bool) {
	var deleteTime int64 = 0
	if deleted {
		deleteTime = time.Now().Unix()
	}

	for _, dbPath := range c.Databases {
		err := c.execDB(ctx, dbPath, func(sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			return queries.MarkDeleted(ctx, database.MarkDeletedParams{
				Path:        path,
				TimeDeleted: sql.NullInt64{Int64: deleteTime, Valid: deleted},
			})
		})
		if err != nil {
			slog.Error("Failed to mark file as deleted", "db", dbPath, "path", path, "error", err)
		}
	}
}

func (c *ServeCmd) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path    string `json:"path"`
		Restore bool   `json:"restore"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.markDeletedInAllDBs(r.Context(), req.Path, !req.Restore)
	w.WriteHeader(http.StatusOK)
}

func (c *ServeCmd) handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path      string `json:"path"`
		Playhead  int64  `json:"playhead"`
		Duration  int64  `json:"duration"`
		Completed bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	increment := 0
	if req.Completed {
		increment = 1
	}

	for _, dbPath := range c.Databases {
		err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
			// Use raw SQL to update progress to avoid complex sqlc param mapping if not existing
			// We want to increment play_count only once per session ideally, but for now we follow simple logic
			if _, err := sqlDB.ExecContext(r.Context(), `
			UPDATE media
			SET time_last_played = ?,
			    time_first_played = COALESCE(time_first_played, ?),
			    playhead = ?,
			    play_count = COALESCE(play_count, 0) + ?
			WHERE path = ?`,
				now, now, req.Playhead, increment, req.Path); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			slog.Error("Failed to update progress", "db", dbPath, "error", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ServeCmd) handleRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path  string  `json:"path"`
		Score float64 `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, dbPath := range c.Databases {
		err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
			if _, err := sqlDB.ExecContext(r.Context(), "UPDATE media SET score = ? WHERE path = ?", req.Score, req.Path); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			slog.Error("Failed to update rating", "db", dbPath, "error", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ServeCmd) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial comment to establish connection
	fmt.Fprintf(w, ": keep-alive\n\n")
	flusher.Flush()

	if c.Dev {
		fmt.Fprintf(w, "data: %d\n\n", c.ApplicationStartTime)
		flusher.Flush()
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (c *ServeCmd) handleLs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	type LsEntry struct {
		Name  string `json:"name"`
		Path  string `json:"path"`
		IsDir bool   `json:"is_dir"`
		Type  string `json:"type,omitempty"`
		InDB  bool   `json:"in_db"`
	}

	resultsMap := make(map[string]LsEntry)

	for _, dbPath := range c.Databases {
		err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
			// Find immediate files
			// Using LIKE 'path/%' AND path NOT LIKE 'path/%/%'
			// In SQLite, we can check if there's another slash after the path
			rows, err := sqlDB.QueryContext(r.Context(), `
				SELECT path, type FROM media 
				WHERE time_deleted = 0 
				  AND path LIKE ? || '%'
				  AND INSTR(SUBSTR(path, LENGTH(?) + 1), '/') = 0`, path, path)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var p, t sql.NullString
				if err := rows.Scan(&p, &t); err == nil && p.Valid {
					name := filepath.Base(p.String)
					resultsMap[p.String] = LsEntry{
						Name:  name,
						Path:  p.String,
						IsDir: false,
						Type:  t.String,
						InDB:  true,
					}
				}
			}

			// Find immediate subfolders
			// Subfolders are prefixes that have at least one more slash
			rows, err = sqlDB.QueryContext(r.Context(), `
				SELECT DISTINCT 
					SUBSTR(path, 1, LENGTH(?) + INSTR(SUBSTR(path, LENGTH(?) + 1), '/')) as folder
				FROM media 
				WHERE time_deleted = 0 
				  AND path LIKE ? || '%'
				  AND INSTR(SUBSTR(path, LENGTH(?) + 1), '/') > 0`, path, path, path, path)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var f string
				if err := rows.Scan(&f); err == nil && f != "" {
					name := filepath.Base(f)
					resultsMap[f] = LsEntry{
						Name:  name,
						Path:  f,
						IsDir: true,
						InDB:  true,
					}
				}
			}
			return nil
		})
		if err != nil {
			slog.Error("handleLs DB query failed", "db", dbPath, "error", err)
		}
	}

	var results []LsEntry
	for _, entry := range resultsMap {
		results = append(results, entry)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDir != results[j].IsDir {
			return results[i].IsDir
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (c *ServeCmd) handleRaw(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	slog.Debug("handleRaw request", "path", path)

	var m models.Media
	found := false
	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		dbMedia, err := queries.GetMediaByPathExact(r.Context(), path)
		sqlDB.Close()
		if err == nil {
			m = models.FromDB(dbMedia)
			found = true
			break
		}
	}

	if !found {
		slog.Warn("Access denied: file not in database", "path", path)
		http.Error(w, "Access denied: file not in database", http.StatusForbidden)
		return
	}

	if !utils.FileExists(path) {
		slog.Warn("File not found on disk, marking as deleted in databases", "path", path)
		c.markDeletedInAllDBs(r.Context(), path, true)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	strategy := utils.GetTranscodeStrategy(m)
	slog.Debug("handleRaw strategy", "path", path, "needs_transcode", strategy.NeedsTranscode, "vcopy", strategy.VideoCopy, "acopy", strategy.AudioCopy)

	if strategy.NeedsTranscode {
		if c.hasFfmpeg {
			c.handleTranscode(w, r, path, m, strategy)
			return
		} else {
			slog.Error("ffmpeg not found in PATH, skipping transcoding", "path", path)
		}
	}

	// Range requests are handled by ServeFile
	slog.Debug("Serving raw file (no transcode)", "path", path)
	http.ServeFile(w, r, path)
}

func (c *ServeCmd) handleTranscode(w http.ResponseWriter, r *http.Request, path string, m models.Media, strategy utils.TranscodeStrategy) {
	w.Header().Set("Content-Type", strategy.TargetMime)
	w.Header().Set("Accept-Ranges", "bytes")

	// Add flags to help with piped streaming duration and timestamp issues
	var args []string

	args = append(args, "-fflags", "+genpts", "-i", path)

	// If we have duration in metadata, tell ffmpeg so it can write it to headers
	if m.Duration != nil && *m.Duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", *m.Duration))
	}

	if strategy.VideoCopy {
		args = append(args, "-c:v", "copy")
	} else {
		if strategy.TargetMime == "video/mp4" {
			args = append(args, "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-crf", "28")
		} else {
			// WebM
			args = append(args, "-c:v", "libvpx-vp9", "-deadline", "realtime", "-cpu-used", "8", "-crf", "30", "-b:v", "0")
		}
	}

	if strategy.AudioCopy {
		args = append(args, "-c:a", "copy")
	} else {
		if strategy.TargetMime == "video/mp4" {
			args = append(args, "-c:a", "aac", "-b:a", "128k", "-ac", "2")
		} else {
			// WebM supports Opus
			args = append(args, "-c:a", "libopus", "-b:a", "128k", "-ac", "2")
		}
	}

	args = append(args, "-avoid_negative_ts", "make_zero", "-map_metadata", "-1", "-sn")

	if strategy.TargetMime == "video/mp4" {
		// frag_keyframe+empty_moov+default_base_moof+global_sidx is the standard for fragmented streaming
		args = append(args, "-f", "mp4", "-movflags", "frag_keyframe+empty_moov+default_base_moof+global_sidx", "pipe:1")
	} else {
		// Matroska with index space reserved and cluster limits can help browsers determine duration
		args = append(args, "-f", "matroska", "-live", "1", "-reserve_index_space", "1024k", "-cluster_size_limit", "2M", "-cluster_time_limit", "5100", "pipe:1")
	}

	ffmpegArgs := append([]string{"-hide_banner", "-loglevel", "error"}, args...)
	slog.Info("Streaming with transcode", "path", path, "strategy", strategy, "args", strings.Join(ffmpegArgs, " "))

	cmd := exec.CommandContext(r.Context(), "ffmpeg", ffmpegArgs...)
	cmd.Stdout = w

	if err := cmd.Start(); err != nil {
		slog.Error("Failed to start ffmpeg", "path", path, "error", err)
		http.Error(w, "Transcoding failed", http.StatusInternalServerError)
		return
	}

	if err := cmd.Wait(); err != nil {
		if r.Context().Err() == nil {
			slog.Error("ffmpeg failed", "path", path, "error", err)
		} else {
			slog.Debug("ffmpeg finished (client disconnected)", "path", path)
		}
	}
}

func (c *ServeCmd) handleSubtitles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	slog.Debug("handleSubtitles request", "path", path, "index", r.URL.Query().Get("index"))

	// Verify path or siblings
	found := false
	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		_, err = queries.GetMediaByPathExact(r.Context(), path)
		if err == nil {
			found = true
			sqlDB.Close()
			break
		}

		// If path doesn't exist, it might be an external subtitle file next to a media file
		// We'll check if any media in the database shares the same directory and base name
		dir := filepath.Dir(path)
		filename := filepath.Base(path)
		base := strings.TrimSuffix(filename, filepath.Ext(filename))
		// Handle cases like movie.en.srt by stripping one more extension if it exists
		if secondExt := filepath.Ext(base); secondExt != "" {
			base = strings.TrimSuffix(base, secondExt)
		}

		// Simple check: does this directory contain ANY media we know with the same base name?
		mediaInDir, _ := queries.GetMedia(r.Context(), 1000)
		for _, m := range mediaInDir {
			if filepath.Dir(m.Path) == dir {
				mBase := strings.TrimSuffix(filepath.Base(m.Path), filepath.Ext(m.Path))
				if mBase == base {
					found = true
					break
				}
			}
		}
		sqlDB.Close()
		if found {
			break
		}
	}

	if !found {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	streamIndex := r.URL.Query().Get("index")

	// If it's a media container but no index is specified, we should try to find an external sidecar
	if streamIndex == "" && (ext == ".mkv" || ext == ".mp4" || ext == ".m4v" || ext == ".mov" || ext == ".webm") {
		// Try to find a sibling subtitle file
		sidecars := utils.GetExternalSubtitles(path)
		if len(sidecars) > 0 {
			// Serve the first found sidecar
			path = sidecars[0]
			ext = strings.ToLower(filepath.Ext(path))
			slog.Debug("Found sidecar for media file", "media", r.URL.Query().Get("path"), "sidecar", path)
		} else {
			http.Error(w, "No index specified and no sidecar found", http.StatusNotFound)
			return
		}
	}

	if ext == ".idx" {
		subPath := strings.TrimSuffix(path, ".idx") + ".sub"
		if !utils.FileExists(subPath) {
			slog.Warn("VobSub conversion requested but .sub file is missing", "idx", path)
			http.Error(w, "Corresponding .sub file not found", http.StatusNotFound)
			return
		}
	}

	if ext == ".vtt" {
		w.Header().Set("Content-Type", "text/vtt")
		http.ServeFile(w, r, path)
		return
	}

	// Convert to VTT using ffmpeg
	w.Header().Set("Content-Type", "text/vtt")

	var args []string
	isImageSub := ext == ".idx" || ext == ".sub" || ext == ".sup"

	if streamIndex != "" {
		// Embedded tracks
		args = append(args, "-i", path, "-map", "0:s:"+streamIndex, "-f", "webvtt", "pipe:1")
	} else {
		// Standalone file (srt, lrc, ass, etc.)
		args = append(args, "-i", path, "-f", "webvtt", "pipe:1")
	}

	ffmpegArgs := append([]string{"-hide_banner", "-loglevel", "error"}, args...)
	slog.Debug("subtitle ffmpeg command", "args", strings.Join(ffmpegArgs, " "))

	cmd := exec.CommandContext(r.Context(), "ffmpeg", ffmpegArgs...)
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		if r.Context().Err() == nil {
			msg := "Failed to convert subtitles"
			if isImageSub || streamIndex != "" {
				msg = "Failed to convert subtitles (image-based formats require OCR which is not yet supported for direct VTT streaming)"
			}
			slog.Error(msg, "path", path, "error", err)
		} else {
			slog.Debug("Subtitle conversion interrupted (client disconnect)", "path", path)
		}
	}
}

func (c *ServeCmd) handleThumbnail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	// Verify path exists in database to prevent arbitrary file access
	found := false
	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		_, err = queries.GetMediaByPathExact(r.Context(), path)
		sqlDB.Close()
		if err == nil {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check cache
	if data, ok := c.thumbnailCache.Load(path); ok {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Write(data.([]byte))
		return
	}

	// Generate thumbnail
	mime := utils.DetectMimeType(path)
	var args []string

	if strings.HasPrefix(mime, "video/") {
		args = []string{"-ss", "25", "-i", path, "-frames:v", "1", "-q:v", "4", "-vf", "scale=320:-1", "-f", "image2", "pipe:1"}
	} else if strings.HasPrefix(mime, "image/") {
		args = []string{"-i", path, "-vf", "scale=320:-1", "-f", "image2", "pipe:1"}
	} else if strings.HasPrefix(mime, "audio/") {
		args = []string{"-i", path, "-an", "-vcodec", "copy", "-f", "image2", "pipe:1"}
	} else {
		http.Error(w, "Unsupported type", http.StatusUnsupportedMediaType)
		return
	}

	cmd := exec.CommandContext(r.Context(), "ffmpeg", append([]string{"-hide_banner", "-loglevel", "error"}, args...)...)
	thumb, err := cmd.Output()
	if err != nil {
		// slog.Debug("Thumbnail generation failed", "path", path, "error", err)
		http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	// Cache it
	c.thumbnailCache.Store(path, thumb)

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(thumb)
}

func (c *ServeCmd) handleTrash(w http.ResponseWriter, r *http.Request) {
	flags := c.GlobalFlags
	flags.OnlyDeleted = true
	flags.HideDeleted = false
	flags.All = true

	media, err := query.MediaQuery(context.Background(), c.Databases, flags)
	if err != nil {
		slog.Error("Trash query failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}

func (c *ServeCmd) handleEmptyBin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flags := c.GlobalFlags
	flags.OnlyDeleted = true
	flags.HideDeleted = false
	flags.All = true

	media, err := query.MediaQuery(context.Background(), c.Databases, flags)
	if err != nil {
		slog.Error("Trash query failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count := 0
	for _, m := range media {
		if utils.FileExists(m.Path) {
			if err := os.Remove(m.Path); err != nil {
				slog.Error("Failed to delete file", "path", m.Path, "error", err)
				continue
			}
		}

		// Remove from DB
		for _, dbPath := range c.Databases {
			err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
				result, err := sqlDB.Exec("DELETE FROM media WHERE path = ?", m.Path)
				if err != nil {
					return err
				}
				rows, _ := result.RowsAffected()
				if rows > 0 {
					count++
				}
				return nil
			})
			if err != nil {
				slog.Error("Failed to delete from DB", "db", dbPath, "error", err)
			}
		}
	}

	slog.Info("Bin emptied", "files_removed", count)
	fmt.Fprintf(w, "Deleted %d files", count)
}

func (c *ServeCmd) handleOPDS(w http.ResponseWriter, r *http.Request) {
	flags := c.GlobalFlags
	flags.TextOnly = true
	flags.All = true

	media, err := query.MediaQuery(r.Context(), c.Databases, flags)
	if err != nil {
		slog.Error("OPDS query failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml;charset=utf-8")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:opds="http://opds-spec.org/2010/catalog">
  <id>discotheque-text</id>
  <title>Discotheque Text</title>
  <updated>`+time.Now().Format(time.RFC3339)+`</updated>
  <author><name>Discotheque</name></author>
`)

	host := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	for _, m := range media {
		title := m.Stem()
		if m.Title != nil && *m.Title != "" {
			title = *m.Title
		}

		author := "Unknown"
		if m.Artist != nil && *m.Artist != "" {
			author = *m.Artist
		}

		mime := "application/octet-stream"
		if m.Type != nil {
			mime = *m.Type
		}

		fmt.Fprintf(w, `
  <entry>
    <title>%s</title>
    <id>%s</id>
    <updated>%s</updated>
    <author><name>%s</name></author>
    <content type="text">%s</content>
    <link rel="http://opds-spec.org/acquisition" href="%s://%s/api/raw?path=%s" type="%s"/>
  </entry>`,
			utils.EscapeXML(title),
			utils.EscapeXML(m.Path),
			time.Now().Format(time.RFC3339), // Ideally use modification time
			utils.EscapeXML(author),
			utils.EscapeXML(m.Path),
			scheme, host, strings.ReplaceAll(url.QueryEscape(m.Path), "+", "%20"),
			mime,
		)
	}

	fmt.Fprint(w, "\n</feed>")
}

func (c *ServeCmd) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		allPlaylists := make([]models.Playlist, 0)
		for _, dbPath := range c.Databases {
			err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
				queries := database.New(sqlDB)
				pls, err := queries.GetPlaylists(r.Context())
				if err != nil {
					return err
				}
				for _, p := range pls {
					allPlaylists = append(allPlaylists, models.PlaylistFromDB(p, dbPath))
				}
				return nil
			})
			if err != nil {
				slog.Error("Failed to fetch playlists", "db", dbPath, "error", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allPlaylists)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Title string `json:"title"`
			Path  string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		playlistPath := req.Path
		if playlistPath == "" {
			playlistPath = "custom:" + utils.RandomString(12)
		}

		dbPath := c.Databases[0]
		var id int64
		err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			var err error
			id, err = queries.InsertPlaylist(r.Context(), database.InsertPlaylistParams{
				Title: sql.NullString{String: req.Title, Valid: true},
				Path:  sql.NullString{String: playlistPath, Valid: true},
			})
			return err
		})
		if err != nil {
			slog.Error("Failed to insert playlist", "title", req.Title, "path", playlistPath, "error", err)
			http.Error(w, fmt.Sprintf("Failed to insert playlist: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id, "db": dbPath})
		return
	}

	if r.Method == http.MethodDelete {
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		dbPath := r.URL.Query().Get("db")
		if dbPath == "" {
			dbPath = c.Databases[0]
		}

		err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			return queries.DeletePlaylist(r.Context(), database.DeletePlaylistParams{
				ID:          id,
				TimeDeleted: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
			})
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (c *ServeCmd) handlePlaylistItems(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		dbPath := r.URL.Query().Get("db")
		var items []database.GetPlaylistItemsRow

		err := c.execDB(r.Context(), dbPath, func(sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			var err error
			items, err = queries.GetPlaylistItems(r.Context(), id)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		media := make([]models.MediaWithDB, 0)
		for _, item := range items {
			m := models.FromDB(database.Media{
				Path:            item.Path,
				Title:           item.Title,
				Duration:        item.Duration,
				Size:            item.Size,
				TimeCreated:     item.TimeCreated,
				TimeModified:    item.TimeModified,
				TimeDeleted:     item.TimeDeleted,
				TimeFirstPlayed: item.TimeFirstPlayed,
				TimeLastPlayed:  item.TimeLastPlayed,
				PlayCount:       item.PlayCount,
				Playhead:        item.Playhead,
				Type:            item.Type,
				Width:           item.Width,
				Height:          item.Height,
				Fps:             item.Fps,
				VideoCodecs:     item.VideoCodecs,
				AudioCodecs:     item.AudioCodecs,
				SubtitleCodecs:  item.SubtitleCodecs,
				VideoCount:      item.VideoCount,
				AudioCount:      item.AudioCount,
				SubtitleCount:   item.SubtitleCount,
				Album:           item.Album,
				Artist:          item.Artist,
				Genre:           item.Genre,
				Mood:            item.Mood,
				Bpm:             item.Bpm,
				Key:             item.Key,
				Decade:          item.Decade,
				Categories:      item.Categories,
				City:            item.City,
				Country:         item.Country,
				Description:     item.Description,
				Language:        item.Language,
				Webpath:         item.Webpath,
				Uploader:        item.Uploader,
				TimeUploaded:    item.TimeUploaded,
				TimeDownloaded:  item.TimeDownloaded,
				ViewCount:       item.ViewCount,
				NumComments:     item.NumComments,
				FavoriteCount:   item.FavoriteCount,
				Score:           item.Score,
				UpvoteRatio:     item.UpvoteRatio,
				Latitude:        item.Latitude,
				Longitude:       item.Longitude,
			})
			m.TrackNumber = models.NullInt64Ptr(item.TrackNumber)
			mw := models.MediaWithDB{
				Media: m,
				DB:    dbPath,
			}
			if c.hasFfmpeg {
				mw.Transcode = utils.GetTranscodeStrategy(m).NeedsTranscode
			}
			media = append(media, mw)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(media)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			PlaylistID  int64  `json:"playlist_id"`
			DB          string `json:"db"`
			MediaPath   string `json:"media_path"`
			TrackNumber int64  `json:"track_number"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := c.execDB(r.Context(), req.DB, func(sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			return queries.AddPlaylistItem(r.Context(), database.AddPlaylistItemParams{
				PlaylistID:  req.PlaylistID,
				MediaPath:   req.MediaPath,
				TrackNumber: sql.NullInt64{Int64: req.TrackNumber, Valid: req.TrackNumber != 0},
			})
		})
		if err != nil {
			slog.Error("Failed to add playlist item", "playlist_id", req.PlaylistID, "media_path", req.MediaPath, "error", err)
			http.Error(w, fmt.Sprintf("Failed to add playlist item: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodDelete {
		var req struct {
			PlaylistID int64  `json:"playlist_id"`
			DB         string `json:"db"`
			MediaPath  string `json:"media_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := c.execDB(r.Context(), req.DB, func(sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			return queries.RemovePlaylistItem(r.Context(), database.RemovePlaylistItemParams{
				PlaylistID: req.PlaylistID,
				MediaPath:  req.MediaPath,
			})
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (c *ServeCmd) handleGenres(w http.ResponseWriter, r *http.Request) {
	counts := make(map[string]int64)

	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		stats, err := queries.GetGenreStats(r.Context())
		sqlDB.Close()
		if err != nil {
			continue
		}

		for _, s := range stats {
			if s.Genre.Valid {
				counts[s.Genre.String] = counts[s.Genre.String] + s.Count
			}
		}
	}

	type genreStat struct {
		Genre string `json:"genre"`
		Count int64  `json:"count"`
	}
	res := make([]genreStat, 0)
	for k, v := range counts {
		if v > 0 {
			res = append(res, genreStat{Genre: k, Count: v})
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Count > res[j].Count
	})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(res)
}

const HLS_SEGMENT_DURATION = 6

func (c *ServeCmd) handleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	// Fetch media to get duration
	var m models.Media
	found := false
	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		dbMedia, err := queries.GetMediaByPathExact(r.Context(), path)
		sqlDB.Close()
		if err == nil {
			m = models.FromDB(dbMedia)
			found = true
			break
		}
	}

	if !found || m.Duration == nil {
		http.Error(w, "Media not found or no duration", http.StatusNotFound)
		return
	}

	duration := float64(*m.Duration)
	segments := int(math.Ceil(duration / HLS_SEGMENT_DURATION))

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")

	fmt.Fprintf(w, "#EXTM3U\n")
	fmt.Fprintf(w, "#EXT-X-VERSION:3\n")
	fmt.Fprintf(w, "#EXT-X-TARGETDURATION:%d\n", HLS_SEGMENT_DURATION)
	fmt.Fprintf(w, "#EXT-X-MEDIA-SEQUENCE:0\n")
	fmt.Fprintf(w, "#EXT-X-PLAYLIST-TYPE:VOD\n")

	for i := range segments {
		segDuration := float64(HLS_SEGMENT_DURATION)
		if i == segments-1 {
			rem := math.Mod(duration, HLS_SEGMENT_DURATION)
			if rem > 0 {
				segDuration = rem
			}
		}
		fmt.Fprintf(w, "#EXTINF:%f,\n", segDuration)
		fmt.Fprintf(w, "/api/hls/segment?path=%s&index=%d\n", url.QueryEscape(path), i)
	}

	fmt.Fprintf(w, "#EXT-X-ENDLIST\n")
}

func (c *ServeCmd) handleHLSSegment(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	indexStr := r.URL.Query().Get("index")
	if path == "" || indexStr == "" {
		http.Error(w, "Path and index required", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	startTime := float64(index * HLS_SEGMENT_DURATION)

	// Check if we have ffmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		http.Error(w, "ffmpeg not found", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "video/MP2T")

	vcopy := r.URL.Query().Get("vcopy") == "true"
	slog.Debug("HLS Segment request", "index", index, "start", startTime, "vcopy", vcopy, "path", path)

	args := []string{
		"-ss", fmt.Sprintf("%f", startTime),
		"-i", path,
		"-t", fmt.Sprintf("%d", HLS_SEGMENT_DURATION),
	}

	if vcopy {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args,
			"-vf", "scale=-2:720", // Downscale to 720p for performance/bandwidth
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-pix_fmt", "yuv420p",
		)
	}

	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		"-f", "mpegts",
		"-output_ts_offset", fmt.Sprintf("%f", startTime), // Align timestamps
		"pipe:1",
	)

	// Skip logging for segments to avoid spam
	// slog.Debug("HLS Segment", "index", index, "start", startTime)

	cmd := exec.CommandContext(r.Context(), "ffmpeg", append([]string{"-hide_banner", "-loglevel", "error"}, args...)...)
	cmd.Stdout = w

	if err := cmd.Run(); err != nil {
		if r.Context().Err() != nil {
			slog.Debug("Client disconnected during HLS transcoding", "path", path, "index", index)
		} else {
			slog.Error("HLS transcoding failed", "path", path, "index", index, "error", err)
		}
	}
}
