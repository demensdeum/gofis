package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DefaultMaxGoroutines = 100

type SearchResult struct {
	Path string
	Info os.FileInfo
}

func main() {
	searchTermFlag := flag.String("n", "", "")
	extensionFlag := flag.String("e", "", "")
	rootPathFlag := flag.String("p", ".", "")
	ignoreDirs := flag.String("i", "node_modules,.git,.svn,vendor", "")
	flag.Parse()

	var searchTerm string
	var startPath string
	var extension string
	maxGoroutines := DefaultMaxGoroutines

	args := flag.Args()
	if len(args) > 0 {
		searchTerm = args[0]
		if len(args) > 1 {
			startPath = args[1]
		} else {
			startPath = *rootPathFlag
		}
		if len(args) > 2 {
			if val, err := strconv.Atoi(args[2]); err == nil && val > 0 {
				maxGoroutines = val
			}
		}
	} else {
		searchTerm = *searchTermFlag
		startPath = *rootPathFlag
		extension = *extensionFlag
	}

	if searchTerm == "" && extension == "" {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go \"regex_or_string\" \"searchPath\" [maxGoroutines]")
		fmt.Println("  go run main.go -n <name> [-e <extension>] [-p <path>]")
		return
	}

	absPath, err := filepath.Abs(startPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var re *regexp.Regexp
	if searchTerm != "" {
		re, _ = regexp.Compile("(?i)" + searchTerm)
	}

	ignoreList := strings.Split(*ignoreDirs, ",")
	fmt.Printf("Searching for: '%s' (Regex supported) in %s (Max Goroutines: %d)...\n\n", searchTerm, absPath, maxGoroutines)

	start := time.Now()
	results := make(chan SearchResult, 100)
	sem := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	wg.Add(1)
	go walkDir(absPath, searchTerm, re, extension, ignoreList, results, &wg, sem)

	go func() {
		wg.Wait()
		close(results)
	}()

	count := 0
	for res := range results {
		count++
		fmt.Printf("[%d] %-15s | %s\n", count, formatSize(res.Info.Size()), res.Path)
	}

	fmt.Printf("\nFinished: %d files found in %v\n", count, time.Since(start))
}

func walkDir(dir string, term string, re *regexp.Regexp, ext string, ignore []string, results chan<- SearchResult, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		shouldIgnore := false
		for _, i := range ignore {
			if name == i {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		if entry.IsDir() {
			wg.Add(1)
			go walkDir(fullPath, term, re, ext, ignore, results, wg, sem)
		} else {
			matchName := false
			if term == "" {
				matchName = true
			} else if re != nil {
				matchName = re.MatchString(name)
			} else {
				matchName = strings.Contains(strings.ToLower(name), strings.ToLower(term))
			}

			matchExt := ext == "" || strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext))

			if matchName && matchExt {
				info, err := entry.Info()
				if err == nil {
					results <- SearchResult{Path: fullPath, Info: info}
				}
			}
		}
	}
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
