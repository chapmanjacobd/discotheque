package commands

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

type CategorizeCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`

	Other bool `help:"Analyze 'other' category to find potential new categories"`
}

func (c CategorizeCmd) IsQueryTrait()  {}
func (c CategorizeCmd) IsFilterTrait() {}
func (c CategorizeCmd) IsActionTrait() {}

var defaultCategories = map[string][]string{
	"sports":      {"sports?", "football", "soccer", "basketball", "tennis", "olympics", "training"},
	"fitness":     {"workout", "fitness", "gym", "yoga", "pilates", "exercise", "bodybuilding", "cardio"},
	"documentary": {"documentaries", "documentary", "docu", "history", "biography", "nature", "science", "planet", "wildlife", "factual"},
	"comedy":      {"comedy", "comedies", "standup", "funny", "sitcom", "humor", "prank", "roast", "satire"},
	"music":       {"music", "concerts?", "performance", "live", "musical", "video clip", "remix(es)?", "feat", "official video", "soundtracks?"},
	"educational": {"educational", "tutorials?", "lessons?", "lectures?", "courses?", "learning", "how to", "explainers?", "masterclass(es)?"},
	"news":        {"news", "reports?", "politics", "interviews?", "journalists?", "coverage", "current affairs", "broadcasts?", "press release"},
	"gaming":      {"gaming", "gameplay", "walkthroughs?", "playthroughs?", "twitch", "nintendo", "playstation", "xbox", "steam", "speedruns?", "lets play"},
	"tech":        {"tech", "technology", "software", "hardware", "programming", "coding", "reviews?", "unboxings?", "gadgets?", "silicon"},
	"audiobook":   {"audiobooks?", "audio book", "narrated", "reading", "unabridged"},
}

func (c *CategorizeCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}
	media = query.FilterMedia(media, c.GlobalFlags)

	if len(media) == 0 {
		return fmt.Errorf("no media found")
	}

	// Compile regexes once
	compiled := make(map[string][]*regexp.Regexp)
	for cat, keywords := range defaultCategories {
		for _, kw := range keywords {
			re, err := regexp.Compile(`(?i)\b` + kw + `\b`)
			if err != nil {
				slog.Error("Failed to compile regex", "keyword", kw, "error", err)
				continue
			}
			compiled[cat] = append(compiled[cat], re)
		}
	}

	if c.Other {
		return c.mineCategories(media, compiled)
	}

	return c.applyCategories(media, compiled)
}

func (c *CategorizeCmd) applyCategories(media []models.MediaWithDB, compiled map[string][]*regexp.Regexp) error {
	categorizedCount := 0
	for _, m := range media {
		foundCategories := []string{}
		pathAndTitle := m.Path
		if m.Title != nil {
			pathAndTitle += " " + *m.Title
		}

		for cat, res := range compiled {
			for _, re := range res {
				if re.MatchString(pathAndTitle) {
					foundCategories = append(foundCategories, cat)
					break
				}
			}
		}

		if len(foundCategories) > 0 {
			newCategories := strings.Join(foundCategories, ";")
			if m.Categories != nil && *m.Categories != "" {
				existing := strings.Split(*m.Categories, ";")
				merged := make(map[string]bool)
				for _, e := range existing {
					merged[strings.TrimSpace(e)] = true
				}
				for _, f := range foundCategories {
					merged[f] = true
				}
				combined := []string{}
				for k := range merged {
					combined = append(combined, k)
				}
				sort.Strings(combined)
				newCategories = strings.Join(combined, ";")
			}

			if !c.Simulate {
				sqlDB, err := db.Connect(m.DB)
				if err != nil {
					slog.Error("Failed to connect to database", "db", m.DB, "error", err)
					continue
				}
				queries := db.New(sqlDB)
				err = queries.UpdateMediaCategories(context.Background(), db.UpdateMediaCategoriesParams{
					Categories: utils.ToNullString(newCategories),
					Path:       m.Path,
				})
				sqlDB.Close()
				if err != nil {
					slog.Error("Failed to update categories", "path", m.Path, "error", err)
					continue
				}
			}

			if c.Verbose {
				fmt.Printf("Categorized: %s -> %s\n", m.Path, newCategories)
			}
			categorizedCount++
		}
	}

	fmt.Printf("Processed %d files, categorized %d\n", len(media), categorizedCount)
	return nil
}

func (c *CategorizeCmd) mineCategories(media []models.MediaWithDB, compiled map[string][]*regexp.Regexp) error {
	wordCounts := make(map[string]int)
	unmatchedCount := 0

	for _, m := range media {
		matched := false
		pathAndTitle := m.Path
		if m.Title != nil {
			pathAndTitle += " " + *m.Title
		}

		for _, res := range compiled {
			for _, re := range res {
				if re.MatchString(pathAndTitle) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}

		if !matched {
			unmatchedCount++
			words := utils.ExtractWords(utils.PathToSentence(m.Path))
			if m.Title != nil {
				words = append(words, utils.ExtractWords(*m.Title)...)
			}

			for _, word := range words {
				if len(word) < 4 {
					continue
				}
				wordCounts[word]++
			}
		}
	}

	type wordFreq struct {
		word  string
		count int
	}
	var freqs []wordFreq
	for w, c := range wordCounts {
		if c > 1 {
			freqs = append(freqs, wordFreq{w, c})
		}
	}

	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].count > freqs[j].count
	})

	fmt.Printf("Mined %d unmatched files. Top potential keywords:\n", unmatchedCount)
	limit := min(len(freqs), 50)
	for i := range limit {
		fmt.Printf("%s: %d\n", freqs[i].word, freqs[i].count)
	}

	return nil
}
