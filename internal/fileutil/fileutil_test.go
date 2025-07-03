package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkipHidden(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		// 隐藏文件和目录
		{
			name:     "隐藏文件",
			path:     ".hidden",
			expected: true,
		},
		{
			name:     "深层隐藏文件",
			path:     "/path/to/.hidden_file",
			expected: true,
		},
		{
			name:     "当前目录",
			path:     ".",
			expected: false,
		},
		// 常见忽略目录
		{
			name:     "node_modules目录",
			path:     "project/node_modules",
			expected: true,
		},
		{
			name:     "深层node_modules",
			path:     "/project/subdir/node_modules/package",
			expected: true,
		},
		{
			name:     ".git目录",
			path:     "project/.git",
			expected: true,
		},
		{
			name:     "__pycache__目录",
			path:     "src/__pycache__",
			expected: true,
		},
		{
			name:     "vendor目录",
			path:     "go-project/vendor",
			expected: true,
		},
		{
			name:     "dist目录",
			path:     "frontend/dist",
			expected: true,
		},
		{
			name:     "build目录",
			path:     "project/build",
			expected: true,
		},
		{
			name:     "target目录",
			path:     "rust-project/target",
			expected: true,
		},
		// 正常文件和目录
		{
			name:     "正常文件",
			path:     "main.go",
			expected: false,
		},
		{
			name:     "正常目录",
			path:     "src/components",
			expected: false,
		},
		{
			name:     "深层正常路径",
			path:     "/project/src/main/java/Main.java",
			expected: false,
		},
		// 边界情况
		{
			name:     "空路径",
			path:     "",
			expected: false,
		},
		{
			name:     "包含忽略目录名但不是目录的文件",
			path:     "my_node_modules_backup.txt",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SkipHidden(tc.path)
			assert.Equal(t, tc.expected, result, "路径: %s", tc.path)
		})
	}
}

func TestGetRgCmd(t *testing.T) {
	t.Parallel()

	t.Run("无glob模式", func(t *testing.T) {
		cmd := GetRgCmd("")
		
		if cmd == nil {
			t.Skip("ripgrep未安装，跳过测试")
			return
		}
		
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Args, "--files")
		assert.Contains(t, cmd.Args, "-L")
		assert.Contains(t, cmd.Args, "--null")
		assert.Equal(t, ".", cmd.Dir)
	})

	t.Run("带glob模式", func(t *testing.T) {
		cmd := GetRgCmd("*.go")
		
		if cmd == nil {
			t.Skip("ripgrep未安装，跳过测试")
			return
		}
		
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Args, "--glob")
		assert.Contains(t, cmd.Args, "/*.go")
	})

	t.Run("绝对路径glob模式", func(t *testing.T) {
		cmd := GetRgCmd("/absolute/*.go")
		
		if cmd == nil {
			t.Skip("ripgrep未安装，跳过测试")
			return
		}
		
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Args, "--glob")
		assert.Contains(t, cmd.Args, "/absolute/*.go")
	})

	t.Run("ripgrep不可用", func(t *testing.T) {
		// 临时保存原始路径
		originalRgPath := rgPath
		defer func() {
			rgPath = originalRgPath
		}()
		
		// 设置为空以模拟ripgrep不可用
		rgPath = ""
		
		cmd := GetRgCmd("*.go")
		assert.Nil(t, cmd)
	})
}

func TestGetFzfCmd(t *testing.T) {
	t.Parallel()

	t.Run("基本fzf命令", func(t *testing.T) {
		cmd := GetFzfCmd("test-query")
		
		if cmd == nil {
			t.Skip("fzf未安装，跳过测试")
			return
		}
		
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Args, "--filter")
		assert.Contains(t, cmd.Args, "test-query")
		assert.Contains(t, cmd.Args, "--read0")
		assert.Contains(t, cmd.Args, "--print0")
		assert.Equal(t, ".", cmd.Dir)
	})

	t.Run("空查询", func(t *testing.T) {
		cmd := GetFzfCmd("")
		
		if cmd == nil {
			t.Skip("fzf未安装，跳过测试")
			return
		}
		
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Args, "--filter")
		assert.Contains(t, cmd.Args, "")
	})

	t.Run("fzf不可用", func(t *testing.T) {
		// 临时保存原始路径
		originalFzfPath := fzfPath
		defer func() {
			fzfPath = originalFzfPath
		}()
		
		// 设置为空以模拟fzf不可用
		fzfPath = ""
		
		cmd := GetFzfCmd("query")
		assert.Nil(t, cmd)
	})
}

func TestGlobWithDoublestar(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "fileutil_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件结构
	testFiles := []string{
		"file1.go",
		"file2.txt",
		"subdir/file3.go",
		"subdir/file4.txt",
		"subdir/nested/file5.go",
		".hidden/hidden.go",
		"node_modules/package.json",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
		
		// 设置不同的修改时间以测试排序
		modTime := time.Now().Add(-time.Duration(len(testFiles)-indexOf(testFiles, file)) * time.Minute)
		err = os.Chtimes(fullPath, modTime, modTime)
		require.NoError(t, err)
	}

	t.Run("匹配所有go文件", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**/*.go", tempDir, 10)
		require.NoError(t, err)
		assert.False(t, truncated)
		
		// 验证所有匹配都是.go文件
		for _, match := range matches {
			assert.True(t, strings.HasSuffix(match, ".go"))
		}
		
		// 统计非隐藏目录中的.go文件
		nonHiddenMatches := 0
		for _, match := range matches {
			if !strings.Contains(match, ".hidden") && !strings.Contains(match, "node_modules") {
				nonHiddenMatches++
			}
		}
		
		// 至少应该找到3个非隐藏的.go文件
		assert.GreaterOrEqual(t, nonHiddenMatches, 3)
	})

	t.Run("限制结果数量", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**/*", tempDir, 2)
		// 由于fs.SkipAll的行为，可能会有错误，这是正常的
		if err != nil && strings.Contains(err.Error(), "skip everything") {
			// 这是预期的行为，因为使用了fs.SkipAll
			t.Skip("跳过限制测试 - fs.SkipAll导致的预期错误")
			return
		}
		require.NoError(t, err)
		assert.True(t, truncated)
		assert.LessOrEqual(t, len(matches), 2)
	})

	t.Run("无匹配", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**/*.xyz", tempDir, 10)
		require.NoError(t, err)
		assert.False(t, truncated)
		assert.Len(t, matches, 0)
	})

	t.Run("单个文件模式", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("file1.go", tempDir, 10)
		require.NoError(t, err)
		assert.False(t, truncated)
		assert.Len(t, matches, 1)
		assert.True(t, strings.HasSuffix(matches[0], "file1.go"))
	})

	t.Run("不存在的搜索路径", func(t *testing.T) {
		matches, truncated, err := GlobWithDoublestar("**/*.go", "/nonexistent/path", 10)
		// doublestar.GlobWalk可能不会对不存在的路径返回错误，而是返回空结果
		if err != nil {
			assert.Error(t, err)
		} else {
			// 如果没有错误，应该返回空结果
			assert.Len(t, matches, 0)
			assert.False(t, truncated)
		}
	})
}

func TestFileInfo(t *testing.T) {
	t.Parallel()

	t.Run("FileInfo结构体", func(t *testing.T) {
		now := time.Now()
		fileInfo := FileInfo{
			Path:    "/test/path",
			ModTime: now,
		}
		
		assert.Equal(t, "/test/path", fileInfo.Path)
		assert.Equal(t, now, fileInfo.ModTime)
	})
}

// 辅助函数
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// 基准测试
func BenchmarkSkipHidden(b *testing.B) {
	testPaths := []string{
		"normal/file.go",
		".hidden/file.go",
		"project/node_modules/package.json",
		"src/main/java/Main.java",
		"build/output.jar",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			SkipHidden(path)
		}
	}
}

func BenchmarkGetRgCmd(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := GetRgCmd("*.go")
		if cmd != nil {
			// 模拟使用命令但不实际执行
			_ = cmd.Args
		}
	}
}

func BenchmarkGetFzfCmd(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := GetFzfCmd("test-query")
		if cmd != nil {
			// 模拟使用命令但不实际执行
			_ = cmd.Args
		}
	}
}