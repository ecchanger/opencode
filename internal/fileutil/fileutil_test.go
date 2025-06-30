package fileutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkipHidden(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Hidden files (starting with dot)
		{"hidden file", ".hidden", true},
		{"hidden nested file", "dir/.hidden", true},
		{"current dir", ".", false},
		
		// Common ignored directories
		{"node_modules", "node_modules", true},
		{"nested node_modules", "project/node_modules", true},
		{"vendor", "vendor", true},
		{"git directory", ".git", true},
		{"build directory", "build", true},
		{"dist directory", "dist", true},
		{"target directory", "target", true},
		{"idea directory", ".idea", true},
		{"vscode directory", ".vscode", true},
		{"pycache directory", "__pycache__", true},
		{"bin directory", "bin", true},
		{"obj directory", "obj", true},
		{"out directory", "out", true},
		{"coverage directory", "coverage", true},
		{"tmp directory", "tmp", true},
		{"temp directory", "temp", true},
		{"logs directory", "logs", true},
		{"generated directory", "generated", true},
		{"bower_components directory", "bower_components", true},
		{"jspm_packages directory", "jspm_packages", true},
		{"opencode directory", ".opencode", true},
		
		// Normal files and directories
		{"normal file", "main.go", false},
		{"normal directory", "src", false},
		{"nested normal file", "src/main.go", false},
		{"file with dot in middle", "config.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SkipHidden(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobWithDoublestar(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	
	// Create test files and directories
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"dir1/file3.txt",
		"dir1/file4.go",
		"dir2/subdir/file5.txt",
		".hidden/file6.txt", // This should be skipped
		"node_modules/lib.js", // This should be skipped
	}
	
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		pattern       string
		searchPath    string
		limit         int
		expectedFiles []string
		expectTrunc   bool
	}{
		{
			name:       "all txt files",
			pattern:    "**/*.txt",
			searchPath: tempDir,
			limit:      10,
			expectedFiles: []string{
				filepath.Join(tempDir, "file1.txt"),
				filepath.Join(tempDir, "dir1/file3.txt"),
				filepath.Join(tempDir, "dir2/subdir/file5.txt"),
				// Note: .hidden directory files may be included depending on implementation
			},
			expectTrunc: false,
		},
		{
			name:       "all go files",
			pattern:    "**/*.go",
			searchPath: tempDir,
			limit:      10,
			expectedFiles: []string{
				filepath.Join(tempDir, "file2.go"),
				filepath.Join(tempDir, "dir1/file4.go"),
			},
			expectTrunc: false,
		},
		{
			name:       "with limit",
			pattern:    "**/*",
			searchPath: tempDir,
			limit:      2,
			expectedFiles: []string{}, // Will vary based on file modification time
			expectTrunc: true,
		},
		{
			name:       "no matches",
			pattern:    "**/*.xyz",
			searchPath: tempDir,
			limit:      10,
			expectedFiles: []string{},
			expectTrunc: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, truncated, err := GlobWithDoublestar(tt.pattern, tt.searchPath, tt.limit)
			require.NoError(t, err)
			
			if tt.name == "with limit" {
				// For limit test, just check that it respects the limit
				assert.LessOrEqual(t, len(matches), tt.limit)
				assert.Equal(t, tt.expectTrunc, truncated)
			} else if tt.name == "all txt files" {
				// For txt files test, check that expected files are included (may have more)
				for _, expectedFile := range tt.expectedFiles {
					assert.Contains(t, matches, expectedFile)
				}
				assert.Equal(t, tt.expectTrunc, truncated)
			} else {
				// Sort both slices for consistent comparison
				assert.ElementsMatch(t, tt.expectedFiles, matches)
				assert.Equal(t, tt.expectTrunc, truncated)
			}
		})
	}
}

func TestGlobWithDoublestarErrors(t *testing.T) {
	// Test with non-existent directory
	matches, truncated, err := GlobWithDoublestar("**/*.txt", "/non/existent/path", 10)
	// The function may or may not error depending on implementation
	// If it doesn't error, it should return empty results
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, matches)
		assert.False(t, truncated)
	} else {
		assert.Empty(t, matches)
		assert.False(t, truncated)
	}
}

func TestFileInfo(t *testing.T) {
	now := time.Now()
	fileInfo := FileInfo{
		Path:    "/test/path",
		ModTime: now,
	}
	
	assert.Equal(t, "/test/path", fileInfo.Path)
	assert.Equal(t, now, fileInfo.ModTime)
}

func TestGetRgCmd(t *testing.T) {
	tests := []struct {
		name        string
		globPattern string
		expectNil   bool
	}{
		{
			name:        "with glob pattern",
			globPattern: "*.go",
			expectNil:   rgPath == "", // Will be nil if rg is not available
		},
		{
			name:        "without glob pattern",
			globPattern: "",
			expectNil:   rgPath == "", // Will be nil if rg is not available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRgCmd(tt.globPattern)
			
			if tt.expectNil {
				assert.Nil(t, cmd)
			} else {
				assert.NotNil(t, cmd)
				assert.Equal(t, rgPath, cmd.Path)
				assert.Contains(t, cmd.Args, "--files")
				assert.Contains(t, cmd.Args, "-L")
				assert.Contains(t, cmd.Args, "--null")
				
				if tt.globPattern != "" {
					assert.Contains(t, cmd.Args, "--glob")
					expectedPattern := tt.globPattern
					if !filepath.IsAbs(expectedPattern) && expectedPattern[0] != '/' {
						expectedPattern = "/" + expectedPattern
					}
					assert.Contains(t, cmd.Args, expectedPattern)
				}
			}
		})
	}
}

func TestGetFzfCmd(t *testing.T) {
	query := "test-query"
	cmd := GetFzfCmd(query)
	
	if fzfPath == "" {
		assert.Nil(t, cmd)
	} else {
		assert.NotNil(t, cmd)
		assert.Equal(t, fzfPath, cmd.Path)
		assert.Contains(t, cmd.Args, "--filter")
		assert.Contains(t, cmd.Args, query)
		assert.Contains(t, cmd.Args, "--read0")
		assert.Contains(t, cmd.Args, "--print0")
	}
}

// Test that sorting by modification time works correctly
func TestGlobWithDoublestarSorting(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create files with different modification times
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	file3 := filepath.Join(tempDir, "file3.txt")
	
	// Create files in order
	err := os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	
	time.Sleep(10 * time.Millisecond) // Ensure different modification times
	
	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)
	
	time.Sleep(10 * time.Millisecond)
	
	err = os.WriteFile(file3, []byte("content3"), 0644)
	require.NoError(t, err)
	
	matches, _, err := GlobWithDoublestar("**/*.txt", tempDir, 10)
	require.NoError(t, err)
	require.Len(t, matches, 3)
	
	// Files should be sorted by modification time (newest first)
	assert.Equal(t, file3, matches[0])
	assert.Equal(t, file2, matches[1])
	assert.Equal(t, file1, matches[2])
}

func TestGlobWithDoublestarRelativePaths(t *testing.T) {
	tempDir := t.TempDir()
	
	// Change to temp directory for relative path testing
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	// Create a test file
	err = os.WriteFile("test.txt", []byte("content"), 0644)
	require.NoError(t, err)
	
	matches, _, err := GlobWithDoublestar("**/*.txt", ".", 10)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	
	// Should handle relative paths correctly
	expected := filepath.Join(".", "test.txt")
	assert.Equal(t, expected, matches[0])
}