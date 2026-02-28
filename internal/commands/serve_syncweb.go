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
	sw, err := syncweb.NewSyncweb(c.SyncwebHome, "disco-syncweb", "")
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
		for _, id := range swInstance.GetFolders() {
			entryPath := fmt.Sprintf("syncweb://%s/", id)
			name := id
			if localPath, ok := swInstance.GetFolderPath(id); ok {
				name += " (" + filepath.Base(localPath) + ")"
			}
			resultsMap[entryPath] = LsEntry{
				Name:  name,
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

	folderIDs := swInstance.GetFolders()
	folders := make([]map[string]string, 0, len(folderIDs))
	for _, id := range folderIDs {
		folders = append(folders, map[string]string{
			"id": id,
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

	// Security check: ensure the folder is one we actually have
	configuredFolders := swInstance.GetFolders()
	found := false
	for _, id := range configuredFolders {
		if id == folderID {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Folder not found or not configured", http.StatusNotFound)
		return
	}

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

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}

	// Try to decode from JSON body first
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Fallback to query param for compatibility if body is empty or malformed
		req.Path = r.URL.Query().Get("path")
	}

	if req.Path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	localPath, folderID, err := swInstance.ResolveLocalPath(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if c.isPathBlacklisted(localPath) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	folderPath, ok := swInstance.GetFolderPath(folderID)
	if !ok {
		http.Error(w, "Folder root not found", http.StatusInternalServerError)
		return
	}
	relativePath, _ := filepath.Rel(folderPath, localPath)

	if err := swInstance.Unignore(folderID, relativePath); err != nil {
		slog.Error("Syncweb download trigger failed", "path", req.Path, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Download triggered")
}
