package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func exportMessage(message map[string]interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}
	log.Printf("%s\n", data)
}

func maybeTrue(value string) bool {
	if value == "" {
		return true
	}
	return value == "true"
}

func main() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create log file with ISO timestamp
	timestamp := time.Now().Format(time.RFC3339)
	logFileName := fmt.Sprintf("logs/%s.log", timestamp)

	largestLogSizeInBytes := 0
	requestCounter := 0
	includeBodyInJSON := maybeTrue(os.Getenv("INCLUDE_BODY_IN_JSON"))
	bodyAsBase64 := maybeTrue(os.Getenv("BODY_AS_BASE64"))

	log.Printf("includeBodyInJSON: %t", includeBodyInJSON)
	log.Printf("bodyAsBase64: %t", bodyAsBase64)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestCounter++

		// Create the log file
		logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, "Failed to create log file", http.StatusInternalServerError)
			return
		}
		defer logFile.Close()

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		exportMessage(map[string]interface{}{
			"type":           "received_size",
			"size":           len(body),
			"request_number": requestCounter,
		})

		// Create metadata map
		metadata := map[string]interface{}{
			"request_number": requestCounter,
			"url":            r.URL.String(),
			"method":         r.Method,
			"headers":        r.Header,
			"is_body_base64": bodyAsBase64,
			"time":           time.Now().Format(time.RFC3339),
		}

		// Add body to metadata if environment variable is set
		if includeBodyInJSON {
			if bodyAsBase64 {
				metadata["body"] = base64.StdEncoding.EncodeToString(body)
			} else {
				metadata["body"] = string(body)
			}
		}

		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %v", err)
		} else {
			// Append the metadata to the log file
			if _, err := logFile.Write(metadataJSON); err != nil {
				log.Fatalf("Failed to write to log file: %v", err)
			} else {
				if _, err := logFile.Write([]byte("\n")); err != nil {
					log.Fatalf("Failed to write to log file: %v", err)
				}
			}
		}

		// Append the request body to the log file only if not including it in metadata
		if !includeBodyInJSON {
			updatedBody := body
			if bodyAsBase64 {
				updatedBody = []byte(base64.StdEncoding.EncodeToString(body))
			}
			if _, err := logFile.Write(updatedBody); err != nil {
				http.Error(w, "Failed to write to log file", http.StatusInternalServerError)
				return
			} else {
				if _, err := logFile.Write([]byte("\n")); err != nil {
					http.Error(w, "Failed to write to log file", http.StatusInternalServerError)
					return
				}
			}
		}

		if len(body) > largestLogSizeInBytes {
			largestLogSizeInBytes = len(body)
			exportMessage(map[string]interface{}{
				"type":           "largest_yet",
				"size":           len(body),
				"request_number": requestCounter,
			})
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte{})
	})

	// Start the server
	port := ":8081"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
