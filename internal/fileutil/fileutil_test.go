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
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "normal file",
			path:     "/path/to/file.txt",
			expected: false,
		},
		{
			name:     "hidden file",
			path:     "/path/to/.hidden",
			expected: true,
		},
		{
			name:     "current directory",
			path:     ".",
			expected: false,
		},
		{
			name:     "hidden directory",
			path:     "/path/.hidden/file.txt",
			expected: true,
		},
		{
			name:     "node_modules directory",
			path:     "/project/node_modules/package/file.js",
			expected: true,
		},
		{
			name:     "git directory",
			path:     "/project/.git/config",
			expected: true,
		},
		{
			name:     "dist directory",
			path:     "/project/dist/main.js",
			expected: true,
		},
		{
			name:     "build directory",
			path:     "/project/build/output.o",
			expected: true,
		},
		{
			name:     "pycache directory",
			path:     "/project/__pycache__/module.pyc",
			expected: true,
		},
		{
			name:     "opencode directory",
			path:     "/project/.opencode/config.json",
			expected: true,
		},
		{
			name:     "vendor directory",
			path:     "/project/vendor/package/lib.go",
			expected: true,
		},
		{
			name:     "target directory",
			path:     "/project/target/release/binary",
			expected: true,
		},
		{
			name:     "vscode directory",
			path:     "/project/.vscode/settings.json",
			expected: true,
		},
		{
			name:     "idea directory",
			path:     "/project/.idea/workspace.xml",
			expected: true,
		},
		{
			name:     "normal nested file",
			path:     "/project/src/main/java/Main.java",
			expected: false,
		},
		{
			name:     "coverage directory",
			path:     "/project/coverage/report.html",
			expected: true,
		},
		{
			name:     "tmp directory",
			path:     "/project/tmp/temp_file.txt",
			expected: true,
		},
		{
			name:     "logs directory",
			path:     "/project/logs/app.log",
			expected: true,
		},
		{
			name:     "generated directory",
			path:     "/project/generated/proto.pb.go",
			expected: true,
		},
		{
			name:     "bower_components directory",
			path:     "/project/bower_components/jquery/jquery.js",
			expected: true,
		},
		{
			name:     "jspm_packages directory",
			path:     "/project/jspm_packages/npm/lodash.js",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SkipHidden(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGlobWithDoublestar(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test file structure
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"dir1/file3.txt",
		"dir1/file4.go",
		"dir1/subdir/file5.txt",
		"dir2/file6.js",
		".hidden/file7.txt",
		"node_modules/package.json",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	t.Run("matches all files with **", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**", tempDir, 0)
		require.NoError(t, err)
		assert.False(t, truncated)

		// Should find visible files, excluding hidden and ignored directories
		expectedFiles := []string{
			"file1.txt",
			"file2.go",
			"dir1/file3.txt",
			"dir1/file4.go",
			"dir1/subdir/file5.txt",
			"dir2/file6.js",
		}

		assert.GreaterOrEqual(t, len(matches), len(expectedFiles))

		// Check that hidden files are not included
		for _, match := range matches {
			assert.NotContains(t, match, ".hidden")
			assert.NotContains(t, match, "node_modules")
		}
	})

	t.Run("matches specific file pattern", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**/*.txt", tempDir, 0)
		require.NoError(t, err)
		assert.False(t, truncated)

		// Should only find .txt files
		for _, match := range matches {
			assert.True(t, filepath.Ext(match) == ".txt")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**", tempDir, 2)
		require.NoError(t, err)
		assert.True(t, truncated)
		assert.Equal(t, 2, len(matches))
	})

	t.Run("returns empty for non-existent pattern", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**/*.nonexistent", tempDir, 0)
		require.NoError(t, err)
		assert.False(t, truncated)
		assert.Empty(t, matches)
	})

	t.Run("handles invalid directory", func(t *testing.T) {
		_, _, err := GlobWithDoublestar("**", "/nonexistent/directory", 0)
		assert.Error(t, err)
	})
}

func TestFileInfoSorting(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files with different modification times
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	file3 := filepath.Join(tempDir, "file3.txt")

	// Create files with specific content
	err := os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = os.WriteFile(file3, []byte("content3"), 0644)
	require.NoError(t, err)

	// Test that files are sorted by modification time (newest first)
	matches, truncated, err := GlobWithDoublestar("*.txt", tempDir, 0)
	require.NoError(t, err)
	assert.False(t, truncated)
	assert.Len(t, matches, 3)

	// file3 should be first (newest), file1 should be last (oldest)
	assert.Contains(t, matches[0], "file3.txt")
	assert.Contains(t, matches[2], "file1.txt")
}

func TestGetRgCmd(t *testing.T) {
	t.Run("creates command with glob pattern", func(t *testing.T) {
		cmd := GetRgCmd("*.go")
		if cmd == nil {
			t.Skip("ripgrep not available")
		}

		args := cmd.Args
		assert.Contains(t, args, "--files")
		assert.Contains(t, args, "-L")
		assert.Contains(t, args, "--null")
		assert.Contains(t, args, "--glob")
		assert.Contains(t, args, "/*.go")
	})

	t.Run("creates command without glob pattern", func(t *testing.T) {
		cmd := GetRgCmd("")
		if cmd == nil {
			t.Skip("ripgrep not available")
		}

		args := cmd.Args
		assert.Contains(t, args, "--files")
		assert.Contains(t, args, "-L")
		assert.Contains(t, args, "--null")
		assert.NotContains(t, args, "--glob")
	})

	t.Run("handles absolute glob pattern", func(t *testing.T) {
		cmd := GetRgCmd("/absolute/path/*.go")
		if cmd == nil {
			t.Skip("ripgrep not available")
		}

		args := cmd.Args
		assert.Contains(t, args, "--glob")
		assert.Contains(t, args, "/absolute/path/*.go")
	})
}

func TestGetFzfCmd(t *testing.T) {
	t.Run("creates fzf command with query", func(t *testing.T) {
		cmd := GetFzfCmd("test")
		if cmd == nil {
			t.Skip("fzf not available")
		}

		args := cmd.Args
		assert.Contains(t, args, "--filter")
		assert.Contains(t, args, "test")
		assert.Contains(t, args, "--read0")
		assert.Contains(t, args, "--print0")
	})

	t.Run("handles empty query", func(t *testing.T) {
		cmd := GetFzfCmd("")
		if cmd == nil {
			t.Skip("fzf not available")
		}

		args := cmd.Args
		assert.Contains(t, args, "--filter")
		assert.Contains(t, args, "")
	})
}

func TestPatternEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"test.go",
		"test.txt",
		"subdir/test.go",
		"subdir/other.py",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	t.Run("matches with leading slash", func(t *testing.T) {
		matches, _, err := GlobWithDoublestar("/**/*.go", tempDir, 0)
		require.NoError(t, err)

		goFiles := 0
		for _, match := range matches {
			if filepath.Ext(match) == ".go" {
				goFiles++
			}
		}
		assert.Equal(t, 2, goFiles)
	})

	t.Run("matches without leading slash", func(t *testing.T) {
		matches, _, err := GlobWithDoublestar("**/*.go", tempDir, 0)
		require.NoError(t, err)

		goFiles := 0
		for _, match := range matches {
			if filepath.Ext(match) == ".go" {
				goFiles++
			}
		}
		assert.Equal(t, 2, goFiles)
	})
}
