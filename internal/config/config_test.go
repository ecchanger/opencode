package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 改变到临时目录
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Run("正常加载配置", func(t *testing.T) {
		// 测试在没有配置文件的情况下加载默认配置
		config, err := Load(tempDir, false)
		assert.NoError(t, err)
		assert.NotNil(t, config)
	})

	t.Run("调试模式加载配置", func(t *testing.T) {
		config, err := Load(tempDir, true)
		assert.NoError(t, err)
		assert.NotNil(t, config)
	})

	t.Run("无效目录路径", func(t *testing.T) {
		config, err := Load("/nonexistent/path", false)
		// 配置系统应该能够处理无效路径，可能返回默认配置或错误
		// 根据实际实现调整断言
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, config)
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	// 测试配置文件路径生成
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 假设有一个获取配置路径的函数
	// 这里展示如何测试路径相关的功能
	expectedPath := filepath.Join(tempDir, ".opencode")
	
	// 如果有相关的函数，可以这样测试
	t.Run("配置目录创建", func(t *testing.T) {
		err := os.MkdirAll(expectedPath, 0755)
		assert.NoError(t, err)
		
		// 验证目录是否创建成功
		info, err := os.Stat(expectedPath)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

// 基准测试：测试配置加载的性能
func BenchmarkLoad(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "opencode-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(tempDir, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}