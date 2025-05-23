package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Debug flag to control verbose logging
var debugMode = false

// Maximum number of previous scans to store
const maxPreviousScans = 10

//go:embed frontend
var frontendFS embed.FS

// ScanRecord represents a record of a previous scan
type ScanRecord struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	ResultID  string    `json:"resultId"`
	Size      int64     `json:"size"`
}

// List of previous scans
var previousScans []ScanRecord

// FileTypeStats represents statistics about file types in a directory
type FileTypeStats struct {
	Image    int64 `json:"image"`
	Video    int64 `json:"video"`
	Audio    int64 `json:"audio"`
	Document int64 `json:"document"`
	Archive  int64 `json:"archive"`
	Other    int64 `json:"other"`
}

// FileInfo represents a file or directory with its size and children
type FileInfo struct {
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Size         int64          `json:"size"`
	IsDir        bool           `json:"isDir"`
	Children     []FileInfo     `json:"children,omitempty"`
	Extension    string         `json:"extension,omitempty"`
	FileTypes    *FileTypeStats `json:"fileTypes,omitempty"`
}

// ScanProgress represents the current progress of a filesystem scan
type ScanProgress struct {
	TotalItems   int     `json:"totalItems"`
	ScannedItems int     `json:"scannedItems"`
	Progress     float64 `json:"progress"`
	CurrentPath  string  `json:"currentPath"`
}

var (
	scanInProgress   bool
	cancelScan       bool
	scanMutex        sync.Mutex
	currentScan      ScanProgress
	resultPath       string
	lastScannedItems int
	lastProgressTime time.Time
	scanDone         chan struct{}
)

func main() {
	// Check for debug flag in args
	for _, arg := range os.Args {
		if arg == "--debug" {
			debugMode = true
			break
		}
	}

	// Start server on a random available port
	port := startServer()

	// Open browser
	url := fmt.Sprintf("http://localhost:%d", port)
	log.Printf("Server started at %s", url)
	openBrowser(url)

	// Keep the server running
	select {}
}

func startServer() int {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/scan/status", handleScanStatus)
	http.HandleFunc("/api/scan/stop", handleScanStop)
	http.HandleFunc("/api/home", handleHome)
	http.HandleFunc("/api/browse", handleBrowse)
	http.HandleFunc("/api/results", handleResults)
	http.HandleFunc("/api/previous-scans", handlePreviousScans)

	// Serve static files with proper MIME types
	http.HandleFunc("/frontend/", serveFrontendFiles)

	// Listen on port 8080
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	port := 8080

	go func() {
		if err := http.Serve(listener, nil); err != nil {
			log.Fatal(err)
		}
	}()

	return port
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) {
	var err error

	log.Printf("Opening browser at URL: %s", url)
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default: // "linux", "freebsd", etc.
		err = exec.Command("xdg-open", url).Start()
	}

	if err != nil {
		log.Printf("Error opening browser: %v", err)
		log.Printf("Please open your browser and navigate to: %s", url)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	content, err := frontendFS.ReadFile("frontend/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func serveFrontendFiles(w http.ResponseWriter, r *http.Request) {
	// Remove the /frontend/ prefix to get the actual file path
	filePath := strings.TrimPrefix(r.URL.Path, "/frontend/")
	fullPath := "frontend/" + filePath

	log.Printf("Serving frontend file: %s", fullPath)

	content, err := frontendFS.ReadFile(fullPath)
	if err != nil {
		log.Printf("Error reading frontend file %s: %v", fullPath, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Set appropriate MIME type based on file extension
	switch filepath.Ext(filePath) {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.Write(content)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	u, err := user.Current()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error getting user home directory: %v", err)
		return
	}

	log.Printf("Returning home directory: %s", u.HomeDir)
	resp := map[string]string{"home": u.HomeDir}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error marshaling home directory response: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received scan request from %s", r.RemoteAddr)

	if r.Method != "POST" {
		log.Printf("Invalid method %s for scan endpoint", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scanMutex.Lock()
	if scanInProgress {
		scanMutex.Unlock()
		http.Error(w, "Scan already in progress", http.StatusConflict)
		return
	}
	scanInProgress = true
	cancelScan = false
	currentScan = ScanProgress{}
	scanDone = make(chan struct{})
	scanMutex.Unlock()

	var requestData struct {
		Path         string `json:"path"`
		IgnoreHidden bool   `json:"ignoreHidden"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Printf("Failed to decode scan request: %v", err)
		scanMutex.Lock()
		scanInProgress = false
		scanMutex.Unlock()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Scan request: path=%s, ignoreHidden=%t", requestData.Path, requestData.IgnoreHidden)

	if requestData.Path == "" {
		log.Printf("Scan request rejected: empty path")
		scanMutex.Lock()
		scanInProgress = false
		scanMutex.Unlock()
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	// Count total items for progress reporting (in a separate goroutine)
	log.Printf("Starting file count for path: %s", requestData.Path)
	go func() {
		countFiles(requestData.Path, requestData.IgnoreHidden)
	}()

	// Start the actual scan
	log.Printf("Starting directory scan for path: %s", requestData.Path)
	go func() {
		// Make sure we always mark the scan as finished
		defer func() {
			scanMutex.Lock()
			scanInProgress = false
			scanMutex.Unlock()
			log.Printf("Scan process finished for path: %s", requestData.Path)
		}()

		result, err := scanDirectory(requestData.Path, requestData.IgnoreHidden)
		if debugMode {
			log.Printf("DEBUG: Scan result root size: %d bytes", result.Size)
		}

		if err != nil {
			log.Printf("Scan failed for path %s: %v", requestData.Path, err)
			return
		}

		log.Printf("Scan completed successfully for path: %s", requestData.Path)

		// Always save result, even if it's partial
		// Ensure result has correct size data
		if debugMode {
			log.Printf("DEBUG: Before saving, root size is %d bytes with %d children",
				result.Size, len(result.Children))
		}

		// Log the top-level directories/files and their sizes
		if debugMode {
			for _, child := range result.Children {
				log.Printf("DEBUG: Top-level item: %s, size: %d bytes, isDir: %v",
					child.Name, child.Size, child.IsDir)
			}
		}

		// Save result to a temporary file
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling result: %v", err)
			return
		}

		tempFile, err := os.CreateTemp("", "storage-shower-*.json")
		if err != nil {
			log.Printf("Error creating temp file: %v", err)
			return
		}

		_, err = tempFile.Write(resultJSON)
		if err != nil {
			log.Printf("Error writing to temp file: %v", err)
			tempFile.Close()
			return
		}

		tempFile.Close()

		scanMutex.Lock()
		resultPath = tempFile.Name()
		// Add to previous scans
		resultID := filepath.Base(tempFile.Name())
		newScan := ScanRecord{
			Path:      requestData.Path,
			Timestamp: time.Now(),
			ResultID:  resultID,
			Size:      result.Size,
		}

		// Add the new scan at the beginning of the slice
		previousScans = append([]ScanRecord{newScan}, previousScans...)

		// Limit the number of previous scans
		if len(previousScans) > maxPreviousScans {
			previousScans = previousScans[:maxPreviousScans]
		}
		scanMutex.Unlock()

		log.Printf("Scan result saved to %s (%s)", tempFile.Name(), formatBytes(int64(len(resultJSON))))
	}()

	log.Printf("Scan started successfully for path: %s", requestData.Path)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "started"}`))
}

func handleScanStatus(w http.ResponseWriter, r *http.Request) {
	scanMutex.Lock()
	progress := currentScan
	inProgress := scanInProgress

	// Add a timeout check - if no progress in 30 seconds, consider scan stalled
	static := false
	if inProgress && progress.ScannedItems > 0 {
		static = true

		// Create a static check state if it doesn't exist
		if lastScannedItems == 0 {
			lastScannedItems = progress.ScannedItems
			lastProgressTime = time.Now()
		} else if progress.ScannedItems > lastScannedItems {
			// Progress is being made, update the last known state
			lastScannedItems = progress.ScannedItems
			lastProgressTime = time.Now()
			static = false
		} else if time.Since(lastProgressTime) > 30*time.Second {
			// No progress for 30 seconds, consider scan stalled
			log.Printf("Scan appears stalled - no progress for 30 seconds")
			static = true
		} else {
			static = false
		}
	}
	scanMutex.Unlock()

	response := struct {
		InProgress bool         `json:"inProgress"`
		Progress   ScanProgress `json:"progress"`
		Stalled    bool         `json:"stalled,omitempty"`
	}{
		InProgress: inProgress,
		Progress:   progress,
		Stalled:    static,
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func handleScanStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Scan stop requested")

	scanMutex.Lock()
	wasScanInProgress := scanInProgress
	cancelScan = true
	scanMutex.Unlock()

	if wasScanInProgress {
		log.Printf("Cancellation flag set, waiting for scan to finish...")
	} else {
		log.Printf("No scan in progress, nothing to stop")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "stopping"}`))
}

func handleResults(w http.ResponseWriter, r *http.Request) {
	// Check if a specific result ID is requested
	resultID := r.URL.Query().Get("id")

	scanMutex.Lock()
	currentResultPath := resultPath
	scans := previousScans
	scanMutex.Unlock()

	// If a specific result ID is requested, find its path
	if resultID != "" {
		found := false
		for _, scan := range scans {
			if scan.ResultID == resultID {
				// Extract the full path from the ResultID
				currentResultPath = filepath.Join(os.TempDir(), resultID)
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Requested scan result not found", http.StatusNotFound)
			return
		}
	} else if currentResultPath == "" {
		http.Error(w, "No scan results available", http.StatusNotFound)
		return
	}

	data, err := os.ReadFile(currentResultPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// handlePreviousScans returns a list of previous scan records
func handlePreviousScans(w http.ResponseWriter, r *http.Request) {
	scanMutex.Lock()
	scans := previousScans
	scanMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scans)
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
	response := map[string]string{"path": selectedPath}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func countFiles(rootPath string, ignoreHidden bool) {
	log.Printf("Starting file count for: %s", rootPath)
	var count int

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip items we can't access
		}

		// Check if we should cancel
		scanMutex.Lock()
		if cancelScan {
			scanMutex.Unlock()
			return filepath.SkipAll
		}
		scanMutex.Unlock()

		if ignoreHidden && isHidden(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		count++

		if count%1000 == 0 {
			log.Printf("Counted %d files so far...", count)
			scanMutex.Lock()
			currentScan.TotalItems = count
			scanMutex.Unlock()
		} else if count%100 == 0 {
			scanMutex.Lock()
			currentScan.TotalItems = count
			scanMutex.Unlock()
		}

		return nil
	})

	log.Printf("File count completed: %d total items found", count)
	scanMutex.Lock()
	currentScan.TotalItems = count
	scanMutex.Unlock()
}

func scanDirectory(rootPath string, ignoreHidden bool) (FileInfo, error) {
	log.Printf("Beginning directory scan of: %s", rootPath)
	// Use filepath.Clean to normalize the path
	rootPath = filepath.Clean(rootPath)
	// Convert to absolute path to ensure consistency
	absPath, err := filepath.Abs(rootPath)
	if err == nil {
		rootPath = absPath
	}
	log.Printf("Normalized path: %s", rootPath)

	rootInfo, err := os.Stat(rootPath)
	if err != nil {
		log.Printf("Failed to stat root path %s: %v", rootPath, err)
		return FileInfo{}, err
	}

	// Get initial size if this is a file
	initialSize := int64(0)
	if !rootInfo.IsDir() {
		initialSize = rootInfo.Size()
		log.Printf("Root is a file with size: %d bytes", initialSize)
	}

	rootName := filepath.Base(rootPath)
	log.Printf("Root name: %s", rootName)

	root := FileInfo{
		Name:  rootName,
		Path:  rootPath,
		IsDir: rootInfo.IsDir(),
		Size:  initialSize,
	}

	if !root.IsDir {
		root.Size = rootInfo.Size()
		root.Extension = strings.TrimPrefix(filepath.Ext(rootPath), ".")
		return root, nil
	}

	// Reset cancelScan flag
	scanMutex.Lock()
	cancelScan = false
	scanMutex.Unlock()

	// Create a map to keep track of directories by path
	dirMap := make(map[string]*FileInfo)
	dirMap[rootPath] = &root

	// Simple recursive scan function - more reliable than worker pool
	var scanRecursive func(path string) error
	scanRecursive = func(path string) error {
		// Check for cancellation
		scanMutex.Lock()
		if cancelScan {
			scanMutex.Unlock()
			return fmt.Errorf("scan canceled")
		}
		scanMutex.Unlock()

		// Get directory entries
		entries, err := os.ReadDir(path)
		if err != nil {
			log.Printf("Error reading directory %s: %v", path, err)
			return nil // Continue with other directories
		}

		// Process each entry
		for _, entry := range entries {
			// Check for cancellation frequently
			scanMutex.Lock()
			if cancelScan {
				scanMutex.Unlock()
				return fmt.Errorf("scan canceled")
			}
			scanMutex.Unlock()

			entryName := entry.Name()
			entryPath := filepath.Join(path, entryName)
			// Normalize entry path to ensure consistency
			entryPath = filepath.Clean(entryPath)
			absEntryPath, err := filepath.Abs(entryPath)
			if err == nil {
				entryPath = absEntryPath
			}
			log.Printf("Processing entry: %s (full path: %s)", entryName, entryPath)

			// Skip hidden files/directories if requested
			if ignoreHidden && isHidden(entryPath) {
				if entry.IsDir() {
					continue // Skip this directory
				}
				continue // Skip this file
			}

			// Get file info
			info, err := entry.Info()
			if err != nil {
				log.Printf("Error getting info for %s: %v", entryPath, err)
				continue // Skip this entry
			}

			// Create file info struct with proper size initialization
			fileSize := info.Size()
			if debugMode {
				log.Printf("DEBUG: Creating FileInfo for %s with size %d bytes (isDir: %v)",
					entryPath, fileSize, entry.IsDir())
			}

			// Initialize size based on file or directory
			var initialSize int64 = 0
			if !entry.IsDir() {
				initialSize = fileSize
			}

			fileInfo := FileInfo{
				Name:  entryName,
				Path:  entryPath,
				IsDir: entry.IsDir(),
				Size:  initialSize, // Only set size for files, directories will accumulate
			}

			log.Printf("Created FileInfo: name=%s, path=%s, isDir=%v, initialSize=%d",
				fileInfo.Name, fileInfo.Path, fileInfo.IsDir, fileInfo.Size)

			// Update progress
			scanMutex.Lock()
			currentScan.ScannedItems++
			if currentScan.TotalItems > 0 {
				currentScan.Progress = float64(currentScan.ScannedItems) / float64(currentScan.TotalItems)
			}
			currentScan.CurrentPath = entryPath
			scanMutex.Unlock()

			// Handle file or directory
			if !entry.IsDir() {
				// For files, ensure size is set correctly and add extension
				// Size should already be set in the struct initialization, but double-check
				if fileInfo.Size <= 0 {
					fileInfo.Size = info.Size()
					if debugMode {
						log.Printf("DEBUG: Set file size for %s to %d bytes", fileInfo.Path, fileInfo.Size)
					}
				}
				fileInfo.Extension = strings.TrimPrefix(filepath.Ext(entryPath), ".")

				// Add to parent directory
				parentPath := filepath.Dir(entryPath)
				// Ensure parent path is normalized too
				parentPath = filepath.Clean(parentPath)
				absParentPath, err := filepath.Abs(parentPath)
				if err == nil {
					parentPath = absParentPath
				}
				log.Printf("Parent path for %s: %s", entryPath, parentPath)
				if parent, ok := dirMap[parentPath]; ok {
					if debugMode {
						log.Printf("DEBUG: Adding file %s to parent directory %s", fileInfo.Path, parentPath)
					}
					parent.Children = append(parent.Children, fileInfo)
					if debugMode {
						log.Printf("DEBUG: Parent now has %d children", len(parent.Children))
					}

					// Update parent sizes for files with non-zero sizes
					size := fileInfo.Size
					if size > 0 {
						if debugMode {
							log.Printf("DEBUG: File size to add: %d bytes", size)
							log.Printf("DEBUG: Adding file size %d bytes for %s", size, fileInfo.Path)
						}
						// Add file size to all parent directories
						for p := parentPath; p != ""; p = filepath.Dir(p) {
							log.Printf("Updating size for parent: %s", p)
							if dir, ok := dirMap[p]; ok {
								dirSizeBefore := dir.Size
								dir.Size += size
								log.Printf("Updated dir %s size: %d + %d = %d bytes",
									dir.Path, dirSizeBefore, size, dir.Size)
							} else {
								log.Printf("Could not find directory %s in map, keys:", p)
								for key := range dirMap {
									log.Printf("  - %s", key)
								}
								break
							}
						}
					} else {
						if debugMode {
							log.Printf("DEBUG: Skipping size update because size is zero")
						}
					}
				} else {
					if debugMode {
						log.Printf("DEBUG: Could not find parent directory %s in map", parentPath)
					}
				}
			} else {
				// For directories, initialize size to 0
				if debugMode {
					log.Printf("DEBUG: Adding directory %s to dirMap", entryPath)
				}
				fileInfo.Size = 0
				dirMap[entryPath] = &fileInfo

				// Add to parent directory
				if entryPath != rootPath {
					parentPath := filepath.Dir(entryPath)
					// Ensure parent path is normalized
					parentPath = filepath.Clean(parentPath)
					absParentPath, err := filepath.Abs(parentPath)
					if err == nil {
						parentPath = absParentPath
					}
					log.Printf("Dir parent path for %s: %s", entryPath, parentPath)
					if parent, ok := dirMap[parentPath]; ok {
						if debugMode {
							log.Printf("DEBUG: Adding directory %s to parent %s", entryPath, parentPath)
						}
						parent.Children = append(parent.Children, fileInfo)
						if debugMode {
							log.Printf("DEBUG: Parent now has %d children", len(parent.Children))
						}
					} else {
						if debugMode {
							log.Printf("DEBUG: Could not find parent directory %s in map", parentPath)
						}
					}
				}

				// Recursively scan this directory
				err = scanRecursive(entryPath)
				if err != nil {
					// If scan was canceled, propagate the error
					if err.Error() == "scan canceled" {
						return err
					}
					// Otherwise, log and continue
					log.Printf("Error scanning subdirectory %s: %v", entryPath, err)
				}
			}
		}
		return nil
	}

	// Start the recursive scan
	log.Printf("Starting recursive scan of: %s", rootPath)
	scanErr := scanRecursive(rootPath)

	// Check if scan was canceled
	if scanErr != nil && scanErr.Error() == "scan canceled" {
		log.Printf("Scan was canceled")
		return root, nil
	}

	// Log dirMap contents
	log.Printf("Directory map contents before fixing sizes:")
	for path, info := range dirMap {
		log.Printf("  %s -> size: %d, isDir: %v", path, info.Size, info.IsDir)
	}

	// Fix any directory sizes that might be wrong
	if root.IsDir {
		log.Printf("Fixing directory sizes for root: %s", rootPath)
		// Make a new map with normalized keys to handle any inconsistencies
		normalizedDirMap := make(map[string]*FileInfo)
		for k, v := range dirMap {
			normalizedKey := filepath.Clean(k)
			absKey, err := filepath.Abs(normalizedKey)
			if err == nil {
				normalizedKey = absKey
			}
			normalizedDirMap[normalizedKey] = v
		}
		fixDirectorySizes(&root, normalizedDirMap)
	}

	log.Printf("Completed recursive scan of: %s (total size: %s)", rootPath, formatBytes(root.Size))
	return root, nil
}

// scanWorker function has been removed in favor of a simpler recursive approach

// Helper function to format bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// fixDirectorySizes ensures all directory sizes are the sum of their children
// and calculates file type statistics for directories
func fixDirectorySizes(dir *FileInfo, dirMap map[string]*FileInfo) int64 {
	if !dir.IsDir {
		return dir.Size
	}

	log.Printf("Fixing directory size for: %s", dir.Path)

	var totalSize int64 = 0
	fileTypeStats := &FileTypeStats{}
	
	for i := range dir.Children {
		log.Printf("  Child %d: %s (initial size: %d, isDir: %v)",
			i, dir.Children[i].Path, dir.Children[i].Size, dir.Children[i].IsDir)

		childSize := dir.Children[i].Size
		if dir.Children[i].IsDir {
			// Recursively fix child directory sizes
			childPath := dir.Children[i].Path
			// Normalize the child path for consistent lookup
			childPath = filepath.Clean(childPath)
			absChildPath, err := filepath.Abs(childPath)
			if err == nil {
				childPath = absChildPath
			}
			log.Printf("  Looking for child path in dirMap: %s", childPath)
			if childDir, ok := dirMap[childPath]; ok {
				log.Printf("  Found child in dirMap, recursing...")
				childSize = fixDirectorySizes(childDir, dirMap)
				// Update the size in our children array too
				dir.Children[i].Size = childSize
				dir.Children[i].FileTypes = childDir.FileTypes
				log.Printf("  Updated child size to: %d", childSize)
				
				// Aggregate file type stats from child directory
				if childDir.FileTypes != nil {
					fileTypeStats.Image += childDir.FileTypes.Image
					fileTypeStats.Video += childDir.FileTypes.Video
					fileTypeStats.Audio += childDir.FileTypes.Audio
					fileTypeStats.Document += childDir.FileTypes.Document
					fileTypeStats.Archive += childDir.FileTypes.Archive
					fileTypeStats.Other += childDir.FileTypes.Other
				}
			} else {
				log.Printf("  WARNING: Child directory not found in dirMap: %s", childPath)
			}
		} else {
			// For files, add their size to the appropriate file type category
			fileType := getFileType(dir.Children[i].Extension)
			switch fileType {
			case "image":
				fileTypeStats.Image += childSize
			case "video":
				fileTypeStats.Video += childSize
			case "audio":
				fileTypeStats.Audio += childSize
			case "document":
				fileTypeStats.Document += childSize
			case "archive":
				fileTypeStats.Archive += childSize
			default:
				fileTypeStats.Other += childSize
			}
		}
		totalSize += childSize
	}

	log.Printf("  Total size for %s: %d bytes", dir.Path, totalSize)
	log.Printf("  File type stats: Image: %d, Video: %d, Audio: %d, Document: %d, Archive: %d, Other: %d", 
		fileTypeStats.Image, fileTypeStats.Video, fileTypeStats.Audio, 
		fileTypeStats.Document, fileTypeStats.Archive, fileTypeStats.Other)

	// Set this directory's size and file type stats
	dir.Size = totalSize
	dir.FileTypes = fileTypeStats
	return totalSize
}

func isHidden(path string) bool {
	name := filepath.Base(path)
	return strings.HasPrefix(name, ".") && name != "." && name != ".."
}

// getFileType categorizes a file based on its extension
func getFileType(extension string) string {
	if extension == "" {
		return "other"
	}
	
	ext := strings.ToLower(extension)
	
	imageExts := map[string]bool{
		"jpg": true, "jpeg": true, "png": true, "gif": true, "svg": true,
		"webp": true, "bmp": true, "tiff": true, "ico": true, "heic": true,
	}
	
	videoExts := map[string]bool{
		"mp4": true, "mov": true, "avi": true, "mkv": true, "wmv": true,
		"flv": true, "webm": true, "m4v": true, "mpg": true, "mpeg": true,
	}
	
	audioExts := map[string]bool{
		"mp3": true, "wav": true, "ogg": true, "flac": true, "aac": true,
		"m4a": true, "wma": true, "aiff": true,
	}
	
	documentExts := map[string]bool{
		"pdf": true, "doc": true, "docx": true, "xls": true, "xlsx": true,
		"ppt": true, "pptx": true, "txt": true, "rtf": true, "md": true,
		"csv": true, "json": true, "xml": true, "html": true, "css": true,
		"js": true, "ts": true, "go": true, "py": true, "java": true,
		"c": true, "cpp": true, "h": true, "rb": true, "php": true,
	}
	
	archiveExts := map[string]bool{
		"zip": true, "rar": true, "7z": true, "tar": true, "gz": true,
		"bz2": true, "xz": true, "iso": true, "dmg": true,
	}
	
	if imageExts[ext] {
		return "image"
	}
	if videoExts[ext] {
		return "video"
	}
	if audioExts[ext] {
		return "audio"
	}
	if documentExts[ext] {
		return "document"
	}
	if archiveExts[ext] {
		return "archive"
	}
	
	return "other"
}
