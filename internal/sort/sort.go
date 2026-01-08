package sort

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

type Method string

const (
	ByPath         Method = "path"
	ByTitle        Method = "title"
	ByDuration     Method = "duration"
	BySize         Method = "size"
	ByTimeCreated  Method = "time_created"
	ByTimeModified Method = "time_modified"
	ByTimePlayed   Method = "time_last_played"
	ByPlayCount    Method = "play_count"
)

func Apply(media []db.Media, method Method, reverse bool, natural bool) {
	less := makeLessFunc(media, method, natural)

	if reverse {
		sort.Slice(media, func(i, j int) bool { return !less(i, j) })
	} else {
		sort.Slice(media, less)
	}
}

func makeLessFunc(media []db.Media, method Method, natural bool) func(i, j int) bool {
	switch method {
	case ByPath:
		if natural {
			return func(i, j int) bool {
				return naturalLess(media[i].Path, media[j].Path)
			}
		}
		return func(i, j int) bool { return media[i].Path < media[j].Path }
	case ByTitle:
		return func(i, j int) bool { return media[i].Title < media[j].Title }
	case ByDuration:
		return func(i, j int) bool { return media[i].Duration < media[j].Duration }
	case BySize:
		return func(i, j int) bool { return media[i].Size < media[j].Size }
	case ByTimeCreated:
		return func(i, j int) bool { return media[i].TimeCreated < media[j].TimeCreated }
	case ByTimeModified:
		return func(i, j int) bool { return media[i].TimeModified < media[j].TimeModified }
	case ByTimePlayed:
		return func(i, j int) bool { return media[i].TimeLastPlayed < media[j].TimeLastPlayed }
	case ByPlayCount:
		return func(i, j int) bool { return media[i].PlayCount < media[j].PlayCount }
	default:
		return func(i, j int) bool { return media[i].Path < media[j].Path }
	}
}

type chunk struct {
	str   string
	num   int
	isNum bool
}

func naturalLess(s1, s2 string) bool {
	n1, n2 := extractNumbers(s1), extractNumbers(s2)

	idx1, idx2 := 0, 0
	for idx1 < len(n1) && idx2 < len(n2) {
		if n1[idx1].isNum && n2[idx2].isNum {
			if n1[idx1].num != n2[idx2].num {
				return n1[idx1].num < n2[idx2].num
			}
		} else {
			if n1[idx1].str != n2[idx2].str {
				return n1[idx1].str < n2[idx2].str
			}
		}
		idx1++
		idx2++
	}

	return len(n1) < len(n2)
}

func extractNumbers(s string) []chunk {
	re := regexp.MustCompile(`\d+|\D+`)
	matches := re.FindAllString(s, -1)

	var chunks []chunk
	for _, m := range matches {
		if num, err := strconv.Atoi(m); err == nil {
			chunks = append(chunks, chunk{num: num, isNum: true})
		} else {
			chunks = append(chunks, chunk{str: strings.ToLower(m), isNum: false})
		}
	}
	return chunks
}
