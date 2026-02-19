package utils

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func SafeMean[T Number](slice []T) float64 {
	if len(slice) == 0 {
		return 0
	}
	var sum float64
	for _, v := range slice {
		sum += float64(v)
	}
	return sum / float64(len(slice))
}

func SafeMedian[T Number](slice []T) float64 {
	if len(slice) == 0 {
		return 0
	}
	sorted := make([]float64, len(slice))
	for i, v := range slice {
		sorted[i] = float64(v)
	}
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

func HumanToBytes(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"KB", 1024},
		{"MB", 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"K", 1024},
		{"M", 1024 * 1024},
		{"G", 1024 * 1024 * 1024},
		{"T", 1024 * 1024 * 1024 * 1024},
		{"B", 1},
	}

	for _, entry := range suffixes {
		if before, ok := strings.CutSuffix(s, entry.suffix); ok {
			numStr := strings.TrimSpace(before)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(entry.mult)), nil
		}
	}

	num, err := strconv.ParseInt(s, 10, 64)
	return num, err
}

func HumanToBits(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"KBIT", 1000},
		{"MBIT", 1000 * 1000},
		{"GBIT", 1000 * 1000 * 1000},
		{"TBIT", 1000 * 1000 * 1000 * 1000},
		{"K", 1000},
		{"M", 1000 * 1000},
		{"G", 1000 * 1000 * 1000},
		{"T", 1000 * 1000 * 1000 * 1000},
		{"BIT", 1},
	}

	for _, entry := range suffixes {
		if before, ok := strings.CutSuffix(s, entry.suffix); ok {
			numStr := strings.TrimSpace(before)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(entry.mult)), nil
		}
	}

	num, err := strconv.ParseInt(s, 10, 64)
	return num, err
}

func HumanToSeconds(s string) (int64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 0, nil
	}

	multipliers := []struct {
		suffix string
		mult   int64
	}{
		{"minutes", 60},
		{"seconds", 1},
		{"months", 2592000},
		{"weeks", 604800},
		{"hours", 3600},
		{"years", 31536000},
		{"minute", 60},
		{"second", 1},
		{"month", 2592000},
		{"week", 604800},
		{"hour", 3600},
		{"year", 31536000},
		{"mins", 60},
		{"secs", 1},
		{"min", 60},
		{"sec", 1},
		{"days", 86400},
		{"day", 86400},
		{"mon", 2592000},
		{"mo", 2592000},
		{"yr", 31536000},
		{"hr", 3600},
		{"w", 604800},
		{"d", 86400},
		{"h", 3600},
		{"m", 60},
		{"s", 1},
		{"y", 31536000},
	}

	for _, entry := range multipliers {
		if before, ok := strings.CutSuffix(s, entry.suffix); ok {
			numStr := strings.TrimSpace(before)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(entry.mult)), nil
		}
	}

	// Default to seconds
	return strconv.ParseInt(s, 10, 64)
}

func ParseRange(s string, humanToX func(string) (int64, error)) (Range, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Range{}, nil
	}

	if strings.Contains(s, "%") {
		parts := strings.Split(s, "%")
		base, err := humanToX(parts[0])
		if err != nil {
			return Range{}, err
		}
		percent, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return Range{}, err
		}
		tolerance := int64(float64(base) * (percent / 100.0))
		min := base - tolerance
		max := base + tolerance
		return Range{Min: &min, Max: &max}, nil
	}

	if strings.HasPrefix(s, ">") {
		min, err := humanToX(s[1:])
		if err != nil {
			return Range{}, err
		}
		min++ // strictly greater
		return Range{Min: &min}, nil
	}
	if strings.HasPrefix(s, "<") {
		max, err := humanToX(s[1:])
		if err != nil {
			return Range{}, err
		}
		max-- // strictly less
		return Range{Max: &max}, nil
	}
	if strings.HasPrefix(s, "+") {
		min, err := humanToX(s[1:])
		if err != nil {
			return Range{}, err
		}
		return Range{Min: &min}, nil
	}
	if strings.HasPrefix(s, "-") {
		max, err := humanToX(s[1:])
		if err != nil {
			return Range{}, err
		}
		return Range{Max: &max}, nil
	}

	val, err := humanToX(s)
	if err != nil {
		return Range{}, err
	}
	return Range{Value: &val}, nil
}

func Percent(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}

func FloatFromPercent(s string) (float64, error) {
	if before, ok := strings.CutSuffix(s, "%"); ok {
		v, err := strconv.ParseFloat(before, 64)
		if err != nil {
			return 0, err
		}
		return v / 100, nil
	}
	return strconv.ParseFloat(s, 64)
}

func PercentageDifference(v1, v2 float64) float64 {
	if v1+v2 == 0 {
		return 100.0
	}
	return math.Abs((v1-v2)/((v1+v2)/2)) * 100
}

func CalculateSegments(total float64, chunk float64, gap float64) []float64 {
	if total <= 0 || chunk <= 0 {
		return nil
	}
	if total <= chunk*3 {
		return []float64{0}
	}

	var segments []float64
	start := 0.0
	endSegmentStart := total - chunk

	g := gap
	if g < 1 {
		g = math.Ceil(total * gap)
	}

	for start+chunk < endSegmentStart {
		segments = append(segments, start)
		start += chunk + g
	}

	return append(segments, endSegmentStart)
}

func SafeInt(s string) *int {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	i := int(f)
	return &i
}

func SafeFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func SqlHumanTime(s string) string {
	if _, err := strconv.Atoi(s); err == nil {
		return s + " minutes"
	}

	unitMapping := map[string]string{
		"min":  "minutes",
		"mins": "minutes",
		"s":    "seconds",
		"sec":  "seconds",
		"secs": "seconds",
	}

	re := regexp.MustCompile(`(\d+\.?\d*)([a-zA-Z]+)`)
	match := re.FindStringSubmatch(s)
	if match != nil {
		value := match[1]
		unit := strings.ToLower(strings.TrimSpace(match[2]))
		if mapped, ok := unitMapping[unit]; ok {
			unit = mapped
		}
		return fmt.Sprintf("%s %s", value, unit)
	}
	return s
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}
