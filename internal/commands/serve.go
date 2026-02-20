package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	Port      int      `short:"p" default:"5555" help:"Port to listen on"`
	PublicDir string   `help:"Override embedded web assets with local directory"`
	Dev       bool     `help:"Enable development mode (auto-reload)"`
}

func (c ServeCmd) IsQueryTrait()    {}
func (c ServeCmd) IsFilterTrait()   {}
func (c ServeCmd) IsSortTrait()     {}
func (c ServeCmd) IsPlaybackTrait() {}

func (c *ServeCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	http.HandleFunc("/api/databases", c.handleDatabases)
	http.HandleFunc("/api/query", c.handleQuery)
	http.HandleFunc("/api/play", c.handlePlay)
	http.HandleFunc("/api/events", c.handleEvents)

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

	fmt.Fprintf(w, "data: ready\n\n")
	flusher.Flush()

	if c.Dev {
		fmt.Fprintf(w, "data: reload\n\n")
		flusher.Flush()
	}

	// Keep connection open until client disconnects
	<-r.Context().Done()
}
