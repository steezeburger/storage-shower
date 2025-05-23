package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/steezeburger/storage-shower/backend/internal/fileinfo"
	"github.com/steezeburger/storage-shower/backend/internal/scan"
)

// StartServer starts the HTTP server on port 8080
func StartServer() int {
	// Set up file server for frontend directory
	fs := http.FileServer(http.Dir("../"))

	// Serve frontend files with explicit MIME types
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			http.ServeFile(w, r, "../frontend/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Set specific MIME types for JavaScript and CSS
	http.HandleFunc("/frontend/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "../frontend/app.js")
	})

	http.HandleFunc("/frontend/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "../frontend/styles.css")
	})

	// API endpoints
	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/scan/status", handleScanStatus)
	http.HandleFunc("/api/scan/stop", handleScanStop)
	http.HandleFunc("/api/home", handleHome)
	http.HandleFunc("/api/browse", handleBrowse)
	http.HandleFunc("/api/results", handleResults)
	http.HandleFunc("/api/previous-scans", handlePreviousScans)

	// Use port 8080
	port := 8080

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", port)
		if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Load previous scans
	scan.LoadPreviousScans()

	// Open the browser
	openBrowser(fmt.Sprintf("http://localhost:%d", port))

	return port
}

// handleScan initiates a new directory scan
func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var requestData struct {
		Path         string `json:"path"`
		IgnoreHidden bool   `json:"ignoreHidden"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if another scan is in progress
	status := scan.GetScanStatus()
	if status.InProgress {
		http.Error(w, "Another scan is already in progress", http.StatusConflict)
		return
	}

	// Start scan in a goroutine
	go func() {
		result, err := scan.ScanDirectory(requestData.Path, requestData.IgnoreHidden)
		if err != nil {
			log.Printf("Scan error: %v", err)
		} else {
			log.Printf("Scan completed: %s, found %d items", requestData.Path, len(result.Children))
		}
	}()

	// Return success
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "scan_started",
	})
}

// handleScanStatus returns the current scan status
func handleScanStatus(w http.ResponseWriter, r *http.Request) {
	status := scan.GetScanStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleScanStop cancels an in-progress scan
func handleScanStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cancelled := scan.CancelScan()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"cancelled": cancelled,
	})
}

// handleHome returns the user's home directory path
func handleHome(w http.ResponseWriter, r *http.Request) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Failed to get home directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": homeDir,
	})
}

// handleBrowse returns a list of directories in the specified path
func handleBrowse(w http.ResponseWriter, r *http.Request) {
	// Get the current path from query parameters
	queryValues := r.URL.Query()
	currentPath := queryValues.Get("path")

	// If no path provided, default to user's home directory
	if currentPath == "" {
		var err error
		currentPath, err = os.UserHomeDir()
		if err != nil {
			http.Error(w, "Failed to get home directory", http.StatusInternalServerError)
			return
		}
	}

	// Get a list of directories at this path
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	type DirectoryEntry struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		IsDir    bool   `json:"isDir"`
		IsHidden bool   `json:"isHidden"`
	}

	var directories []DirectoryEntry

	// Process only directories
	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := filepath.Join(currentPath, entry.Name())
			isHidden := fileinfo.IsHidden(entryPath)
			
			directories = append(directories, DirectoryEntry{
				Name:     entry.Name(),
				Path:     entryPath,
				IsDir:    true,
				IsHidden: isHidden,
			})
		}
	}

	// Prepare the response
	response := struct {
		CurrentPath  string           `json:"currentPath"`
		ParentPath   string           `json:"parentPath"`
		Directories  []DirectoryEntry `json:"directories"`
	}{
		CurrentPath: currentPath,
		ParentPath:  filepath.Dir(currentPath),
		Directories: directories,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleResults returns the most recent scan results
func handleResults(w http.ResponseWriter, r *http.Request) {
	if len(scan.PreviousScans) == 0 {
		http.Error(w, "No scan results available", http.StatusNotFound)
		return
	}

	// Return the most recent scan
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scan.PreviousScans[0])
}

// handlePreviousScans returns all previous scan records
func handlePreviousScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scan.PreviousScans)
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser: %v", err)
		log.Printf("Please open your browser and navigate to: %s", url)
	}
}