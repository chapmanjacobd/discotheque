//go:build !syncweb

package commands

import (
	"fmt"
	"net/http"
)

func (c *ServeCmd) setupSyncweb() {
	// No-op
}

func (c *ServeCmd) addSyncwebRoots(resultsMap map[string]LsEntry, counts map[string]int, path string) {
	// No-op
}

func (c *ServeCmd) resolveSyncwebPath(path string) (string, string, error) {
	return "", "", fmt.Errorf("syncweb support not compiled in")
}

func (c *ServeCmd) serveSyncwebContent(w http.ResponseWriter, r *http.Request, folderID, path, localPath string) {
	http.Error(w, "Syncweb support not compiled in", http.StatusNotImplemented)
}

func (c *ServeCmd) handleSyncwebFolders(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Syncweb support not compiled in", http.StatusNotImplemented)
}

func (c *ServeCmd) handleSyncwebLs(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Syncweb support not compiled in", http.StatusNotImplemented)
}

func (c *ServeCmd) handleSyncwebDownload(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Syncweb support not compiled in", http.StatusNotImplemented)
}
