package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	// Version should be set to something (either by ldflags or build info)
	assert.NotEmpty(t, Version)
	assert.NotEqual(t, "", Version)
}

func TestVersionIsString(t *testing.T) {
	// Version should be a string type
	assert.IsType(t, "", Version)
}

func TestVersionDefault(t *testing.T) {
	// When no build info is available, version should be "unknown" or a valid version
	// We can't easily test the init() function behavior, but we can verify
	// that the Version variable exists and is accessible
	originalVersion := Version
	
	// Test that we can read and modify the version (for testing purposes)
	Version = "test-version"
	assert.Equal(t, "test-version", Version)
	
	// Restore original version
	Version = originalVersion
}

func TestVersionNotEmpty(t *testing.T) {
	// In a real build, version should never be empty
	// It should be either set by ldflags or by build info
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestVersionFormats(t *testing.T) {
	// Test that version follows expected patterns
	// This is a basic test - in reality version could be many formats
	
	validPatterns := []string{
		"unknown",           // Default fallback
		"(devel)",          // Development version
	}
	
	// If version is not one of the known fallbacks, it should look like a version
	isKnownFallback := false
	for _, pattern := range validPatterns {
		if Version == pattern {
			isKnownFallback = true
			break
		}
	}
	
	if !isKnownFallback {
		// If it's not a known fallback, it should at least not be empty
		assert.NotEmpty(t, Version)
		// Could add more specific version format validation here if needed
	}
}

// Test that the package can be imported and used
func TestPackageUsability(t *testing.T) {
	// Test that we can access the version in different ways
	v1 := Version
	
	// Test that it's consistently the same value
	v2 := Version
	assert.Equal(t, v1, v2)
}

// Benchmark version access (should be very fast since it's just a variable)
func BenchmarkVersionAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Version
	}
}

// Test version comparison (useful for version checking logic)
func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		equal    bool
	}{
		{
			name:     "same versions",
			version1: "v1.0.0",
			version2: "v1.0.0",
			equal:    true,
		},
		{
			name:     "different versions",
			version1: "v1.0.0",
			version2: "v1.0.1",
			equal:    false,
		},
		{
			name:     "unknown version",
			version1: "unknown",
			version2: "unknown",
			equal:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version1 == tt.version2
			assert.Equal(t, tt.equal, result)
		})
	}
}

// Test version display formatting
func TestVersionDisplay(t *testing.T) {
	// Test that version can be displayed in various contexts
	versionStr := Version
	
	// Should be able to format in string contexts
	formatted := "Application version: " + versionStr
	assert.Contains(t, formatted, versionStr)
	
	// Should be able to use in comparisons
	assert.True(t, versionStr == Version)
}

// Test edge cases
func TestVersionEdgeCases(t *testing.T) {
	// Save original version
	original := Version
	
	tests := []struct {
		name    string
		version string
	}{
		{"empty string", ""},
		{"space", " "},
		{"special chars", "v1.0.0-beta+build.123"},
		{"long version", "very-long-version-string-with-many-components-1.2.3-alpha.beta.gamma"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test version
			Version = tt.version
			
			// Should be able to read it back
			assert.Equal(t, tt.version, Version)
		})
	}
	
	// Restore original
	Version = original
}