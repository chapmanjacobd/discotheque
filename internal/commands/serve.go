package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
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
}

func (c *ServeCmd) IsQueryTrait()    {}
func (c *ServeCmd) IsFilterTrait()   {}
func (c *ServeCmd) IsSortTrait()     {}
func (c *ServeCmd) IsPlaybackTrait() {}

func (c *ServeCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	c.ApplicationStartTime = time.Now().UnixNano()

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		slog.Warn("ffmpeg not found in PATH, on-the-fly transcoding will be unavailable")
	}

	http.HandleFunc("/api/databases", c.handleDatabases)
	http.HandleFunc("/api/categories", c.handleCategories)
	http.HandleFunc("/api/ratings", c.handleRatings)
	http.HandleFunc("/api/query", c.handleQuery)
	http.HandleFunc("/api/play", c.handlePlay)
	http.HandleFunc("/api/delete", c.handleDelete)
	http.HandleFunc("/api/progress", c.handleProgress)
	http.HandleFunc("/api/rate", c.handleRate)
	http.HandleFunc("/api/events", c.handleEvents)
	http.HandleFunc("/api/raw", c.handleRaw)
	http.HandleFunc("/api/subtitles", c.handleSubtitles)
	http.HandleFunc("/api/thumbnail", c.handleThumbnail)
	http.HandleFunc("/opds", c.handleOPDS)

	if c.Trashcan {
		http.HandleFunc("/api/trash", c.handleTrash)
		http.HandleFunc("/api/empty-bin", c.handleEmptyBin)
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
	http.Handle("/", handler)

	slog.Info("Server starting", "port", c.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
}

func (c *ServeCmd) handleDatabases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Databases      []string `json:"databases"`
		Trashcan       bool     `json:"trashcan"`
		GlobalProgress bool     `json:"global_progress"`
	}{
		Databases:      c.Databases,
		Trashcan:       c.Trashcan,
		GlobalProgress: c.GlobalProgress,
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
	}
	if reverse := q.Get("reverse"); reverse == "true" {
		flags.Reverse = true
	}
	if limit := q.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			flags.Limit = l
		}
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

	media, err := query.MediaQuery(context.Background(), c.Databases, flags)
	if err != nil {
		slog.Error("Query failed", "dbs", c.Databases, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query.SortMedia(media, flags)

	w.Header().Set("Content-Type", "application/json")
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
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			slog.Error("Failed to connect to database", "db", dbPath, "error", err)
			continue
		}
		queries := database.New(sqlDB)
		err = queries.MarkDeleted(ctx, database.MarkDeletedParams{
			Path:        path,
			TimeDeleted: sql.NullInt64{Int64: deleteTime, Valid: deleted},
		})
		sqlDB.Close()
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
		Path     string `json:"path"`
		Playhead int64  `json:"playhead"`
		Duration int64  `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}

		// Use raw SQL to update progress to avoid complex sqlc param mapping if not existing
		// We want to increment play_count only once per session ideally, but for now we follow simple logic
		if _, err := sqlDB.ExecContext(r.Context(), `
			UPDATE media
			SET time_last_played = ?,
			    time_first_played = COALESCE(time_first_played, ?),
			    playhead = ?
			WHERE path = ?`,
			now, now, req.Playhead, req.Path); err != nil {
			slog.Error("Failed to update progress", "db", dbPath, "error", err)
		}

		sqlDB.Close()
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
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		if _, err := sqlDB.ExecContext(r.Context(), "UPDATE media SET score = ? WHERE path = ?", req.Score, req.Path); err != nil {
			slog.Error("Failed to update rating", "db", dbPath, "error", err)
		}
		sqlDB.Close()
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

	if c.Dev {
		fmt.Fprintf(w, "data: %d\n\n", c.ApplicationStartTime)
		flusher.Flush()
	}

	// Keep connection open until client disconnects
	<-r.Context().Done()
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

	strategy := c.getTranscodeStrategy(m)
	slog.Info("Transcode strategy determined", "path", path, "needsTranscode", strategy.needsTranscode, "videoCopy", strategy.videoCopy, "audioCopy", strategy.audioCopy, "targetMime", strategy.targetMime)

	if strategy.needsTranscode {
		if _, err := exec.LookPath("ffmpeg"); err == nil {
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

type transcodeStrategy struct {
	needsTranscode bool
	videoCopy      bool
	audioCopy      bool
	targetMime     string
}

func (c *ServeCmd) getTranscodeStrategy(m models.Media) transcodeStrategy {
	ext := strings.ToLower(filepath.Ext(m.Path))

	// If it's a known non-media format, don't even try
	if ext == ".sqlite" || ext == ".db" || ext == ".txt" {
		return transcodeStrategy{needsTranscode: false}
	}

	isSupportedVideoCodec := func(codec string) bool {
		codec = strings.ToLower(codec)
		return strings.Contains(codec, "h264") || strings.Contains(codec, "avc1") || strings.Contains(codec, "vp8") || strings.Contains(codec, "vp9") || strings.Contains(codec, "av1")
	}

	isSupportedAudioCodec := func(codec string) bool {
		if codec == "" {
			return false
		}
		codec = strings.ToLower(codec)
		// If it contains any incompatible codec, return false
		incompatible := []string{"eac3", "ac3", "dts", "truehd", "mlp"}
		for _, inc := range incompatible {
			if strings.Contains(codec, inc) {
				return false
			}
		}

		// It must contain at least one supported codec
		supported := []string{"aac", "mp3", "opus", "vorbis", "flac", "pcm", "wav"}
		for _, sup := range supported {
			if strings.Contains(codec, sup) {
				return true
			}
		}
		return false
	}

	vCodecs := ""
	if m.VideoCodecs != nil {
		vCodecs = *m.VideoCodecs
	}
	aCodecs := ""
	if m.AudioCodecs != nil {
		aCodecs = *m.AudioCodecs
	}
	sCodecs := ""
	if m.SubtitleCodecs != nil {
		sCodecs = *m.SubtitleCodecs
	}

	mime := ""
	if m.Type != nil && *m.Type != "" {
		mime = *m.Type
	} else {
		mime = utils.DetectMimeType(m.Path)
	}

	slog.Debug("Analyzing codecs for transcode", "path", m.Path, "vCodecs", vCodecs, "aCodecs", aCodecs, "sCodecs", sCodecs, "mime", mime, "ext", ext)

	if strings.HasPrefix(mime, "image") {
		return transcodeStrategy{needsTranscode: false}
	}

	if strings.HasPrefix(mime, "video") {
		vNeeds := !isSupportedVideoCodec(vCodecs)
		aNeeds := !isSupportedAudioCodec(aCodecs)

		// Prefer WebM for VP9/VP8/AV1/Opus/Vorbis
		preferWebm := strings.Contains(strings.ToLower(vCodecs), "vp9") || strings.Contains(strings.ToLower(vCodecs), "vp8") || strings.Contains(strings.ToLower(vCodecs), "av1") ||
			strings.Contains(strings.ToLower(aCodecs), "opus") || strings.Contains(strings.ToLower(aCodecs), "vorbis")

		targetMime := "video/mp4"
		if preferWebm {
			targetMime = "video/webm"
		}

		// Check if container already matches the target mime type
		containerMatches := false
		if targetMime == "video/mp4" {
			// Most browsers support H264/AAC in MKV or MOV as well, but we'll be slightly conservative
			if ext == ".mp4" || ext == ".m4v" || ext == ".mov" || ext == ".mkv" {
				containerMatches = true
			}
		} else if targetMime == "video/webm" {
			if ext == ".webm" || ext == ".mkv" {
				containerMatches = true
			}
		}

		slog.Debug("Transcode decision details", "vNeeds", vNeeds, "aNeeds", aNeeds, "preferWebm", preferWebm, "containerMatches", containerMatches, "targetMime", targetMime)

		if vNeeds || aNeeds || !containerMatches {
			return transcodeStrategy{
				needsTranscode: true,
				videoCopy:      !vNeeds,
				audioCopy:      !aNeeds,
				targetMime:     targetMime,
			}
		}
	} else if strings.HasPrefix(mime, "audio") {
		if !isSupportedAudioCodec(aCodecs) || (ext != ".mp3" && ext != ".m4a" && ext != ".ogg" && ext != ".flac" && ext != ".wav" && ext != ".opus") {
			return transcodeStrategy{
				needsTranscode: true,
				audioCopy:      isSupportedAudioCodec(aCodecs),
				targetMime:     "audio/mpeg",
			}
		}
	}

	return transcodeStrategy{needsTranscode: false}
}

func (c *ServeCmd) handleTranscode(w http.ResponseWriter, r *http.Request, path string, m models.Media, strategy transcodeStrategy) {
	w.Header().Set("Content-Type", strategy.targetMime)
	w.Header().Set("Accept-Ranges", "bytes")

	// Add flags to help with piped streaming duration and timestamp issues
	var args []string

	args = append(args, "-fflags", "+genpts", "-i", path)

	// If we have duration in metadata, tell ffmpeg so it can write it to headers
	if m.Duration != nil && *m.Duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", *m.Duration))
	}

	if strategy.videoCopy {
		args = append(args, "-c:v", "copy")
	} else {
		if strategy.targetMime == "video/mp4" {
			args = append(args, "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-crf", "28")
		} else {
			// WebM
			args = append(args, "-c:v", "libvpx-vp9", "-deadline", "realtime", "-cpu-used", "8", "-crf", "30", "-b:v", "0")
		}
	}

	if strategy.audioCopy {
		args = append(args, "-c:a", "copy")
	} else {
		if strategy.targetMime == "video/mp4" {
			args = append(args, "-c:a", "aac", "-b:a", "128k", "-ac", "2")
		} else {
			// WebM supports Opus
			args = append(args, "-c:a", "libopus", "-b:a", "128k", "-ac", "2")
		}
	}

	args = append(args, "-avoid_negative_ts", "make_zero", "-map_metadata", "-1", "-sn")

	if strategy.targetMime == "video/mp4" {
		// frag_keyframe+empty_moov+default_base_moof+global_sidx is the standard for fragmented streaming
		args = append(args, "-f", "mp4", "-movflags", "frag_keyframe+empty_moov+default_base_moof+global_sidx", "pipe:1")
	} else {
		// Matroska with index space reserved and cluster limits can help browsers determine duration
		args = append(args, "-f", "matroska", "-live", "1", "-reserve_index_space", "1024k", "-cluster_size_limit", "2M", "-cluster_time_limit", "5100", "pipe:1")
	}

	slog.Info("Streaming with transcode/remux", "path", path, "strategy", strategy)
	ffmpegArgs := append([]string{"-hide_banner", "-loglevel", "error"}, args...)
	slog.Debug("ffmpeg command", "args", strings.Join(ffmpegArgs, " "))

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
		args = []string{"-i", path, "-map", "0:s:" + streamIndex, "-f", "webvtt", "pipe:1"}
	} else {
		// Standalone file (srt, lrc, ass, etc.)
		args = []string{"-i", path, "-f", "webvtt", "pipe:1"}
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
			sqlDB, err := database.Connect(dbPath)
			if err != nil {
				continue
			}
			_, err = sqlDB.Exec("DELETE FROM media WHERE path = ?", m.Path)
			sqlDB.Close()
			if err == nil {
				count++
				break
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
