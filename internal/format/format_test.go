package format

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormat(t *testing.T) {
	t.Parallel()

	t.Run("OutputFormat类型", func(t *testing.T) {
		// 测试常量定义
		assert.Equal(t, "text", string(Text))
		assert.Equal(t, "json", string(JSON))
	})

	t.Run("String方法", func(t *testing.T) {
		assert.Equal(t, "text", Text.String())
		assert.Equal(t, "json", JSON.String())
	})

	t.Run("SupportedFormats", func(t *testing.T) {
		assert.Len(t, SupportedFormats, 2)
		assert.Contains(t, SupportedFormats, "text")
		assert.Contains(t, SupportedFormats, "json")
	})
}

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected OutputFormat
		hasError bool
	}{
		{
			name:     "text格式",
			input:    "text",
			expected: Text,
			hasError: false,
		},
		{
			name:     "json格式",
			input:    "json",
			expected: JSON,
			hasError: false,
		},
		{
			name:     "大写text",
			input:    "TEXT",
			expected: Text,
			hasError: false,
		},
		{
			name:     "大写json",
			input:    "JSON",
			expected: JSON,
			hasError: false,
		},
		{
			name:     "混合大小写",
			input:    "TeXt",
			expected: Text,
			hasError: false,
		},
		{
			name:     "带空格",
			input:    "  json  ",
			expected: JSON,
			hasError: false,
		},
		{
			name:     "无效格式",
			input:    "xml",
			expected: "",
			hasError: true,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
			hasError: true,
		},
		{
			name:     "无效字符",
			input:    "invalid@format",
			expected: "",
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse(tc.input)
			
			if tc.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid format")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	validCases := []string{
		"text",
		"json",
		"TEXT",
		"JSON",
		"  text  ",
		"  json  ",
	}

	invalidCases := []string{
		"xml",
		"yaml",
		"",
		"invalid",
		"text,json",
		"123",
	}

	for _, input := range validCases {
		t.Run("valid_"+input, func(t *testing.T) {
			assert.True(t, IsValid(input))
		})
	}

	for _, input := range invalidCases {
		t.Run("invalid_"+input, func(t *testing.T) {
			assert.False(t, IsValid(input))
		})
	}
}

func TestGetHelpText(t *testing.T) {
	t.Parallel()

	helpText := GetHelpText()
	
	assert.NotEmpty(t, helpText)
	assert.Contains(t, helpText, "Supported output formats:")
	assert.Contains(t, helpText, "text")
	assert.Contains(t, helpText, "json")
	assert.Contains(t, helpText, "Plain text output")
	assert.Contains(t, helpText, "JSON object")
}

func TestFormatOutput(t *testing.T) {
	t.Parallel()

	testContent := "Hello, World!"

	t.Run("text格式", func(t *testing.T) {
		result := FormatOutput(testContent, "text")
		assert.Equal(t, testContent, result)
	})

	t.Run("json格式", func(t *testing.T) {
		result := FormatOutput(testContent, "json")
		
		// 验证JSON格式
		assert.Contains(t, result, "response")
		assert.Contains(t, result, testContent)
		
		// 验证是有效的JSON
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, testContent, jsonData["response"])
	})

	t.Run("大写JSON格式", func(t *testing.T) {
		result := FormatOutput(testContent, "JSON")
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, testContent, jsonData["response"])
	})

	t.Run("无效格式默认为text", func(t *testing.T) {
		result := FormatOutput(testContent, "invalid")
		assert.Equal(t, testContent, result)
	})

	t.Run("空格式默认为text", func(t *testing.T) {
		result := FormatOutput(testContent, "")
		assert.Equal(t, testContent, result)
	})
}

func TestFormatAsJSON(t *testing.T) {
	t.Parallel()

	t.Run("普通文本", func(t *testing.T) {
		content := "Simple text"
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})

	t.Run("包含特殊字符", func(t *testing.T) {
		content := `Text with "quotes" and \backslashes`
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})

	t.Run("多行文本", func(t *testing.T) {
		content := "Line 1\nLine 2\nLine 3"
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})

	t.Run("包含制表符", func(t *testing.T) {
		content := "Text\twith\ttabs"
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})

	t.Run("包含回车符", func(t *testing.T) {
		content := "Text\rwith\rcarriage\rreturns"
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})

	t.Run("空字符串", func(t *testing.T) {
		content := ""
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, "", jsonData["response"])
	})

	t.Run("JSON格式验证", func(t *testing.T) {
		content := "test content"
		result := formatAsJSON(content)
		
		// 验证JSON格式
		assert.True(t, strings.HasPrefix(result, "{"))
		assert.True(t, strings.HasSuffix(result, "}"))
		assert.Contains(t, result, `"response"`)
		assert.Contains(t, result, `"test content"`)
	})

	t.Run("Unicode字符", func(t *testing.T) {
		content := "Hello 世界 🌍"
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})

	t.Run("复杂JSON字符", func(t *testing.T) {
		content := `{"nested": "json", "array": [1, 2, 3], "escaped": "\"quotes\""}`
		result := formatAsJSON(content)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, content, jsonData["response"])
	})
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("极长文本", func(t *testing.T) {
		longContent := strings.Repeat("a", 10000)
		result := FormatOutput(longContent, "json")
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		assert.Equal(t, longContent, jsonData["response"])
	})

	t.Run("包含所有特殊字符", func(t *testing.T) {
		specialChars := "\"\\'\n\r\t\b\f\v\x00"
		result := formatAsJSON(specialChars)
		
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result), &jsonData)
		assert.NoError(t, err)
		// 注意：\x00在JSON中可能会被处理为空字符
		assert.Contains(t, jsonData["response"].(string), "\"")
		assert.Contains(t, jsonData["response"].(string), "\\")
	})
}

// 基准测试
func BenchmarkParse(b *testing.B) {
	testInputs := []string{"text", "json", "TEXT", "JSON", "invalid"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := testInputs[i%len(testInputs)]
		_, _ = Parse(input)
	}
}

func BenchmarkIsValid(b *testing.B) {
	testInputs := []string{"text", "json", "invalid", "xml"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := testInputs[i%len(testInputs)]
		_ = IsValid(input)
	}
}

func BenchmarkFormatOutput(b *testing.B) {
	content := "This is a test content for benchmarking the format output function"
	
	b.Run("text_format", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = FormatOutput(content, "text")
		}
	})
	
	b.Run("json_format", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = FormatOutput(content, "json")
		}
	})
}

func BenchmarkFormatAsJSON(b *testing.B) {
	content := "This is a test content with some special characters: \"quotes\", \\backslashes\\, and\nnewlines."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatAsJSON(content)
	}
}