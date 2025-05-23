package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/steezeburger/storage-shower/backend/internal/fileinfo"
	"github.com/steezeburger/storage-shower/backend/internal/scan"
)

// StartServer starts the HTTP server and returns the port it's listening on
func StartServer(frontendFS embed.FS) int {
	// Set up API routes
	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/scan/status", handleScanStatus)
	http.HandleFunc("/api/scan/stop", handleScanStop)
	http.HandleFunc("/api/home", handleHome)
	http.HandleFunc("/api/browse", handleBrowse)
	http.HandleFunc("/api/results", handleResults)
	http.HandleFunc("/api/previous-scans", handlePreviousScans)

	// Serve frontend files
	setupFrontendHandlers(frontendFS)

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

// setupFrontendHandlers configures handlers for serving the embedded frontend files
func setupFrontendHandlers(frontendFS embed.FS) {
	// Get the frontend subfolder
	frontendSubFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to get frontend subfolder: %v", err)
	}

	// Serve the index.html file for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			content, err := fs.ReadFile(frontendSubFS, "index.html")
			if err != nil {
				http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(content)
			return
		}
		
		// For non-root paths, try to serve static files
		filePath := r.URL.Path
		if filePath != "/" {
			filePath = strings.TrimPrefix(filePath, "/")
			serveStaticFile(w, r, frontendSubFS, filePath)
			return
		}
		
		// If we get here, return 404
		http.NotFound(w, r)
	})
	
	// Add explicit handlers for the main frontend assets
	http.HandleFunc("/frontend/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		content, err := fs.ReadFile(frontendSubFS, "app.js")
		if err != nil {
			http.Error(w, "Failed to read app.js", http.StatusInternalServerError)
			return
		}
		w.Write(content)
	})
	
	http.HandleFunc("/frontend/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		content, err := fs.ReadFile(frontendSubFS, "styles.css")
		if err != nil {
			http.Error(w, "Failed to read styles.css", http.StatusInternalServerError)
			return
		}
		w.Write(content)
	})
}

// serveStaticFile serves a static file from the embedded filesystem with the correct MIME type
func serveStaticFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, path string) {
	// Set appropriate MIME type based on file extension
	switch {
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html")
	case strings.HasSuffix(path, ".json"):
		w.Header().Set("Content-Type", "application/json")
	case strings.HasSuffix(path, ".png"):
		w.Header().Set("Content-Type", "image/png")
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		w.Header().Set("Content-Type", "image/jpeg")
	case strings.HasSuffix(path, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	}
	
	// Try to read the file
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	// Write the content to the response
	w.Write(content)
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

	// Validate path
	if requestData.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
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
		_, err := scan.ScanDirectory(requestData.Path, requestData.IgnoreHidden)
		if err != nil {
			log.Printf("Scan error: %v", err)
		} else {
			log.Printf("Scan completed: %s", requestData.Path)
		}
	}()

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "started",
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
	json.NewEncoder(w).Encode(map[string]string{
		"status": cancelled,
	})
}

// handleHome returns the user's home directory path
func handleHome(w http.ResponseWriter, r *http.Request) {
	u, err := user.Current()
	if err != nil {
		http.Error(w, "Failed to get home directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"home": u.HomeDir,
	})
}

// handleBrowse opens a native file dialog and returns the selected directory
func handleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var selectedPath string
	var err error

	switch runtime.GOOS {
	case "darwin":
		// Use AppleScript to open a folder selection dialog
		cmd := exec.Command("osascript", "-e", `choose folder with prompt "Select a directory to scan"`)
		output, err := cmd.Output()
		if err == nil {
			// AppleScript returns something like 'alias Macintosh HD:Users:username:Documents:'
			// Convert it to a regular path
			selectedPath = strings.TrimSpace(string(output))
			// Remove 'alias ' prefix if present
			selectedPath = strings.TrimPrefix(selectedPath, "alias ")
			// Convert to POSIX path
			cmd = exec.Command("osascript", "-e", fmt.Sprintf(`POSIX path of "%s"`, selectedPath))
			output, err = cmd.Output()
			if err == nil {
				selectedPath = strings.TrimSpace(string(output))
			}
		}
	case "windows":
		// Use PowerShell to open folder browser dialog
		script := `
		Add-Type -AssemblyName System.Windows.Forms
		$folderBrowser = New-Object System.Windows.Forms.FolderBrowserDialog
		$folderBrowser.Description = "Select a directory to scan"
		$folderBrowser.RootFolder = 'MyComputer'
		if ($folderBrowser.ShowDialog() -eq 'OK') {
			$folderBrowser.SelectedPath
		}
		`
		cmd := exec.Command("powershell", "-Command", script)
		output, err := cmd.Output()
		if err == nil {
			selectedPath = strings.TrimSpace(string(output))
		}
	default: // Linux
		// Use zenity for Linux
		cmd := exec.Command("zenity", "--file-selection", "--directory", "--title=Select a directory to scan")
		output, err := cmd.Output()
		if err == nil {
			selectedPath = strings.TrimSpace(string(output))
		}
	}

	if err != nil || selectedPath == "" {
		log.Printf("Error opening file dialog: %v", err)
		http.Error(w, "Failed to open file dialog", http.StatusInternalServerError)
		return
	}

	log.Printf("User selected directory: %s", selectedPath)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": selectedPath,
	})
}

// handleResults returns the scan results
func handleResults(w http.ResponseWriter, r *http.Request) {
	// Check if a specific result ID is requested
	resultID := r.URL.Query().Get("id")

	var result fileinfo.FileInfo
	var err error

	if resultID != "" {
		// Get a specific scan result by ID
		result, err = scan.GetScanResultByID(resultID)
	} else {
		// Get the most recent scan result
		result, err = scan.GetLatestScanResult()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handlePreviousScans returns a list of previous scan records
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