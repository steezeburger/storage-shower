package main

import (
	"testing"
	"path/filepath"
	"strings"
)

func TestFileInfo_Size(t *testing.T) {
	// Create a test file info structure
	root := FileInfo{
		Name:  "root",
		Path:  "/test/root",
		IsDir: true,
		Size:  0,
		Children: []FileInfo{
			{
				Name:  "file1.txt",
				Path:  "/test/root/file1.txt",
				IsDir: false,
				Size:  100,
			},
			{
				Name:  "file2.txt",
				Path:  "/test/root/file2.txt",
				IsDir: false,
				Size:  200,
			},
			{
				Name:  "subdir",
				Path:  "/test/root/subdir",
				IsDir: true,
				Size:  0,
				Children: []FileInfo{
					{
						Name:  "file3.txt",
						Path:  "/test/root/subdir/file3.txt",
						IsDir: false,
						Size:  300,
					},
				},
			},
		},
	}

	// Create dirMap to help fixDirectorySizes function find subdirectories
	dirMap := make(map[string]*FileInfo)
	
	// Add subdir to dirMap
	for i := range root.Children {
		if root.Children[i].IsDir {
			dirMap[root.Children[i].Path] = &root.Children[i]
		}
	}
	
	// Fix directory sizes
	fixDirectorySizes(&root, dirMap)

	// Test the root size (should be sum of all children)
	expectedSize := int64(600) // 100 + 200 + 300
	if root.Size != expectedSize {
		t.Errorf("Root size incorrect, got: %d, want: %d", root.Size, expectedSize)
	}

	// Test subdirectory size
	var subdir *FileInfo
	for i := range root.Children {
		if root.Children[i].Name == "subdir" {
			subdir = &root.Children[i]
			break
		}
	}

	if subdir == nil {
		t.Fatal("Subdir not found in children")
	}

	expectedSubdirSize := int64(300)
	if subdir.Size != expectedSubdirSize {
		t.Errorf("Subdir size incorrect, got: %d, want: %d", subdir.Size, expectedSubdirSize)
	}
}

func TestIsHidden(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/test/normal.txt", false},
		{"/test/.hidden.txt", true},
		{"/test/.hidden/file.txt", false},
		{"/test/.hidden/", false},
		{"/test/.", false},
		{"/test/..", false},
	}

	for _, test := range tests {
		result := isHidden(test.path)
		if result != test.expected {
			t.Errorf("isHidden(%s) = %v, want %v", test.path, result, test.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, test := range tests {
		result := formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", test.bytes, result, test.expected)
		}
	}
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		path           string
		expectedExt    string
	}{
		{"/test/file.txt", "txt"},
		{"/test/file.tar.gz", "gz"},
		{"/test/file", ""},
		{"/test/.hidden", ""}, // Hidden files should have no extension
		{"/test/image.jpg", "jpg"},
		{"/test/script.sh", "sh"},
	}

	for _, test := range tests {
		fileInfo := FileInfo{
			Name: filepath.Base(test.path),
			Path: test.path,
		}
		
		for _, test := range tests {
			// Extract extension using the same logic as in the main code
			extension := ""
			// Skip extension extraction for hidden files
			filename := filepath.Base(test.path)
			if !strings.HasPrefix(filename, ".") || filename == "." || filename == ".." {
				if ext := filepath.Ext(test.path); ext != "" {
					extension = ext[1:] // Remove the leading dot
				}
			}
		
			if extension != test.expectedExt {
			t.Errorf("File extension for %s = %s, want %s", test.path, extension, test.expectedExt)
		}
	}
}
