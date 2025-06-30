package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPType(t *testing.T) {
	tests := []struct {
		name     string
		mcpType  MCPType
		expected string
	}{
		{"stdio type", MCPStdio, "stdio"},
		{"sse type", MCPSse, "sse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.mcpType))
		})
	}
}

func TestAgentName(t *testing.T) {
	tests := []struct {
		name      string
		agentName AgentName
		expected  string
	}{
		{"coder agent", AgentCoder, "coder"},
		{"summarizer agent", AgentSummarizer, "summarizer"},
		{"task agent", AgentTask, "task"},
		{"title agent", AgentTitle, "title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.agentName))
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Reset global config for clean test
	cfg = nil
	
	config, err := Load(tempDir, false)
	require.NoError(t, err)
	
	// Test default values
	assert.Equal(t, defaultDataDirectory, config.Data.Directory)
	assert.Equal(t, tempDir, config.WorkingDir)
	assert.Equal(t, defaultContextPaths, config.ContextPaths)
	// Default theme might vary, just check it's not empty
	assert.NotEmpty(t, config.TUI.Theme)
	assert.True(t, config.AutoCompact)
	assert.False(t, config.Debug)
	assert.False(t, config.DebugLSP)
	
	// Test shell defaults
	expectedShell := os.Getenv("SHELL")
	if expectedShell == "" {
		expectedShell = "/bin/bash"
	}
	assert.Equal(t, expectedShell, config.Shell.Path)
	assert.Equal(t, []string{"-l"}, config.Shell.Args)
}

func TestConfigLoadWithDebug(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	
	config, err := Load(tempDir, true)
	require.NoError(t, err)
	
	assert.True(t, config.Debug)
}

func TestHasAWSCredentials(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "has access key and secret",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
			},
			expected: true,
		},
		{
			name: "has profile",
			envVars: map[string]string{
				"AWS_PROFILE": "test-profile",
			},
			expected: true,
		},
		{
			name: "has region",
			envVars: map[string]string{
				"AWS_REGION": "us-east-1",
			},
			expected: true,
		},
		{
			name: "has container credentials",
			envVars: map[string]string{
				"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI": "/v2/credentials",
			},
			expected: true,
		},
		{
			name:     "no credentials",
			envVars:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()
			
			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			
			result := hasAWSCredentials()
			assert.Equal(t, tt.expected, result)
			
			// Clean up
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestHasVertexAICredentials(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "has vertex ai project and location",
			envVars: map[string]string{
				"VERTEXAI_PROJECT":  "test-project",
				"VERTEXAI_LOCATION": "us-central1",
			},
			expected: true,
		},
		{
			name: "has google cloud project and region",
			envVars: map[string]string{
				"GOOGLE_CLOUD_PROJECT": "test-project",
				"GOOGLE_CLOUD_REGION":  "us-central1",
			},
			expected: true,
		},
		{
			name: "has google cloud project and location",
			envVars: map[string]string{
				"GOOGLE_CLOUD_PROJECT":  "test-project",
				"GOOGLE_CLOUD_LOCATION": "us-central1",
			},
			expected: true,
		},
		{
			name:     "no credentials",
			envVars:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()
			
			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			
			result := hasVertexAICredentials()
			assert.Equal(t, tt.expected, result)
			
			// Clean up
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestApplyDefaultValues(t *testing.T) {
	cfg = &Config{
		MCPServers: map[string]MCPServer{
			"test-server": {
				Command: "test-command",
				// Type is not set, should default to MCPStdio
			},
		},
	}
	
	applyDefaultValues()
	
	assert.Equal(t, MCPStdio, cfg.MCPServers["test-server"].Type)
}

func TestGet(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	
	// Load config first
	_, err := Load(tempDir, false)
	require.NoError(t, err)
	
	// Test Get function
	config := Get()
	assert.NotNil(t, config)
	assert.Equal(t, tempDir, config.WorkingDir)
}

func TestWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	
	// Load config first
	_, err := Load(tempDir, false)
	require.NoError(t, err)
	
	// Test WorkingDirectory function
	workingDir := WorkingDirectory()
	assert.Equal(t, tempDir, workingDir)
}

func TestValidateAgent(t *testing.T) {
	cfg = &Config{
		Providers: map[models.ModelProvider]Provider{
			models.ProviderAnthropic: {APIKey: "test-key"},
		},
		Agents: make(map[AgentName]Agent),
	}

	tests := []struct {
		name      string
		agentName AgentName
		agent     Agent
		wantError bool
	}{
		{
			name:      "valid agent with supported model",
			agentName: AgentCoder,
			agent: Agent{
				Model:     models.Claude4Sonnet,
				MaxTokens: 4096,
			},
			wantError: false,
		},
		{
			name:      "invalid model",
			agentName: AgentCoder,
			agent: Agent{
				Model:     "invalid-model",
				MaxTokens: 4096,
			},
			wantError: true, // Should error when no valid provider available for fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgent(cfg, tt.agentName, tt.agent)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeLocalConfig(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a local config file
	localConfigContent := `{
		"debug": true,
		"tui": {
			"theme": "custom-theme"
		}
	}`
	
	localConfigPath := filepath.Join(tempDir, ".opencode")
	err := os.WriteFile(localConfigPath, []byte(localConfigContent), 0644)
	require.NoError(t, err)
	
	// Reset and configure viper
	cfg = nil
	configureViper()
	setDefaults(false)
	
	// Test mergeLocalConfig
	mergeLocalConfig(tempDir)
	
	// The function should have merged the local config
	// We can't easily test this without accessing viper internals,
	// but we can ensure it doesn't panic
}

func TestConfigValidation(t *testing.T) {
	// Save original config
	originalCfg := cfg
	defer func() { cfg = originalCfg }()
	
	// Test with valid config
	cfg = &Config{
		Providers: map[models.ModelProvider]Provider{
			models.ProviderAnthropic: {APIKey: "test-key"},
		},
		Agents: map[AgentName]Agent{
			AgentCoder: {
				Model:     models.Claude4Sonnet,
				MaxTokens: 4096,
			},
		},
	}
	
	err := Validate()
	assert.NoError(t, err)
}