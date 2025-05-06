package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LogFile struct {
	Name string `json:"name"`
}

type LogEntry struct {
	RequestNumber int                 `json:"request_number"`
	URL           string              `json:"url"`
	Method        string              `json:"method"`
	Headers       map[string][]string `json:"headers"`
	Time          string              `json:"time,omitempty"`
	IsBodyBase64  bool                `json:"is_body_base64"`
	Body          interface{}         `json:"body,omitempty"`
	RawEntry      string              `json:"raw_entry,omitempty"`
}

type PaginatedResponse struct {
	Entries    []LogEntry `json:"entries"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	HasMore    bool       `json:"has_more"`
}

func main() {
	// Enable CORS
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Serve static files
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	// List log files endpoint
	http.HandleFunc("/api/logs", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		files, err := filepath.Glob("logs/*.log")
		if err != nil {
			http.Error(w, "Failed to read logs directory", http.StatusInternalServerError)
			return
		}

		logFiles := make([]LogFile, 0, len(files))
		for _, file := range files {
			// Get just the filename without path and extension
			baseName := filepath.Base(file)
			nameWithoutExt := strings.TrimSuffix(baseName, ".log")
			logFiles = append(logFiles, LogFile{Name: nameWithoutExt})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logFiles)
	}))

	// Search log entries endpoint
	http.HandleFunc("/api/logs/search", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get query parameters
		fileName := r.URL.Query().Get("file")
		searchTerm := r.URL.Query().Get("q")
		startTime := strings.ReplaceAll(r.URL.Query().Get("start_time"), " ", "+")
		endTime := strings.ReplaceAll(r.URL.Query().Get("end_time"), " ", "+")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		// Set defaults
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}

		// Validate file name
		if fileName == "" {
			http.Error(w, "File name is required", http.StatusBadRequest)
			return
		}

		// Replace space back to '+' for file system operations
		fileName = strings.ReplaceAll(fileName, " ", "+")
		relativeFileName := fmt.Sprintf("logs/%s.log", fileName)

		// Open the log file
		file, err := os.Open(relativeFileName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Log file by name %s not found", relativeFileName), http.StatusNotFound)
			return
		}
		defer file.Close()

		// Parse time range if provided
		var startTimeParsed, endTimeParsed time.Time
		if startTime != "" {
			startTimeParsed, err = time.Parse(time.RFC3339, startTime)
			if err != nil {
				http.Error(w, "Invalid start_time format. Use RFC3339 format", http.StatusBadRequest)
				return
			}
		}
		if endTime != "" {
			endTimeParsed, err = time.Parse(time.RFC3339, endTime)
			if err != nil {
				http.Error(w, "Invalid end_time format. Use RFC3339 format", http.StatusBadRequest)
				return
			}
		}

		// Read and parse entries
		scanner := bufio.NewScanner(file)
		var entries []LogEntry = []LogEntry{}
		var totalCount int

		for scanner.Scan() {
			line := scanner.Text()
			if searchTerm != "" && !strings.Contains(line, searchTerm) {
				continue
			}

			var entry LogEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				// If JSON parsing fails, include the raw entry
				entry = LogEntry{RawEntry: line}
			}

			// Filter by time if time range is provided
			if entry.Time != "" {
				entryTime, err := time.Parse(time.RFC3339, entry.Time)
				if err == nil {
					if !startTimeParsed.IsZero() && entryTime.Before(startTimeParsed) {
						continue
					}
					if !endTimeParsed.IsZero() && entryTime.After(endTimeParsed) {
						continue
					}
				}
			}

			totalCount++
			if totalCount <= (page-1)*pageSize || totalCount > page*pageSize {
				continue
			}

			entries = append(entries, entry)
		}

		if err := scanner.Err(); err != nil {
			http.Error(w, "Error reading log file", http.StatusInternalServerError)
			return
		}

		response := PaginatedResponse{
			Entries:    entries,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
			HasMore:    totalCount > page*pageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Start the server
	port := ":8082"
	fmt.Printf("View server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
