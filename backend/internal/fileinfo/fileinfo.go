package fileinfo

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

// FileTypeStats represents statistics about file types in a directory
type FileTypeStats struct {
	Image    int64 `json:"image"`
	Video    int64 `json:"video"`
	Audio    int64 `json:"audio"`
	Document int64 `json:"document"`
	Archive  int64 `json:"archive"`
	Other    int64 `json:"other"`
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Size      int64          `json:"size"`
	IsDir     bool           `json:"isDir"`
	Children  []FileInfo     `json:"children,omitempty"`
	Extension string         `json:"extension,omitempty"`
	FileTypes *FileTypeStats `json:"fileTypes,omitempty"`
}

// FixDirectorySizes updates directory sizes based on their children
func FixDirectorySizes(dir *FileInfo, dirMap map[string]*FileInfo) int64 {
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
				childSize = FixDirectorySizes(childDir, dirMap)
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
			fileType := GetFileType(dir.Children[i].Extension)
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

// IsHidden determines if a file is hidden (starts with a dot)
func IsHidden(path string) bool {
	name := filepath.Base(path)
	// Only consider a file hidden if it starts with a dot and isn't a directory ending with /
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		return false
	}
	return strings.HasPrefix(name, ".") && name != "." && name != ".."
}

// GetFileType categorizes a file based on its extension
func GetFileType(extension string) string {
	if extension == "" {
		return "other"
	}

	// Convert extension to lowercase for case-insensitive matching
	ext := strings.ToLower(extension)

	// Handle compound extensions like tar.gz
	if strings.Contains(ext, ".") {
		parts := strings.Split(ext, ".")
		lastPart := parts[len(parts)-1]
		
		// Special case for common archive formats
		if lastPart == "gz" || lastPart == "bz2" || lastPart == "xz" {
			return "archive"
		}
	}

	// Image formats
	imageFormats := []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp", "svg", "ico", "heic", "heif"}
	for _, format := range imageFormats {
		if ext == format {
			return "image"
		}
	}

	// Video formats
	videoFormats := []string{"mp4", "avi", "mov", "wmv", "flv", "mkv", "webm", "m4v", "mpg", "mpeg", "3gp"}
	for _, format := range videoFormats {
		if ext == format {
			return "video"
		}
	}

	// Audio formats
	audioFormats := []string{"mp3", "wav", "ogg", "flac", "aac", "wma", "m4a", "opus"}
	for _, format := range audioFormats {
		if ext == format {
			return "audio"
		}
	}

	// Document formats
	documentFormats := []string{
		"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", 
		"txt", "rtf", "odt", "ods", "odp", "md", "csv",
		"pages", "numbers", "key", "html", "htm", "xml", "json",
	}
	for _, format := range documentFormats {
		if ext == format {
			return "document"
		}
	}

	// Archive formats
	archiveFormats := []string{"zip", "rar", "7z", "tar", "gz", "bz2", "xz", "iso", "dmg"}
	for _, format := range archiveFormats {
		if ext == format {
			return "archive"
		}
	}

	// Default to "other" for unrecognized extensions
	return "other"
}

// FormatBytes converts a byte count to a human-readable string
func FormatBytes(bytes int64) string {
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

