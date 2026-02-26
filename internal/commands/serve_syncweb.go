//go:build syncweb

package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chapmanjacobd/discotheque/internal/syncweb"
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

var swInstance *syncweb.Syncweb

func (c *ServeCmd) setupSyncweb() {
	sw, err := syncweb.NewSyncweb(c.SyncwebHome, "disco-syncweb", c.SyncwebPublic_, c.SyncwebPrivate_, "")
	if err != nil {
		slog.Warn("Failed to initialize Syncweb instance", "error", err)
	} else {
		swInstance = sw
		if err := sw.Start(); err != nil {
			slog.Error("Failed to start Syncweb instance", "error", err)
		} else {
			slog.Info("Syncweb instance started", "myID", sw.Node.MyID())
		}
	}
}

func (c *ServeCmd) addSyncwebRoots(resultsMap map[string]LsEntry, counts map[string]int, path string) {
	if swInstance != nil && (path == "/" || path == "") {
		for id, localPath := range swInstance.GetFolders() {
			entryPath := fmt.Sprintf("syncweb://%s/", id)
			resultsMap[entryPath] = LsEntry{
				Name:  id + " (" + filepath.Base(localPath) + ")",
				Path:  entryPath,
				IsDir: true,
			}
			counts[entryPath] = 1000 // High priority for roots
		}
	}
}

func (c *ServeCmd) resolveSyncwebPath(path string) (string, string, error) {
	if swInstance == nil {
		return "", "", fmt.Errorf("syncweb not configured")
	}
	return swInstance.ResolveLocalPath(path)
}

func (c *ServeCmd) serveSyncwebContent(w http.ResponseWriter, r *http.Request, folderID, path, localPath string) {
	if swInstance == nil {
		http.Error(w, "Syncweb not configured", http.StatusInternalServerError)
		return
	}
	slog.Info("Serving remote Syncweb file via block pulling", "path", path)
	rs, err := swInstance.NewReadSeeker(r.Context(), folderID, strings.TrimPrefix(path, "syncweb://"+folderID+"/"))
	if err != nil {
		slog.Error("Failed to create SyncwebReadSeeker", "path", path, "error", err)
		http.Error(w, "Failed to stream remote file", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, filepath.Base(localPath), time.Now(), rs)
}

func (c *ServeCmd) handleSyncwebFolders(w http.ResponseWriter, r *http.Request) {
	if swInstance == nil {
		http.Error(w, "Syncweb not configured", http.StatusServiceUnavailable)
		return
	}

	folders := make([]map[string]string, 0)
	for id, path := range swInstance.GetFolders() {
		folders = append(folders, map[string]string{
			"id":   id,
			"path": path,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}

func (c *ServeCmd) handleSyncwebLs(w http.ResponseWriter, r *http.Request) {
	if swInstance == nil {
		http.Error(w, "Syncweb not configured", http.StatusServiceUnavailable)
		return
	}

	folderID := r.URL.Query().Get("folder")
	prefix := r.URL.Query().Get("prefix")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	seq, cancel := swInstance.Node.App.Internals.AllGlobalFiles(folderID)
	defer cancel()

	resultsMap := make(map[string]LsEntry)
	for meta := range seq {
		name := meta.Name
		if !strings.HasPrefix(name, prefix) || name == prefix {
			continue
		}

		rel := strings.TrimPrefix(name, prefix)
		parts := strings.Split(rel, "/")
		entryName := parts[0]
		isDir := len(parts) > 1

		fullSyncwebPath := fmt.Sprintf("syncweb://%s/%s", folderID, filepath.Join(prefix, entryName))
		if _, ok := resultsMap[fullSyncwebPath]; ok {
			continue
		}

		localPath, _, _ := swInstance.ResolveLocalPath(fullSyncwebPath)
		isLocal := utils.FileExists(localPath)

		entry := LsEntry{
			Name:  entryName,
			Path:  fullSyncwebPath,
			IsDir: isDir,
			Local: isLocal,
		}
		if !isDir {
			entry.Type = utils.DetectMimeType(entryName)
		}
		resultsMap[fullSyncwebPath] = entry
	}

	results := make([]LsEntry, 0, len(resultsMap))
	for _, entry := range resultsMap {
		results = append(results, entry)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDir != results[j].IsDir {
			return results[i].IsDir
		}
		return results[i].Name < results[j].Name
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (c *ServeCmd) handleSyncwebDownload(w http.ResponseWriter, r *http.Request) {
	if swInstance == nil {
		http.Error(w, "Syncweb not configured", http.StatusServiceUnavailable)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	localPath, folderID, err := swInstance.ResolveLocalPath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	relativePath, _ := filepath.Rel(swInstance.GetFolders()[folderID], localPath)

	if err := swInstance.Unignore(folderID, relativePath); err != nil {
		slog.Error("Syncweb download trigger failed", "path", path, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Download triggered")
}
