package models

type CatStat struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type RatStat struct {
	Rating int64 `json:"rating"`
	Count  int64 `json:"count"`
}

type GenreStat struct {
	Genre string `json:"genre"`
	Count int64  `json:"count"`
}

type DatabaseInfo struct {
	Databases      []string `json:"databases"`
	Trashcan       bool     `json:"trashcan"`
	GlobalProgress bool     `json:"global_progress"`
	Dev            bool     `json:"dev"`
}

type PlayResponse struct {
	Path string `json:"path"`
}

type DeleteRequest struct {
	Path    string `json:"path"`
	Restore bool   `json:"restore"`
}

type ProgressRequest struct {
	Path      string `json:"path"`
	Playhead  int64  `json:"playhead"`
	Duration  int64  `json:"duration"`
	Completed bool   `json:"completed"`
}

type LsEntry struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Type  string `json:"type"`
}
