package version

import (
	"strings"
	"testing"
)

func TestVersionInitialization(t *testing.T) {
	// 测试版本变量不为空
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestVersionFormat(t *testing.T) {
	// 测试版本格式是否合理
	if Version == "unknown" {
		// 这是默认值，在开发环境中是正常的
		return
	}
	
	// 如果不是"unknown"，应该包含某种版本格式
	if !strings.Contains(Version, ".") && !strings.HasPrefix(Version, "v") {
		t.Logf("Version format may be non-standard: %s", Version)
	}
}

func TestVersionNonEmpty(t *testing.T) {
	// 测试版本不为空字符串
	if len(strings.TrimSpace(Version)) == 0 {
		t.Error("Version should not be empty after trimming whitespace")
	}
}