package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
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
	http.HandleFunc("/api/query", c.handleQuery)
	http.HandleFunc("/api/play", c.handlePlay)
	http.HandleFunc("/api/events", c.handleEvents)
	http.HandleFunc("/api/raw", c.handleRaw)
	http.HandleFunc("/api/subtitles", c.handleSubtitles)
	http.HandleFunc("/api/thumbnail", c.handleThumbnail)

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
	json.NewEncoder(w).Encode(c.Databases)
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
		now := time.Now().Unix()
		for _, dbPath := range c.Databases {
			sqlDB, err := database.Connect(dbPath)
			if err != nil {
				slog.Error("Failed to connect to database", "db", dbPath, "error", err)
				continue
			}
			queries := database.New(sqlDB)
			err = queries.MarkDeleted(r.Context(), database.MarkDeletedParams{
				Path:        req.Path,
				TimeDeleted: sql.NullInt64{Int64: now, Valid: true},
			})
			sqlDB.Close()
			if err != nil {
				slog.Error("Failed to mark file as deleted", "db", dbPath, "path", req.Path, "error", err)
			}
		}

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
		http.Error(w, "Access denied: file not in database", http.StatusForbidden)
		return
	}

	if !utils.FileExists(path) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	strategy := c.getTranscodeStrategy(m)
	if strategy.needsTranscode {
		if _, err := exec.LookPath("ffmpeg"); err == nil {
			c.handleTranscode(w, r, path, strategy)
			return
		}
	}

	// Range requests are handled by ServeFile
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
		codec = strings.ToLower(codec)
		// flac and wav are supported by most modern browsers
		return strings.Contains(codec, "aac") || strings.Contains(codec, "mp3") || strings.Contains(codec, "opus") || strings.Contains(codec, "vorbis") || strings.Contains(codec, "flac") || strings.Contains(codec, "pcm")
	}

	vCodecs := ""
	if m.VideoCodecs != nil {
		vCodecs = *m.VideoCodecs
	}
	aCodecs := ""
	if m.AudioCodecs != nil {
		aCodecs = *m.AudioCodecs
	}

	mime := ""
	if m.Type != nil {
		mime = *m.Type
	}

	if strings.HasPrefix(mime, "image/") {
		return transcodeStrategy{needsTranscode: false}
	}

	if strings.HasPrefix(mime, "video/") {
		vNeeds := !isSupportedVideoCodec(vCodecs)
		aNeeds := !isSupportedAudioCodec(aCodecs)
		// Even if codecs are supported, some containers (like MKV) might need remuxing to MP4 for browser compatibility
		needsRemux := ext != ".mp4" && ext != ".webm" && ext != ".m4v"

		if vNeeds || aNeeds || needsRemux {
			return transcodeStrategy{
				needsTranscode: true,
				videoCopy:      !vNeeds,
				audioCopy:      !aNeeds,
				targetMime:     "video/mp4",
			}
		}
	} else if strings.HasPrefix(mime, "audio/") {
		if !isSupportedAudioCodec(aCodecs) || (ext != ".mp3" && ext != ".m4a" && ext != ".ogg" && ext != ".flac" && ext != ".wav") {
			return transcodeStrategy{
				needsTranscode: true,
				audioCopy:      isSupportedAudioCodec(aCodecs),
				targetMime:     "audio/mpeg",
			}
		}
	}

	return transcodeStrategy{needsTranscode: false}
}

func (c *ServeCmd) handleTranscode(w http.ResponseWriter, r *http.Request, path string, strategy transcodeStrategy) {
	w.Header().Set("Content-Type", strategy.targetMime)

	var args []string
	if strings.HasPrefix(strategy.targetMime, "video/") {
		args = []string{"-i", path}

		if strategy.videoCopy {
			args = append(args, "-c:v", "copy")
		} else {
			args = append(args, "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-crf", "28")
		}

		if strategy.audioCopy {
			args = append(args, "-c:a", "copy")
		} else {
			args = append(args, "-c:a", "aac", "-b:a", "128k")
		}

		args = append(args, "-f", "mp4", "-movflags", "frag_keyframe+empty_moov", "pipe:1")
	} else {
		// Audio only
		args = []string{"-i", path}
		if strategy.audioCopy {
			args = append(args, "-c:a", "copy")
		} else {
			args = append(args, "-c:a", "libmp3lame", "-q:a", "2")
		}
		args = append(args, "-f", "mp3", "pipe:1")
	}

	slog.Info("Streaming with transcode/remux", "path", path, "strategy", strategy)
	cmd := exec.CommandContext(r.Context(), "ffmpeg", append([]string{"-hide_banner", "-loglevel", "error"}, args...)...)
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

	// For subtitles, the path could be the media file (for embedded subs) or a separate srt/vtt file
	// Verification logic...
	found := false
	for _, dbPath := range c.Databases {
		sqlDB, err := database.Connect(dbPath)
		if err != nil {
			continue
		}
		queries := database.New(sqlDB)
		// If it's a direct srt/vtt file, its parent media should be in DB
		// This is a bit simplified; ideally we check if it's a known sibling
		_, err = queries.GetMediaByPathExact(r.Context(), path)
		if err == nil {
			found = true
			sqlDB.Close()
			break
		}
		// Check if it's a sibling of any media
		dir := filepath.Dir(path)
		mediaInDir, _ := queries.GetMedia(r.Context(), 1000) // just a sample
		for _, m := range mediaInDir {
			if filepath.Dir(m.Path) == dir {
				found = true
				break
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
	if ext == ".vtt" {
		w.Header().Set("Content-Type", "text/vtt")
		http.ServeFile(w, r, path)
		return
	}

	// Convert to VTT using ffmpeg
	w.Header().Set("Content-Type", "text/vtt")
	
	streamIndex := r.URL.Query().Get("index")
	var args []string
	if streamIndex != "" {
		// Use -map to select the specific subtitle stream from the media file
		args = []string{"-i", path, "-map", "0:s:" + streamIndex, "-f", "webvtt", "pipe:1"}
	} else {
		args = []string{"-i", path, "-f", "webvtt", "pipe:1"}
	}
	
	cmd := exec.CommandContext(r.Context(), "ffmpeg", append([]string{"-hide_banner", "-loglevel", "error"}, args...)...)
	cmd.Stdout = w
	if err := cmd.Run(); err != nil {
		slog.Error("Failed to convert subtitles", "path", path, "error", err)
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
		slog.Debug("Thumbnail generation failed", "path", path, "error", err)
		http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	// Cache it
	c.thumbnailCache.Store(path, thumb)

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(thumb)
}
