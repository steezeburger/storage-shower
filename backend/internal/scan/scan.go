package scan

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/steezeburger/storage-shower/backend/internal/fileinfo"
	"github.com/steezeburger/storage-shower/backend/pkg/utils"
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
	InProgress       bool      `json:"inProgress"`
	CurrentDirectory string    `json:"currentDirectory"`
	FilesScanned     int       `json:"filesScanned"`
	StartTime        time.Time `json:"startTime"`
	LastUpdate       time.Time `json:"lastUpdate"`
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
		InProgress:       true,
		CurrentDirectory: rootPath,
		FilesScanned:     0,
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
	}
	statusMutex.Unlock()

	// Create a new cancel channel
	cancelScan = make(chan struct{})

	// Map to track directories by path
	dirMap := make(map[string]*fileinfo.FileInfo)

	// Get basic info about the root directory
	fileInfo, err := os.Stat(rootPath)
	if err != nil {
		return fileinfo.FileInfo{}, fmt.Errorf("failed to access path: %v", err)
	}

	// Create the root file info
	root := fileinfo.FileInfo{
		Name:  filepath.Base(rootPath),
		Path:  rootPath,
		IsDir: fileInfo.IsDir(),
		Size:  fileInfo.Size(),
	}

	// Initialize a stall detector
	detector := utils.NewStallDetector()

	// Scan the directory structure recursively
	scanResult, err := scanRecursive(rootPath, &root, 0, ignoreHidden, dirMap, detector)
	if err != nil {
		// Clean up scan status on error
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

	// Generate a unique ID for this scan result
	resultID := generateResultID()

	// Record this scan
	if len(root.Path) > 0 {
		newScan := ScanRecord{
			Path:      root.Path,
			Timestamp: time.Now(),
			ResultID:  resultID,
			Size:      root.Size,
		}

		// Add to the beginning of the list
		PreviousScans = append([]ScanRecord{newScan}, PreviousScans...)

		// Trim list if needed
		if len(PreviousScans) > MaxPreviousScans {
			PreviousScans = PreviousScans[:MaxPreviousScans]
		}

		// Save previous scans to persistent storage
		savePreviousScans()
	}

	// Update scan status to completed
	statusMutex.Lock()
	scanStatus.InProgress = false
	statusMutex.Unlock()

	// Return the scan result
	return scanResult, nil
}

// Recursive function to scan directories
func scanRecursive(rootPath string, root *fileinfo.FileInfo, depth int, ignoreHidden bool, dirMap map[string]*fileinfo.FileInfo, detector *utils.StallDetector) (fileinfo.FileInfo, error) {
	if depth == 0 {
		// Store the root directory in dirMap
		dirMap[root.Path] = root
	}

	// Update scan status
	statusMutex.Lock()
	scanStatus.CurrentDirectory = rootPath
	scanStatus.FilesScanned++
	scanStatus.LastUpdate = time.Now()
	filesScanned := scanStatus.FilesScanned
	statusMutex.Unlock()

	// Update stall detector
	detector.UpdateActivity(filesScanned)

	// Check for scan cancellation
	select {
	case <-cancelScan:
		return fileinfo.FileInfo{}, fmt.Errorf("scan cancelled")
	default:
		// Continue with scan
	}

	// Check for stall
	if detector.IsStalled() {
		log.Printf("Warning: Scan appears to be stalled in directory: %s", rootPath)
		log.Printf("Files scanned: %d, Last activity: %v", filesScanned, detector.GetLastActivityTime())
	}

	// Only read the directory if it's a directory
	if !root.IsDir {
		return *root, nil
	}

	// Read directory contents
	dirEntries, err := os.ReadDir(rootPath)
	if err != nil {
		// If we can't read the directory, log it but continue
		log.Printf("Warning: Cannot read directory %s: %v", rootPath, err)
		return *root, nil
	}

	// Process each entry in the directory
	for _, entry := range dirEntries {
		// Skip hidden files/directories if requested
		entryPath := filepath.Join(rootPath, entry.Name())
		isHiddenEntry := fileinfo.IsHidden(entryPath)
		if ignoreHidden && isHiddenEntry {
			if DebugMode {
				log.Printf("Skipping hidden: %s", entryPath)
			}
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
		entryInfo := fileinfo.FileInfo{
			Name:      entry.Name(),
			Path:      entryPath,
			Size:      info.Size(),
			IsDir:     entry.IsDir(),
			Extension: extension,
		}

		// If it's a directory, recursively scan it
		if entry.IsDir() {
			// Add this directory to the dirMap before recursion
			dirMap[entryPath] = &entryInfo

			// Recursively scan this directory
			_, err := scanRecursive(entryPath, &entryInfo, depth+1, ignoreHidden, dirMap, detector)
			if err != nil {
				// On error, still add partial results but log the error
				log.Printf("Warning: Error scanning subdirectory %s: %v", entryPath, err)
			}
		}

		// Add this entry to the parent's children
		root.Children = append(root.Children, entryInfo)
	}

	return *root, nil
}

// GetScanStatus returns the current scan status
func GetScanStatus() ScanStatus {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	return scanStatus
}

// CancelScan cancels any scan in progress
func CancelScan() bool {
	statusMutex.Lock()
	inProgress := scanStatus.InProgress
	statusMutex.Unlock()

	if inProgress {
		// Signal cancellation
		close(cancelScan)

		// Update status
		statusMutex.Lock()
		scanStatus.InProgress = false
		statusMutex.Unlock()

		return true
	}

	return false
}

// generateResultID creates a unique ID for scan results
func generateResultID() string {
	// Create a random ID with timestamp
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("scan_%d_%d", time.Now().Unix(), rand.Intn(10000))
}

// savePreviousScans saves the list of previous scans to a file
func savePreviousScans() {
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
	data, err := json.MarshalIndent(PreviousScans, "", "  ")
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
	PreviousScans = scans
}
