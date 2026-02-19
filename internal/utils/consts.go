package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	DefaultTableLimit         = 350
	DefaultPlayQueue          = 120
	DefaultSubtitleMix        = 0.35
	DefaultFileRowsReadLimit  = 500000
	DefaultMultiplePlayback   = -1
	DefaultOpenLimit          = 7
	DefaultMpvSocket          = "/tmp/mpv_socket"
)

func GetMpvWatchLaterDir() string {
	home, _ := os.UserHomeDir()
	if IsWindows {
		return filepath.Join(home, "AppData", "Roaming", "mpv", "watch_later")
	}
	if IsMac {
		return filepath.Join(home, "Library", "Application Support", "mpv", "watch_later")
	}
	return filepath.Join(home, ".config", "mpv", "watch_later")
}

var (
	ApplicationStart = time.Now().Unix()
	IsWindows        = runtime.GOOS == "windows"
	IsLinux          = runtime.GOOS == "linux"
	IsMac            = runtime.GOOS == "darwin"
	TERMINAL_SIZE    = struct{ columns, rows int }{80, 24}
)

var SQLiteExtensions = []string{".sqlite", ".sqlite3", ".db", ".db3", ".s3db", ".sl3"}

var AudioExtensions = []string{
	"mka", "opus", "oga", "ogg", "mp3", "mpga", "m2a", "m4a", "m4r", "caf", "m4b", "flac", "wav", "pcm", "aif", "aiff", "wma", "aac", "aa3", "ac3", "ape", "dsf", "dff",
}

var VideoExtensions = []string{
	"avi", "flv", "m4v", "mkv", "mov", "mp4", "mpeg", "mpg", "ogm", "ogv", "ts", "webm", "wmv", "3gp", "3g2",
}

var ImageExtensions = []string{
	"ai", "avif", "bmp", "gif", "heic", "heif", "ico", "jpeg", "jpg", "png", "psd", "svg", "tif", "tiff", "webp",
}

var SubtitleExtensions = []string{
	"srt", "vtt", "mks", "ass", "ssa",
}

var ArchiveExtensions = []string{
	"7z", "bz2", "gz", "rar", "tar", "xz", "zip",
}

func GetTempDir() string {
	return os.TempDir()
}

func GetConfigDir() string {
	home, _ := os.UserHomeDir()
	if IsWindows {
		return filepath.Join(home, "AppData", "Roaming", "disco")
	}
	return filepath.Join(home, ".config", "disco")
}
