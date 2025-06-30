package format

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormat_String(t *testing.T) {
	tests := []struct {
		name     string
		format   OutputFormat
		expected string
	}{
		{"text format", Text, "text"},
		{"json format", JSON, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.String())
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  OutputFormat
		wantError bool
	}{
		{"parse text", "text", Text, false},
		{"parse json", "json", JSON, false},
		{"parse TEXT uppercase", "TEXT", Text, false},
		{"parse JSON uppercase", "JSON", JSON, false},
		{"parse with spaces", "  text  ", Text, false},
		{"parse invalid", "invalid", "", true},
		{"parse empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			
			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, OutputFormat(""), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid text", "text", true},
		{"valid json", "json", true},
		{"valid TEXT uppercase", "TEXT", true},
		{"valid JSON uppercase", "JSON", true},
		{"valid with spaces", "  json  ", true},
		{"invalid format", "invalid", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValid(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetHelpText(t *testing.T) {
	helpText := GetHelpText()
	
	assert.Contains(t, helpText, "text")
	assert.Contains(t, helpText, "json")
	assert.Contains(t, helpText, "Plain text output")
	assert.Contains(t, helpText, "JSON object")
}

func TestSupportedFormats(t *testing.T) {
	assert.Contains(t, SupportedFormats, "text")
	assert.Contains(t, SupportedFormats, "json")
	assert.Len(t, SupportedFormats, 2)
}

func TestFormatOutput(t *testing.T) {
	testContent := "Hello, world!"
	
	tests := []struct {
		name      string
		content   string
		formatStr string
		expected  string
	}{
		{
			name:      "format as text",
			content:   testContent,
			formatStr: "text",
			expected:  testContent,
		},
		{
			name:      "format as json",
			content:   testContent,
			formatStr: "json",
			expected:  formatAsJSON(testContent),
		},
		{
			name:      "invalid format defaults to text",
			content:   testContent,
			formatStr: "invalid",
			expected:  testContent,
		},
		{
			name:      "empty format defaults to text",
			content:   testContent,
			formatStr: "",
			expected:  testContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOutput(tt.content, tt.formatStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatAsJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected struct {
			Response string `json:"response"`
		}
	}{
		{
			name:    "simple text",
			content: "Hello, world!",
			expected: struct {
				Response string `json:"response"`
			}{Response: "Hello, world!"},
		},
		{
			name:    "text with quotes",
			content: `He said "Hello"`,
			expected: struct {
				Response string `json:"response"`
			}{Response: `He said "Hello"`},
		},
		{
			name:    "text with newlines",
			content: "Line 1\nLine 2\nLine 3",
			expected: struct {
				Response string `json:"response"`
			}{Response: "Line 1\nLine 2\nLine 3"},
		},
		{
			name:    "text with backslashes",
			content: "Path: C:\\Users\\test",
			expected: struct {
				Response string `json:"response"`
			}{Response: "Path: C:\\Users\\test"},
		},
		{
			name:    "text with tabs",
			content: "Column 1\tColumn 2",
			expected: struct {
				Response string `json:"response"`
			}{Response: "Column 1\tColumn 2"},
		},
		{
			name:    "empty content",
			content: "",
			expected: struct {
				Response string `json:"response"`
			}{Response: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAsJSON(tt.content)
			
			// Parse the result back to verify it's valid JSON
			var parsed struct {
				Response string `json:"response"`
			}
			err := json.Unmarshal([]byte(result), &parsed)
			assert.NoError(t, err)
			
			// Check that the parsed content matches expected
			assert.Equal(t, tt.expected.Response, parsed.Response)
			
			// Verify the JSON is properly formatted (indented)
			expectedJSON, err := json.MarshalIndent(tt.expected, "", "  ")
			assert.NoError(t, err)
			assert.Equal(t, string(expectedJSON), result)
		})
	}
}

func TestFormatAsJSON_ComplexContent(t *testing.T) {
	// Test with complex content that includes various special characters
	complexContent := `{
		"message": "Hello \"world\"",
		"path": "C:\\Users\\test",
		"multiline": "Line 1\nLine 2\r\nLine 3",
		"tab": "Column1\tColumn2"
	}`
	
	result := formatAsJSON(complexContent)
	
	// Should be valid JSON
	var parsed struct {
		Response string `json:"response"`
	}
	err := json.Unmarshal([]byte(result), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, complexContent, parsed.Response)
}

func TestFormatOutput_JSONFormatting(t *testing.T) {
	content := "Test content with special chars: \"quotes\", \\backslashes\\, \nnewlines\n"
	
	// Test JSON format
	jsonResult := FormatOutput(content, "json")
	
	// Should be valid JSON
	var parsed struct {
		Response string `json:"response"`
	}
	err := json.Unmarshal([]byte(jsonResult), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, content, parsed.Response)
	
	// Test text format
	textResult := FormatOutput(content, "text")
	assert.Equal(t, content, textResult)
}

func TestFormatOutput_CaseSensitivity(t *testing.T) {
	content := "Test content"
	
	tests := []string{"json", "JSON", "Json", "TEXT", "text", "Text"}
	
	for _, formatStr := range tests {
		t.Run(formatStr, func(t *testing.T) {
			result := FormatOutput(content, formatStr)
			
			// Should not panic and should return some result
			assert.NotEmpty(t, result)
			
			// If it's a valid format, behavior should be consistent
			if IsValid(formatStr) {
				format, _ := Parse(formatStr)
				if format == JSON {
					// Should be valid JSON
					var parsed struct {
						Response string `json:"response"`
					}
					err := json.Unmarshal([]byte(result), &parsed)
					assert.NoError(t, err)
				} else {
					// Should be plain text
					assert.Equal(t, content, result)
				}
			}
		})
	}
}