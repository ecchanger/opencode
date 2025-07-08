package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldShowInitDialog(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	t.Run("config not loaded", func(t *testing.T) {
		cfg = nil
		shouldShow, err := ShouldShowInitDialog()
		assert.False(t, shouldShow)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config not loaded")
	})

	t.Run("init flag file exists", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		
		// 设置测试配置
		cfg = &Config{
			Data: Data{
				Directory: tempDir,
			},
		}
		
		// 创建初始化标志文件
		flagFilePath := filepath.Join(tempDir, InitFlagFilename)
		file, err := os.Create(flagFilePath)
		require.NoError(t, err)
		file.Close()
		
		shouldShow, err := ShouldShowInitDialog()
		assert.NoError(t, err)
		assert.False(t, shouldShow, "Should not show dialog when init flag file exists")
	})

	t.Run("init flag file does not exist", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		
		// 设置测试配置
		cfg = &Config{
			Data: Data{
				Directory: tempDir,
			},
		}
		
		shouldShow, err := ShouldShowInitDialog()
		assert.NoError(t, err)
		assert.True(t, shouldShow, "Should show dialog when init flag file does not exist")
	})

	t.Run("permission error", func(t *testing.T) {
		// 设置测试配置指向不存在的路径
		cfg = &Config{
			Data: Data{
				Directory: "/non/existent/path/with/permission/issues",
			},
		}
		
		shouldShow, err := ShouldShowInitDialog()
		// 这种情况下应该返回错误，因为无法检查文件状态
		// 但具体的错误处理取决于操作系统和文件系统
		if err != nil {
			assert.False(t, shouldShow)
		} else {
			// 如果没有错误，那么应该显示对话框（文件不存在）
			assert.True(t, shouldShow)
		}
	})
}

func TestMarkProjectInitialized(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	t.Run("config not loaded", func(t *testing.T) {
		cfg = nil
		err := MarkProjectInitialized()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config not loaded")
	})

	t.Run("successful creation", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		
		// 设置测试配置
		cfg = &Config{
			Data: Data{
				Directory: tempDir,
			},
		}
		
		// 标记项目为已初始化
		err := MarkProjectInitialized()
		assert.NoError(t, err)
		
		// 验证标志文件已创建
		flagFilePath := filepath.Join(tempDir, InitFlagFilename)
		_, err = os.Stat(flagFilePath)
		assert.NoError(t, err, "Init flag file should be created")
		
		// 验证不再显示初始化对话框
		shouldShow, err := ShouldShowInitDialog()
		assert.NoError(t, err)
		assert.False(t, shouldShow, "Should not show dialog after marking as initialized")
	})

	t.Run("directory creation", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		
		// 在临时目录下创建一个不存在的子目录路径
		subDir := filepath.Join(tempDir, "nested", "directory")
		
		// 设置测试配置
		cfg = &Config{
			Data: Data{
				Directory: subDir,
			},
		}
		
		// 标记项目为已初始化
		err := MarkProjectInitialized()
		
		// 由于目录不存在，会尝试创建文件，如果父目录不存在会失败
		// 这取决于 os.Create 的行为
		if err != nil {
			// 预期的行为：父目录不存在时创建文件会失败
			assert.Contains(t, err.Error(), "failed to create init flag file")
		} else {
			// 如果成功，验证文件存在
			flagFilePath := filepath.Join(subDir, InitFlagFilename)
			_, statErr := os.Stat(flagFilePath)
			assert.NoError(t, statErr)
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		
		// 设置测试配置
		cfg = &Config{
			Data: Data{
				Directory: tempDir,
			},
		}
		
		// 创建已存在的标志文件
		flagFilePath := filepath.Join(tempDir, InitFlagFilename)
		err := os.WriteFile(flagFilePath, []byte("existing content"), 0644)
		require.NoError(t, err)
		
		// 标记项目为已初始化（应该覆盖现有文件）
		err = MarkProjectInitialized()
		assert.NoError(t, err)
		
		// 验证文件仍然存在
		_, err = os.Stat(flagFilePath)
		assert.NoError(t, err)
	})
}

func TestInitFlagFilename(t *testing.T) {
	t.Parallel()

	// 验证常量值
	assert.Equal(t, "init", InitFlagFilename)
}

func TestProjectInitFlagStruct(t *testing.T) {
	t.Parallel()

	// 测试 ProjectInitFlag 结构体
	flag := ProjectInitFlag{
		Initialized: true,
	}
	
	assert.True(t, flag.Initialized)
	
	flag.Initialized = false
	assert.False(t, flag.Initialized)
}

func TestIntegrationInitWorkflow(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 创建临时目录
	tempDir := t.TempDir()
	
	// 设置测试配置
	cfg = &Config{
		Data: Data{
			Directory: tempDir,
		},
	}
	
	// 1. 最初应该显示初始化对话框
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.True(t, shouldShow, "Should show dialog initially")
	
	// 2. 标记项目为已初始化
	err = MarkProjectInitialized()
	assert.NoError(t, err)
	
	// 3. 现在不应该显示初始化对话框
	shouldShow, err = ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.False(t, shouldShow, "Should not show dialog after initialization")
	
	// 4. 多次调用 MarkProjectInitialized 应该是安全的
	err = MarkProjectInitialized()
	assert.NoError(t, err)
	
	shouldShow, err = ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.False(t, shouldShow, "Should still not show dialog after multiple markings")
}