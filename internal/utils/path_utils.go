package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// RandomString returns a random hexadecimal string of the given length
func RandomString(n int) string {
	b := make([]byte, n/2+1)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)[:n]
}

// RandomFilename appends a random string to the filename before the extension
func RandomFilename(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return fmt.Sprintf("%s.%s%s", base, RandomString(6), ext)
}

// TrimPathSegments reduces the length of path segments to fit within desiredLength
func TrimPathSegments(path string, desiredLength int) string {
	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)

	// Split directory into segments
	sep := string(filepath.Separator)
	var segments []string
	if filepath.IsAbs(path) {
		segments = append(segments, sep)
		dir = strings.TrimPrefix(dir, sep)
	}
	
	if dir != "" && dir != "." {
		segments = append(segments, strings.Split(dir, sep)...)
	}
	
	// Add stem as the last segment
	segments = append(segments, stem)

	targetLength := desiredLength - len(ext)

	currentLength := 0
	for _, s := range segments {
		currentLength += len(s)
	}

	for currentLength > targetLength {
		// Find the longest segment, skipping the separator if it's there
		longestIdx := -1
		for i := 0; i < len(segments); i++ {
			if segments[i] == sep {
				continue
			}
			if longestIdx == -1 || len(segments[i]) > len(segments[longestIdx]) {
				longestIdx = i
			}
		}

		if longestIdx == -1 || len(segments[longestIdx]) <= 1 {
			break // Cannot shorten anymore
		}

		segments[longestIdx] = segments[longestIdx][:len(segments[longestIdx])-1]
		currentLength--

		allEven := true
		for _, s := range segments {
			if s == sep {
				continue
			}
			if len(s)%2 != 0 {
				allEven = false
				break
			}
		}
		if allEven {
			for i := range segments {
				if segments[i] == sep {
					continue
				}
				if len(segments[i]) > 0 {
					segments[i] = segments[i][:len(segments[i])-1]
					currentLength--
				}
			}
		}
	}

	// Reconstruct path
	var res string
	if len(segments) > 0 && segments[0] == sep {
		res = sep + filepath.Join(segments[1:]...)
	} else {
		res = filepath.Join(segments...)
	}
	
	return res + ext
}

// SafeJoin joins a base path with a user-provided path, preventing directory traversal
func SafeJoin(base string, userPath string) string {
	// Clean the user path to remove .. and other traversal elements
	userPath = filepath.Clean(userPath)
	
	// Split and filter out traversal elements just in case Clean didn't handle everything as expected for "safe" join
	parts := strings.Split(userPath, string(filepath.Separator))
	var safeParts []string
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			continue
		}
		safeParts = append(safeParts, p)
	}
	
	return filepath.Join(append([]string{base}, safeParts...)...)
}

// Relativize removes leading slashes and drive letters
func Relativize(path string) string {
	// Remove drive letter on Windows
	if len(path) >= 2 && path[1] == ':' {
		path = path[2:]
	}
	
	// Remove leading slashes
	path = strings.TrimLeft(path, `/\`)
	
	return path
}

// StripMountSyntax is a repeated relativize
func StripMountSyntax(path string) string {
	return Relativize(path)
}

// IsEmptyFolder checks if a folder contains no files (recursively)
func IsEmptyFolder(path string) bool {
	empty := true
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil && info.Size() > 0 {
				empty = false
				return filepath.SkipDir // Found a non-empty file, can stop
			}
		}
		return nil
	})
	if err != nil {
		return false
	}
	return empty
}

// FolderSize calculates the total size of all files in a folder (recursively)
func FolderSize(path string) int64 {
	var size int64
	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size
}

// PathTupleFromURL returns (parentDir, filename) from a URL
func PathTupleFromURL(rawURL string) (string, string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", filepath.Base(rawURL)
	}

	path := u.Path
	host := strings.ReplaceAll(u.Host, ":", ".")
	
	if path == "" || path == "/" {
		return host, ""
	}

	filename := filepath.Base(path)
	parent := SafeJoin(host, filepath.Dir(path))
	
	return StripMountSyntax(parent), filename
}

type CleanPathOptions struct {
	MaxNameLen       int
	DotSpace         bool
	CaseInsensitive  bool
	LowercaseFolders bool
	DedupeParts      bool
}

func CleanPath(path string, opts CleanPathOptions) string {
	if opts.MaxNameLen == 0 {
		opts.MaxNameLen = 255
	}

	pre := ""
	sep := string(filepath.Separator)
	if strings.HasPrefix(path, sep) {
		pre = sep
		path = strings.TrimPrefix(path, sep)
	} else if len(path) >= 2 && path[1] == ':' {
		pre = path[:2]
		path = path[2:]
	}

	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)

	// Split directory into parts
	var parts []string
	if dir != "." && dir != "" {
		parts = strings.Split(dir, sep)
	}
	
	var cleanParts []string
	for _, p := range parts {
		if p == "." || p == "" || p == sep {
			continue
		}
		cp := CleanString(p)
		cp = strings.TrimLeft(cp, " -")
		cp = strings.TrimRight(cp, " -_.")
		if cp == "" {
			cp = "_"
		}
		
		if opts.LowercaseFolders {
			cp = strings.ToLower(cp)
		} else if opts.CaseInsensitive {
			if strings.ContainsAny(cp, " _.") {
				cp = Title(cp)
			} else {
				cp = strings.ToLower(cp)
			}
		}
		cleanParts = append(cleanParts, cp)
	}

	if opts.DedupeParts {
		cleanParts = OrderedSet(cleanParts)
	}

	// Clean stem
	cleanStem := CleanString(stem)
	cleanStem = strings.TrimLeft(cleanStem, " -")
	cleanStem = strings.TrimRight(cleanStem, " -.")

	// Shorten stem if too long
	fsLimit := opts.MaxNameLen - len(ext) - 1
	if len(cleanStem) > fsLimit && fsLimit > 3 {
		cleanStem = ShortenMiddle(cleanStem, fsLimit)
	}

	res := filepath.Join(append(cleanParts, cleanStem)...) + ext
	if opts.DotSpace {
		res = strings.ReplaceAll(res, " ", ".")
	}

	return pre + res
}
