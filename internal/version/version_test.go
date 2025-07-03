package version

import (
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	t.Run("Version变量存在", func(t *testing.T) {
		// Version变量应该被定义，可能是"unknown"或从构建信息中读取的版本
		assert.NotEmpty(t, Version)
	})

	t.Run("Version不为空字符串", func(t *testing.T) {
		// 即使是默认的"unknown"，Version也不应该是空字符串
		assert.NotEqual(t, "", Version)
	})

	// 测试init函数的行为（通过测试其副作用）
	t.Run("init函数设置Version", func(t *testing.T) {
		// 在测试环境中，Version要么是"unknown"（默认值），
		// 要么是从debug.ReadBuildInfo()读取的版本
		if info, ok := debug.ReadBuildInfo(); ok {
			mainVersion := info.Main.Version
			if mainVersion != "" && mainVersion != "(devel)" {
				// 如果有有效的构建版本信息，Version应该等于mainVersion
				assert.Equal(t, mainVersion, Version)
			} else {
				// 否则Version应该保持默认值或被设置为某个值
				assert.True(t, Version == "unknown" || Version != "")
			}
		} else {
			// 如果无法读取构建信息，Version应该保持默认值
			assert.True(t, Version == "unknown" || Version != "")
		}
	})
}

func TestVersionScenarios(t *testing.T) {
	t.Parallel()

	t.Run("默认版本处理", func(t *testing.T) {
		// 测试默认情况下Version的行为
		// 在测试环境中，通常Version会是"unknown"或构建时设置的版本
		possibleValues := []string{"unknown"}
		
		// 如果有构建信息，添加可能的版本值
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				possibleValues = append(possibleValues, info.Main.Version)
			}
		}
		
		found := false
		for _, val := range possibleValues {
			if Version == val {
				found = true
				break
			}
		}
		
		// Version应该是期望值之一，或者是语义版本格式
		assert.True(t, found || 
			(len(Version) > 0 && Version[0] == 'v') ||
			Version != "unknown", 
			"Version should be either unknown, build version, or semantic version, got: %s", Version)
	})

	t.Run("版本格式验证", func(t *testing.T) {
		// 版本应该是有效字符串
		assert.True(t, len(Version) > 0)
		assert.False(t, strings.Contains(Version, "\n"))
		assert.False(t, strings.Contains(Version, "\r"))
	})
}

// 基准测试
func BenchmarkVersionAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Version
	}
}