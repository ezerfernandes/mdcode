package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type fileInfo struct {
	File         string  `json:"file"`
	TopAuthor    string  `json:"top_author"`
	TopPercent   float64 `json:"top_percent"`
	TotalAuthors int     `json:"total_authors"`
	TotalCommits int     `json:"total_commits"`
}

type report struct {
	BusFactor int        `json:"bus_factor"`
	Files     []fileInfo `json:"files"`
}

func main() {
	jsonFlag := flag.Bool("json", false, "output as JSON")
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	files, err := trackedFiles(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "busfactor: %v\n", err)
		os.Exit(1)
	}

	var results []fileInfo
	for _, f := range files {
		info, err := analyzeFile(f)
		if err != nil {
			continue
		}
		results = append(results, info)
	}

	// Sort by top_percent descending (highest concentration = highest risk first).
	sort.Slice(results, func(i, j int) bool {
		return results[i].TopPercent > results[j].TopPercent
	})

	bf := busFactor(results)

	if *jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report{BusFactor: bf, Files: results})
		return
	}

	fmt.Printf("Bus Factor: %d\n\n", bf)
	fmt.Printf("%-60s %-25s %8s %8s\n", "FILE", "TOP AUTHOR", "COMMIT%", "AUTHORS")
	fmt.Println(strings.Repeat("-", 105))
	for _, r := range results {
		fmt.Printf("%-60s %-25s %7.1f%% %8d\n", r.File, r.TopAuthor, r.TopPercent, r.TotalAuthors)
	}
}

// trackedFiles returns git-tracked files under the given path.
func trackedFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", root)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// Only include regular files (skip directories, submodules, etc.)
		info, err := os.Stat(line)
		if err != nil || info.IsDir() {
			continue
		}
		// Skip binary-looking files by extension.
		ext := strings.ToLower(filepath.Ext(line))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".zip", ".tar", ".gz", ".exe", ".bin":
			continue
		}
		files = append(files, line)
	}
	return files, nil
}

// analyzeFile collects per-author commit counts for a single file.
func analyzeFile(path string) (fileInfo, error) {
	cmd := exec.Command("git", "log", "--format=%aN", "--no-merges", "--follow", "--", path)
	out, err := cmd.Output()
	if err != nil {
		return fileInfo{}, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return fileInfo{}, fmt.Errorf("no commits for %s", path)
	}

	counts := map[string]int{}
	for _, author := range strings.Split(raw, "\n") {
		author = strings.TrimSpace(author)
		if author != "" {
			counts[author]++
		}
	}

	var topAuthor string
	var topCount, total int
	for author, c := range counts {
		total += c
		if c > topCount {
			topCount = c
			topAuthor = author
		}
	}

	return fileInfo{
		File:         path,
		TopAuthor:    topAuthor,
		TopPercent:   float64(topCount) / float64(total) * 100,
		TotalAuthors: len(counts),
		TotalCommits: total,
	}, nil
}

// busFactor computes the minimum number of authors whose departure would
// leave >50% of files without their top contributor.
func busFactor(files []fileInfo) int {
	// Count how many files each author is the top contributor for.
	topFor := map[string]int{}
	for _, f := range files {
		topFor[f.TopAuthor]++
	}

	// Sort authors by number of files they are top contributor for (descending).
	type authorCount struct {
		author string
		count  int
	}
	var authors []authorCount
	for a, c := range topFor {
		authors = append(authors, authorCount{a, c})
	}
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].count > authors[j].count
	})

	threshold := len(files) / 2
	affected := 0
	for i, ac := range authors {
		affected += ac.count
		if affected > threshold {
			return i + 1
		}
	}
	return len(authors)
}
