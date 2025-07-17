package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileOperations(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "fileutil-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("文件存在性检查", func(t *testing.T) {
		// 创建一个测试文件
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// 测试文件是否存在
		exists := fileExists(testFile)
		assert.True(t, exists, "文件应该存在")

		// 测试不存在的文件
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		exists = fileExists(nonExistentFile)
		assert.False(t, exists, "文件不应该存在")
	})

	t.Run("目录操作", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "subdir")
		
		// 测试目录创建
		err := os.MkdirAll(testDir, 0755)
		assert.NoError(t, err)

		// 验证目录是否存在
		info, err := os.Stat(testDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestPathOperations(t *testing.T) {
	t.Run("路径清理", func(t *testing.T) {
		// 测试路径清理功能
		dirtyPath := "/path/to/../file.txt"
		cleanPath := filepath.Clean(dirtyPath)
		assert.Equal(t, "/path/file.txt", cleanPath)
	})

	t.Run("相对路径转换", func(t *testing.T) {
		// 测试相对路径转换
		basePath := "/home/user/project"
		targetPath := "/home/user/project/src/main.go"
		
		relPath, err := filepath.Rel(basePath, targetPath)
		assert.NoError(t, err)
		assert.Equal(t, "src/main.go", relPath)
	})
}

// 辅助函数：检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 基准测试：文件操作性能
func BenchmarkFileExists(b *testing.B) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "bench-*")
	require.NoError(b, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fileExists(tempFile.Name())
	}
}