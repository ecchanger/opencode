package version

import (
	"testing"
	"regexp"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Run("版本不为空", func(t *testing.T) {
		assert.NotEmpty(t, Version, "版本字符串不应为空")
	})

	t.Run("版本格式验证", func(t *testing.T) {
		// 验证版本格式是否合理（如：v1.0.0, unknown, (devel), 或其他格式）
		assert.NotEqual(t, "", Version, "版本不能为空字符串")
		
		// 验证版本是否为已知的有效值
		validVersions := []string{"unknown", "(devel)", "dev"}
		isValidVersion := false
		
		// 检查是否为预设的有效值
		for _, valid := range validVersions {
			if Version == valid {
				isValidVersion = true
				break
			}
		}
		
		// 如果不是预设值，检查是否符合语义版本格式
		if !isValidVersion {
			matched, err := regexp.MatchString(`^v?\d+\.\d+\.\d+`, Version)
			assert.NoError(t, err)
			isValidVersion = matched
		}
		
		assert.True(t, isValidVersion, "版本格式应该符合语义版本规范或为已知的有效值，当前版本: %s", Version)
	})
}

// 如果有其他版本相关的函数，可以在这里添加测试
func TestVersionInfo(t *testing.T) {
	// 假设有一个获取版本信息的函数
	t.Run("版本信息结构", func(t *testing.T) {
		// 测试版本信息的结构和内容
		info := Version
		assert.IsType(t, "", info, "版本应该是字符串类型")
	})
}

// 基准测试：测试版本字符串访问的性能
func BenchmarkVersionAccess(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Version
	}
}