package scan

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/steezeburger/storage-shower/backend/internal/fileinfo"
)

// Maximum number of previous scans to store
const MaxPreviousScans = 10

// ScanRecord represents a record of a previous scan
type ScanRecord struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	ResultID  string    `json:"resultId"`
	Size      int64     `json:"size"`
}

// ScanStatus contains information about a scan in progress
type ScanStatus struct {
	InProgress   bool    `json:"inProgress"`
	CurrentPath  string  `json:"currentPath"`
	ScannedItems int     `json:"scannedItems"`
	TotalItems   int     `json:"totalItems"`
	Progress     float64 `json:"progress"`
	Stalled      bool    `json:"stalled,omitempty"`
}

// Shared variables
var (
	// Track currently active scan
	scanStatus = ScanStatus{
		InProgress: false,
	}

	// List of previous scans
	PreviousScans []ScanRecord

	// Mutex for thread-safe access to scan status
	statusMutex sync.Mutex

	// Channel to signal scan cancellation
	cancelScan chan struct{}

	// Debug flag to control verbose logging
	DebugMode bool

	// To track stalled scans
	lastScannedItems int
	lastProgressTime time.Time

	// Current result path
	resultPath string
)

// ScanDirectory scans a directory and returns file information
func ScanDirectory(rootPath string, ignoreHidden bool) (fileinfo.FileInfo, error) {
	log.Printf("Beginning directory scan of: %s", rootPath)
	// Use filepath.Clean to normalize the path
	rootPath = filepath.Clean(rootPath)
	// Convert to absolute path to ensure consistency
	absPath, err := filepath.Abs(rootPath)
	if err == nil {
		rootPath = absPath
	}
	log.Printf("Normalized path: %s", rootPath)

	// Initialize scan status
	statusMutex.Lock()
	scanStatus = ScanStatus{
		InProgress:   true,
		CurrentPath:  rootPath,
		ScannedItems: 0,
		TotalItems:   1, // Start with at least 1 to avoid division by zero
		Progress:     0.0,
	}
	lastScannedItems = 0
	lastProgressTime = time.Now()
	statusMutex.Unlock()

	// Create a new cancel channel
	cancelScan = make(chan struct{})

	// Start counting files in a separate goroutine
	go countFiles(rootPath, ignoreHidden)

	// Get basic info about the root directory
	fileInfo, err := os.Stat(rootPath)
	if err != nil {
		statusMutex.Lock()
		scanStatus.InProgress = false
		statusMutex.Unlock()
		return fileinfo.FileInfo{}, fmt.Errorf("failed to access path: %v", err)
	}

	// Create the root file info
	root := fileinfo.FileInfo{
		Name:  filepath.Base(rootPath),
		Path:  rootPath,
		IsDir: fileInfo.IsDir(),
		Size:  fileInfo.Size(),
	}

	// Map to track directories by path
	dirMap := make(map[string]*fileinfo.FileInfo)
	dirMap[rootPath] = &root

	// Scan the directory structure recursively
	err = scanRecursive(rootPath, &root, dirMap, ignoreHidden)

	// Check if scan was canceled
	if err != nil && err.Error() == "scan canceled" {
		log.Printf("Scan was canceled")
		// Still save the partial result
		err = nil
	} else if err != nil {
		// On error, update status and return
		statusMutex.Lock()
		scanStatus.InProgress = false
		statusMutex.Unlock()
		return fileinfo.FileInfo{}, err
	}

	// Fix directory sizes
	if DebugMode {
		log.Println("Fixing directory sizes...")
	}
	fileinfo.FixDirectorySizes(&root, dirMap)

	// Trim the tree to reduce size before saving
	trimmedRoot := trimTreeForStorage(&root, 0)

	// Save result to a temporary file
	resultJSON, err := json.Marshal(trimmedRoot)
	if err != nil {
		log.Printf("Error marshaling result: %v", err)
		statusMutex.Lock()
		scanStatus.InProgress = false
		statusMutex.Unlock()
		return root, nil
	}

	tempFile, err := os.CreateTemp("", "storage-shower-*.json")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		statusMutex.Lock()
		scanStatus.InProgress = false
		statusMutex.Unlock()
		return root, nil
	}

	_, err = tempFile.Write(resultJSON)
	if err != nil {
		log.Printf("Error writing to temp file: %v", err)
		tempFile.Close()
		statusMutex.Lock()
		scanStatus.InProgress = false
		statusMutex.Unlock()
		return root, nil
	}

	tempFile.Close()

	// Generate a unique ID for this scan result
	resultID := filepath.Base(tempFile.Name())

	// Record this scan
	newScan := ScanRecord{
		Path:      rootPath,
		Timestamp: time.Now(),
		ResultID:  resultID,
		Size:      root.Size,
	}

	// Update global variables
	statusMutex.Lock()
	resultPath = tempFile.Name()
	PreviousScans = append([]ScanRecord{newScan}, PreviousScans...)
	if len(PreviousScans) > MaxPreviousScans {
		PreviousScans = PreviousScans[:MaxPreviousScans]
	}
	scanStatus.InProgress = false
	statusMutex.Unlock()

	// Save previous scans to persistent storage
	SavePreviousScans()

	log.Printf("Scan completed successfully for path: %s", rootPath)
	log.Printf("Scan result saved to %s (%s)", tempFile.Name(), fileinfo.FormatBytes(int64(len(resultJSON))))

	return root, nil
}

// countFiles counts files in a directory to provide progress information
func countFiles(rootPath string, ignoreHidden bool) {
	log.Printf("Starting file count for: %s", rootPath)
	var count int

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip items we can't access
		}

		// Check if we should cancel
		select {
		case <-cancelScan:
			return filepath.SkipAll
		default:
			// Continue with scan
		}

		if ignoreHidden && fileinfo.IsHidden(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		count++

		if count%1000 == 0 {
			log.Printf("Counted %d files so far...", count)
		}

		statusMutex.Lock()
		scanStatus.TotalItems = count
		statusMutex.Unlock()

		return nil
	})

	log.Printf("File count completed: %d total items found", count)
	statusMutex.Lock()
	scanStatus.TotalItems = count
	if scanStatus.TotalItems == 0 {
		scanStatus.TotalItems = 1 // Ensure we never have zero total items
	}
	statusMutex.Unlock()
}

// scanRecursive recursively scans a directory
func scanRecursive(path string, dir *fileinfo.FileInfo, dirMap map[string]*fileinfo.FileInfo, ignoreHidden bool) error {
	// Check for cancellation
	select {
	case <-cancelScan:
		return fmt.Errorf("scan canceled")
	default:
		// Continue with scan
	}

	// Update scan status
	statusMutex.Lock()
	scanStatus.CurrentPath = path
	scanStatus.ScannedItems++
	if scanStatus.TotalItems > 0 {
		scanStatus.Progress = float64(scanStatus.ScannedItems) / float64(scanStatus.TotalItems)
	} else {
		scanStatus.Progress = 0.0
	}

	// Check for stalled scan
	if scanStatus.ScannedItems > 0 {
		if scanStatus.ScannedItems > lastScannedItems {
			// Progress is being made, update the last known state
			lastScannedItems = scanStatus.ScannedItems
			lastProgressTime = time.Now()
			scanStatus.Stalled = false
		} else if time.Since(lastProgressTime) > 30*time.Second {
			// No progress for 30 seconds, consider scan stalled
			log.Printf("Scan appears stalled - no progress for 30 seconds")
			scanStatus.Stalled = true
		}
	}
	statusMutex.Unlock()

	// Only read the directory if it's a directory
	if !dir.IsDir {
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Warning: Cannot read directory %s: %v", path, err)
		return nil
	}

	// Process each entry in the directory
	for _, entry := range entries {
		// Check for cancellation again
		select {
		case <-cancelScan:
			return fmt.Errorf("scan canceled")
		default:
			// Continue with scan
		}

		entryName := entry.Name()
		entryPath := filepath.Join(path, entryName)

		// Skip hidden files/directories if requested
		if ignoreHidden && fileinfo.IsHidden(entryPath) {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			log.Printf("Warning: Cannot get info for %s: %v", entryPath, err)
			continue
		}

		// Extract extension for files
		extension := ""
		if !info.IsDir() {
			if ext := filepath.Ext(entryPath); ext != "" {
				extension = ext[1:] // Remove the leading dot
			}
		}

		// Create file info for this entry
		fileSize := info.Size()
		entryInfo := fileinfo.FileInfo{
			Name:      entryName,
			Path:      entryPath,
			Size:      fileSize,
			IsDir:     entry.IsDir(),
			Extension: extension,
		}

		// Add to parent's children
		dir.Children = append(dir.Children, entryInfo)

		// If it's a directory, add to dirMap and recursively scan
		if entry.IsDir() {
			// Store a reference to the directory in the map
			dirIndex := len(dir.Children) - 1
			dirMap[entryPath] = &dir.Children[dirIndex]

			// Recursively scan the directory
			err := scanRecursive(entryPath, &dir.Children[dirIndex], dirMap, ignoreHidden)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetScanStatus returns the current scan status
func GetScanStatus() ScanStatus {
	statusMutex.Lock()
	status := scanStatus
	// Ensure we never return NaN for progress
	if status.TotalItems == 0 {
		status.TotalItems = 1
		status.Progress = 0.0
	}
	statusMutex.Unlock()
	return status
}

// trimTreeForStorage reduces the size of the directory tree by limiting depth
// and trimming nodes with small sizes
func trimTreeForStorage(node *fileinfo.FileInfo, depth int) fileinfo.FileInfo {
	// Create a copy of the node
	result := *node

	// For deep trees, limit the depth to reduce JSON size
	maxDepth := 6
	minSizeToKeep := int64(1024 * 1024) // 1MB minimum size to keep at deeper levels

	// If we're at the max depth, only keep children above the size threshold
	if depth >= maxDepth {
		// For deep levels, only keep significant items
		if len(result.Children) > 0 {
			keptChildren := make([]fileinfo.FileInfo, 0)
			for _, child := range result.Children {
				if child.Size >= minSizeToKeep || child.IsDir && len(child.Children) > 0 {
					// For these deep nodes, don't include their children
					trimmedChild := child
					trimmedChild.Children = nil
					keptChildren = append(keptChildren, trimmedChild)
				}
			}

			// If we have too many children, keep only the largest ones
			maxChildren := 10
			if len(keptChildren) > maxChildren {
				// Sort by size, descending
				sort.Slice(keptChildren, func(i, j int) bool {
					return keptChildren[i].Size > keptChildren[j].Size
				})

				// Keep only the largest items
				keptChildren = keptChildren[:maxChildren]
			}

			result.Children = keptChildren
		}
		return result
	}

	// For normal depth, recursively process children
	if len(result.Children) > 0 {
		newChildren := make([]fileinfo.FileInfo, len(result.Children))
		for i, child := range result.Children {
			newChildren[i] = trimTreeForStorage(&child, depth+1)
		}
		result.Children = newChildren
	}

	return result
}

// CancelScan cancels any scan in progress
func CancelScan() string {
	statusMutex.Lock()
	inProgress := scanStatus.InProgress
	statusMutex.Unlock()

	if inProgress {
		// Signal cancellation
		close(cancelScan)
		return "stopping"
	}

	return "not_running"
}

// GetLatestScanResult returns the most recent scan result
func GetLatestScanResult() (fileinfo.FileInfo, error) {
	statusMutex.Lock()
	path := resultPath
	scans := PreviousScans
	statusMutex.Unlock()

	if path == "" || len(scans) == 0 {
		return fileinfo.FileInfo{}, fmt.Errorf("no scan results available")
	}

	// Read the result file
	data, err := os.ReadFile(path)
	if err != nil {
		return fileinfo.FileInfo{}, fmt.Errorf("failed to read scan result: %v", err)
	}

	// Parse the result
	var result fileinfo.FileInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return fileinfo.FileInfo{}, fmt.Errorf("failed to parse scan result: %v", err)
	}

	return result, nil
}

// GetScanResultByID returns a specific scan result by ID
func GetScanResultByID(resultID string) (fileinfo.FileInfo, error) {
	statusMutex.Lock()
	scans := PreviousScans
	statusMutex.Unlock()

	// Find the scan record
	found := false
	for _, scan := range scans {
		if scan.ResultID == resultID {
			found = true
			break
		}
	}

	if !found {
		return fileinfo.FileInfo{}, fmt.Errorf("scan result with ID %s not found", resultID)
	}

	// Extract the full path from the ResultID
	resultPath := filepath.Join(os.TempDir(), resultID)

	// Read the result file
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return fileinfo.FileInfo{}, fmt.Errorf("failed to read scan result: %v", err)
	}

	// Parse the result
	var result fileinfo.FileInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return fileinfo.FileInfo{}, fmt.Errorf("failed to parse scan result: %v", err)
	}

	return result, nil
}

// SavePreviousScans saves the list of previous scans to a file
func SavePreviousScans() {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Cannot get home directory: %v", err)
		return
	}

	// Create storage-shower directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".storage-shower")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("Warning: Cannot create config directory: %v", err)
		return
	}

	// Save to file
	scansFile := filepath.Join(configDir, "previous-scans.json")

	statusMutex.Lock()
	data, err := json.MarshalIndent(PreviousScans, "", "  ")
	statusMutex.Unlock()

	if err != nil {
		log.Printf("Warning: Cannot marshal scan history: %v", err)
		return
	}

	if err := os.WriteFile(scansFile, data, 0644); err != nil {
		log.Printf("Warning: Cannot save scan history: %v", err)
	}
}

// LoadPreviousScans loads the list of previous scans from a file
func LoadPreviousScans() {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Cannot get home directory: %v", err)
		return
	}

	// Check if file exists
	scansFile := filepath.Join(homeDir, ".storage-shower", "previous-scans.json")
	data, err := os.ReadFile(scansFile)
	if err != nil {
		// This is not an error, file might not exist yet
		return
	}

	// Parse the file
	var scans []ScanRecord
	if err := json.Unmarshal(data, &scans); err != nil {
		log.Printf("Warning: Cannot parse scan history: %v", err)
		return
	}

	// Update global variable
	statusMutex.Lock()
	PreviousScans = scans
	statusMutex.Unlock()
}
