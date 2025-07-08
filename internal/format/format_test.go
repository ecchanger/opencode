package format

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputFormat_String(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		format   OutputFormat
		expected string
	}{
		{Text, "text"},
		{JSON, "json"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.format.String())
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       string
		expected    OutputFormat
		expectError bool
	}{
		{"text", Text, false},
		{"json", JSON, false},
		{"TEXT", Text, false}, // 测试大小写不敏感
		{"JSON", JSON, false},
		{" text ", Text, false}, // 测试前后空格
		{"invalid", "", true},
		{"", "", true},
		{"xml", "", true},
	}

	for _, tt := range tests {
		result, err := Parse(tt.input)
		if tt.expectError {
			assert.Error(t, err, "Expected error for input: %s", tt.input)
		} else {
			assert.NoError(t, err, "Unexpected error for input: %s", tt.input)
			assert.Equal(t, tt.expected, result, "Unexpected result for input: %s", tt.input)
		}
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"text", true},
		{"json", true},
		{"TEXT", true},
		{"JSON", true},
		{" text ", true},
		{"invalid", false},
		{"", false},
		{"xml", false},
	}

	for _, tt := range tests {
		result := IsValid(tt.input)
		assert.Equal(t, tt.expected, result, "Unexpected result for input: %s", tt.input)
	}
}

func TestGetHelpText(t *testing.T) {
	t.Parallel()

	helpText := GetHelpText()
	
	// 检查帮助文本包含所有支持的格式
	assert.Contains(t, helpText, "text")
	assert.Contains(t, helpText, "json")
	assert.Contains(t, helpText, "Plain text output")
	assert.Contains(t, helpText, "JSON object")
}

func TestFormatOutput(t *testing.T) {
	t.Parallel()

	testContent := "Hello, world!"

	tests := []struct {
		content    string
		formatStr  string
		expected   string
		shouldTest func(t *testing.T, result string)
	}{
		{
			content:   testContent,
			formatStr: "text",
			expected:  testContent,
		},
		{
			content:   testContent,
			formatStr: "json",
			shouldTest: func(t *testing.T, result string) {
				// 验证 JSON 格式
				var parsed struct {
					Response string `json:"response"`
				}
				err := json.Unmarshal([]byte(result), &parsed)
				require.NoError(t, err)
				assert.Equal(t, testContent, parsed.Response)
			},
		},
		{
			content:   testContent,
			formatStr: "invalid",
			expected:  testContent, // 应该回退到文本格式
		},
		{
			content:   "",
			formatStr: "json",
			shouldTest: func(t *testing.T, result string) {
				var parsed struct {
					Response string `json:"response"`
				}
				err := json.Unmarshal([]byte(result), &parsed)
				require.NoError(t, err)
				assert.Equal(t, "", parsed.Response)
			},
		},
	}

	for _, tt := range tests {
		result := FormatOutput(tt.content, tt.formatStr)
		if tt.shouldTest != nil {
			tt.shouldTest(t, result)
		} else {
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestFormatAsJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{"simple text", "Hello, world!"},
		{"empty string", ""},
		{"with quotes", `This has "quotes" in it`},
		{"with newlines", "Line 1\nLine 2\nLine 3"},
		{"with special chars", "Tab:\t Backslash:\\ Quote:\""},
		{"unicode", "Hello 世界! 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAsJSON(tt.content)
			
			// 验证结果是有效的 JSON
			var parsed struct {
				Response string `json:"response"`
			}
			err := json.Unmarshal([]byte(result), &parsed)
			require.NoError(t, err, "Result should be valid JSON: %s", result)
			
			// 验证内容正确
			assert.Equal(t, tt.content, parsed.Response)
			
			// 验证格式化的结构
			assert.Contains(t, result, "\"response\":")
			assert.True(t, strings.HasPrefix(result, "{"))
			assert.True(t, strings.HasSuffix(result, "}"))
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	t.Parallel()

	// 验证支持的格式列表
	expectedFormats := []string{"text", "json"}
	assert.Equal(t, expectedFormats, SupportedFormats)
	
	// 验证所有支持的格式都是有效的
	for _, format := range SupportedFormats {
		assert.True(t, IsValid(format), "Supported format should be valid: %s", format)
	}
}