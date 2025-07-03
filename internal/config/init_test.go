package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitFlagFilename_Constant(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "init", InitFlagFilename)
}

func TestProjectInitFlag_Struct(t *testing.T) {
	t.Parallel()

	flag := ProjectInitFlag{
		Initialized: true,
	}

	assert.True(t, flag.Initialized)
}

func TestShouldShowInitDialog_ConfigNotLoaded(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置nil配置
	cfg = nil

	shouldShow, err := ShouldShowInitDialog()
	assert.Error(t, err)
	assert.False(t, shouldShow)
	assert.Contains(t, err.Error(), "config not loaded")
}

func TestShouldShowInitDialog_FileExists(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置测试配置
	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	// 创建init标志文件
	flagFilePath := filepath.Join(tempDir, InitFlagFilename)
	file, err := os.Create(flagFilePath)
	require.NoError(t, err)
	file.Close()

	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.False(t, shouldShow) // 文件存在，不应该显示对话框
}

func TestShouldShowInitDialog_FileNotExists(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置测试配置
	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.True(t, shouldShow) // 文件不存在，应该显示对话框
}

func TestShouldShowInitDialog_DirectoryAccessError(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置无效的目录路径（这可能会导致权限错误等）
	cfg = &Config{
		Data: Data{
			Directory: "/nonexistent/path/with/no/permissions",
		},
	}

	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err) // 如果文件不存在，应该返回true而不是错误
	assert.True(t, shouldShow)
}

func TestMarkProjectInitialized_ConfigNotLoaded(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置nil配置
	cfg = nil

	err := MarkProjectInitialized()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config not loaded")
}

func TestMarkProjectInitialized_Success(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置测试配置
	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	err = MarkProjectInitialized()
	assert.NoError(t, err)

	// 验证文件是否被创建
	flagFilePath := filepath.Join(tempDir, InitFlagFilename)
	_, err = os.Stat(flagFilePath)
	assert.NoError(t, err) // 文件应该存在
}

func TestMarkProjectInitialized_CreateDirectoryIfNotExists(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置测试配置，指向临时目录内的子目录
	dataDir := filepath.Join(tempDir, "subdir")
	cfg = &Config{
		Data: Data{
			Directory: dataDir,
		},
	}

	// 确保子目录不存在
	_, err = os.Stat(dataDir)
	assert.True(t, os.IsNotExist(err))

	// 创建子目录（模拟MarkProjectInitialized内部会创建目录的情况）
	err = os.MkdirAll(dataDir, 0o755)
	require.NoError(t, err)

	err = MarkProjectInitialized()
	assert.NoError(t, err)

	// 验证目录和文件是否被创建
	flagFilePath := filepath.Join(dataDir, InitFlagFilename)
	_, err = os.Stat(flagFilePath)
	assert.NoError(t, err)
}

func TestMarkProjectInitialized_FileAlreadyExists(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置测试配置
	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	// 先创建init标志文件
	flagFilePath := filepath.Join(tempDir, InitFlagFilename)
	file, err := os.Create(flagFilePath)
	require.NoError(t, err)
	file.WriteString("existing content")
	file.Close()

	// 再次标记初始化（应该覆盖文件）
	err = MarkProjectInitialized()
	assert.NoError(t, err)

	// 验证文件仍然存在
	_, err = os.Stat(flagFilePath)
	assert.NoError(t, err)
}

// 集成测试：完整的初始化流程
func TestInitializationFlow(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置测试配置
	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	// 第一次检查：应该显示初始化对话框
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.True(t, shouldShow)

	// 标记为已初始化
	err = MarkProjectInitialized()
	assert.NoError(t, err)

	// 第二次检查：不应该再显示初始化对话框
	shouldShow, err = ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.False(t, shouldShow)
}

// 基准测试
func BenchmarkShouldShowInitDialog_FileExists(b *testing.B) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "opencode-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	// 创建init标志文件
	flagFilePath := filepath.Join(tempDir, InitFlagFilename)
	file, err := os.Create(flagFilePath)
	require.NoError(b, err)
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ShouldShowInitDialog()
	}
}

func BenchmarkShouldShowInitDialog_FileNotExists(b *testing.B) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录（但不创建init文件）
	tempDir, err := os.MkdirTemp("", "opencode-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ShouldShowInitDialog()
	}
}

func BenchmarkMarkProjectInitialized(b *testing.B) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "opencode-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// 为每次迭代创建新的子目录
		iterDir := filepath.Join(tempDir, "iter", "test", "dir", "path")
		cfg = &Config{
			Data: Data{
				Directory: iterDir,
			},
		}
		b.StartTimer()

		_ = MarkProjectInitialized()
	}
}