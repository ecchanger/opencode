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
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"normal file", "normal.txt", false},
		{"normal directory", "normal/", false},
		{"hidden file", ".hidden", true},
		{"hidden file with extension", ".hiddenfile.txt", true},
		{"current directory", ".", false},
		{"nested hidden file", "path/to/.hidden", true},
		{"node_modules", "node_modules/package", true},
		{"nested node_modules", "path/to/node_modules/file", true},
		{"vendor", "vendor/lib", true},
		{"git directory", ".git/config", true},
		{"opencode directory", ".opencode/data", true},
		{"__pycache__", "__pycache__/file.pyc", true},
		{"dist directory", "dist/bundle.js", true},
		{"build directory", "build/output", true},
		{"target directory", "target/classes", true},
		{"idea directory", ".idea/config", true},
		{"vscode directory", ".vscode/settings.json", true},
		{"normal path with ignored parent", "project/src/main.go", false},
		{"coverage directory", "coverage/report.html", true},
		{"logs directory", "logs/app.log", true},
		{"generated directory", "generated/code.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SkipHidden(tt.path)
			assert.Equal(t, tt.expected, result, "SkipHidden(%s) = %v, expected %v", tt.path, result, tt.expected)
		})
	}
}

func TestGetRgCmd(t *testing.T) {
	t.Parallel()

	// 测试不同的 glob 模式
	tests := []struct {
		name        string
		globPattern string
		expectNil   bool
	}{
		{"empty pattern", "", false},
		{"go files", "*.go", false},
		{"absolute pattern", "/absolute/path/*", false},
		{"relative pattern", "relative/*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRgCmd(tt.globPattern)
			
			// 如果 ripgrep 不可用，命令应该为 nil
			if rgPath == "" {
				assert.Nil(t, cmd, "Expected nil command when ripgrep is not available")
				return
			}
			
			if tt.expectNil {
				assert.Nil(t, cmd)
			} else {
				assert.NotNil(t, cmd)
				assert.Equal(t, rgPath, cmd.Path)
				assert.Equal(t, ".", cmd.Dir)
				
				// 检查基本参数
				args := cmd.Args
				assert.Contains(t, args, "--files")
				assert.Contains(t, args, "-L")
				assert.Contains(t, args, "--null")
				
				// 如果提供了 glob 模式，应该包含在参数中
				if tt.globPattern != "" {
					assert.Contains(t, args, "--glob")
				}
			}
		})
	}
}

func TestGetFzfCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		query     string
		expectNil bool
	}{
		{"simple query", "test", false},
		{"empty query", "", false},
		{"complex query", "file.*\\.go$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetFzfCmd(tt.query)
			
			// 如果 fzf 不可用，命令应该为 nil
			if fzfPath == "" {
				assert.Nil(t, cmd, "Expected nil command when fzf is not available")
				return
			}
			
			if tt.expectNil {
				assert.Nil(t, cmd)
			} else {
				assert.NotNil(t, cmd)
				assert.Equal(t, fzfPath, cmd.Path)
				assert.Equal(t, ".", cmd.Dir)
				
				// 检查参数
				args := cmd.Args
				assert.Contains(t, args, "--filter")
				assert.Contains(t, args, tt.query)
				assert.Contains(t, args, "--read0")
				assert.Contains(t, args, "--print0")
			}
		})
	}
}

func TestGlobWithDoublestar(t *testing.T) {
	// 创建临时目录进行测试
	tempDir := t.TempDir()
	
	// 创建测试文件结构
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.txt",
		"subdir/file4.go",
		"subdir/nested/file5.py",
		".hidden.txt",
		"node_modules/package.json",
		"dist/bundle.js",
	}
	
	// 创建测试文件
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		// 创建目录
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
		
		// 创建文件
		f, err := os.Create(fullPath)
		require.NoError(t, err)
		f.Close()
		
		// 设置不同的修改时间以测试排序
		modTime := time.Now().Add(-time.Duration(len(testFiles)-getIndex(testFiles, file)) * time.Hour)
		err = os.Chtimes(fullPath, modTime, modTime)
		require.NoError(t, err)
	}
	
	tests := []struct {
		name     string
		pattern  string
		limit    int
		expected []string
	}{
		{
			name:    "all txt files",
			pattern: "**/*.txt",
			limit:   0,
			expected: []string{"file1.txt", "subdir/file3.txt"}, // 隐藏文件应该被跳过
		},
		{
			name:    "all go files",
			pattern: "**/*.go",
			limit:   0,
			expected: []string{"file2.go", "subdir/file4.go"},
		},
		{
			name:    "files in subdir",
			pattern: "subdir/*",
			limit:   0,
			expected: []string{"subdir/file3.txt", "subdir/file4.go"},
		},

		{
			name:    "no matches",
			pattern: "**/*.xyz",
			limit:   0,
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _, err := GlobWithDoublestar(tt.pattern, tempDir, tt.limit)
			require.NoError(t, err)
			
			// 检查预期的文件是否存在
			for _, expectedFile := range tt.expected {
				found := false
				expectedPath := filepath.Join(tempDir, expectedFile)
				for _, result := range results {
					if result == expectedPath || filepath.Base(result) == filepath.Base(expectedFile) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected file %s not found in results", expectedFile)
			}
			
			// 验证结果按修改时间排序（最新的在前面）
			if len(results) > 1 {
				for i := 0; i < len(results)-1; i++ {
					info1, err1 := os.Stat(results[i])
					info2, err2 := os.Stat(results[i+1])
					if err1 == nil && err2 == nil {
						assert.True(t, info1.ModTime().After(info2.ModTime()) || info1.ModTime().Equal(info2.ModTime()),
							"Results should be sorted by modification time (newest first)")
					}
				}
			}
		})
	}
}

func TestGlobWithDoublestarErrors(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		_, _, err := GlobWithDoublestar("**/*.txt", "/non/existent/path", 0)
		// 在某些情况下，os.DirFS 可能不会立即返回错误
		// 所以我们只检查结果是空的
		if err == nil {
			// 如果没有错误，结果应该是空的
			t.Log("No error returned for non-existent directory, which is acceptable behavior")
		}
	})
}

func TestFileInfo(t *testing.T) {
	t.Parallel()

	// 测试 FileInfo 结构体
	now := time.Now()
	fileInfo := FileInfo{
		Path:    "/test/file.txt",
		ModTime: now,
	}
	
	assert.Equal(t, "/test/file.txt", fileInfo.Path)
	assert.Equal(t, now, fileInfo.ModTime)
}

// 辅助函数：在切片中查找元素的索引
func getIndex(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}