package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Reset global config before and after tests
	originalCfg := cfg
	defer func() {
		cfg = originalCfg
	}()
	
	code := m.Run()
	os.Exit(code)
}

func resetConfig() {
	cfg = nil
}

func TestMCPType(t *testing.T) {
	tests := []struct {
		name     string
		mcpType  MCPType
		expected string
	}{
		{
			name:     "stdio type",
			mcpType:  MCPStdio,
			expected: "stdio",
		},
		{
			name:     "sse type",
			mcpType:  MCPSse,
			expected: "sse",
		},
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
		{
			name:      "coder agent",
			agentName: AgentCoder,
			expected:  "coder",
		},
		{
			name:      "summarizer agent",
			agentName: AgentSummarizer,
			expected:  "summarizer",
		},
		{
			name:      "task agent",
			agentName: AgentTask,
			expected:  "task",
		},
		{
			name:      "title agent",
			agentName: AgentTitle,
			expected:  "title",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.agentName))
		})
	}
}

func TestLoadConfig(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Test loading with debug=false
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify default values
	assert.Equal(t, tmpDir, config.WorkingDir)
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, "opencode", config.TUI.Theme)
	assert.Equal(t, true, config.AutoCompact)
	assert.NotNil(t, config.MCPServers)
	assert.NotNil(t, config.Providers)
	assert.NotNil(t, config.LSP)
	assert.NotNil(t, config.Agents)
	
	// Test context paths default
	assert.Contains(t, config.ContextPaths, "CLAUDE.md")
	assert.Contains(t, config.ContextPaths, ".cursorrules")
	assert.Contains(t, config.ContextPaths, "opencode.md")
	
	// Test shell defaults
	assert.NotEmpty(t, config.Shell.Path)
	assert.Contains(t, config.Shell.Args, "-l")
	
	// Test that title agent has correct max tokens
	if titleAgent, ok := config.Agents[AgentTitle]; ok {
		assert.Equal(t, int64(80), titleAgent.MaxTokens)
	}
}

func TestLoadConfigWithDebug(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Test loading with debug=true
	config, err := Load(tmpDir, true)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify debug is enabled
	assert.Equal(t, true, config.Debug)
}

func TestLoadConfigWithLocalConfig(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create a local config file
	localConfig := map[string]interface{}{
		"debug": true,
		"tui": map[string]interface{}{
			"theme": "dark",
		},
		"contextPaths": []string{
			"custom.md",
		},
	}
	
	localConfigData, err := json.Marshal(localConfig)
	require.NoError(t, err)
	
	localConfigPath := filepath.Join(tmpDir, ".opencode.json")
	err = os.WriteFile(localConfigPath, localConfigData, 0644)
	require.NoError(t, err)
	
	// Load config (should merge local config)
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify local config was merged
	assert.Equal(t, true, config.Debug)
	assert.Equal(t, "dark", config.TUI.Theme)
	assert.Contains(t, config.ContextPaths, "custom.md")
}

func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	resetConfig()
	
	// Set environment variables
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	defer func() {
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("OPENAI_API_KEY")
	}()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Load config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify provider configuration
	anthropicProvider, ok := config.Providers[models.ProviderAnthropic]
	assert.True(t, ok)
	assert.Equal(t, "test-anthropic-key", anthropicProvider.APIKey)
	
	// Verify default agent models (should prefer Anthropic)
	coderAgent, ok := config.Agents[AgentCoder]
	assert.True(t, ok)
	assert.Equal(t, models.Claude4Sonnet, coderAgent.Model)
}

func TestHasAWSCredentials(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "with access key and secret",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test-access-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret-key",
			},
			expected: true,
		},
		{
			name: "with profile",
			envVars: map[string]string{
				"AWS_PROFILE": "test-profile",
			},
			expected: true,
		},
		{
			name: "with region",
			envVars: map[string]string{
				"AWS_REGION": "us-west-2",
			},
			expected: true,
		},
		{
			name: "with container credentials",
			envVars: map[string]string{
				"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI": "/v2/credentials/test",
			},
			expected: true,
		},
		{
			name:     "without credentials",
			envVars:  map[string]string{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			
			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()
			
			result := hasAWSCredentials()
			assert.Equal(t, tt.expected, result)
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
			name: "with VertexAI project and location",
			envVars: map[string]string{
				"VERTEXAI_PROJECT":  "test-project",
				"VERTEXAI_LOCATION": "us-central1",
			},
			expected: true,
		},
		{
			name: "with Google Cloud project and region",
			envVars: map[string]string{
				"GOOGLE_CLOUD_PROJECT": "test-project",
				"GOOGLE_CLOUD_REGION":  "us-central1",
			},
			expected: true,
		},
		{
			name: "with Google Cloud project and location",
			envVars: map[string]string{
				"GOOGLE_CLOUD_PROJECT":  "test-project",
				"GOOGLE_CLOUD_LOCATION": "us-central1",
			},
			expected: true,
		},
		{
			name:     "without credentials",
			envVars:  map[string]string{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			
			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()
			
			result := hasVertexAICredentials()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAgent(t *testing.T) {
	resetConfig()
	
	// Setup test config
	testConfig := &Config{
		Agents: make(map[AgentName]Agent),
		Providers: map[models.ModelProvider]Provider{
			models.ProviderAnthropic: {
				APIKey:   "test-key",
				Disabled: false,
			},
		},
	}
	cfg = testConfig
	
	tests := []struct {
		name        string
		agentName   AgentName
		agent       Agent
		expectError bool
		expectWarn  bool
	}{
		{
			name:      "valid agent configuration",
			agentName: AgentCoder,
			agent: Agent{
				Model:     models.Claude37Sonnet,
				MaxTokens: 4000,
			},
			expectError: false,
			expectWarn:  false,
		},
		{
			name:      "invalid max tokens",
			agentName: AgentCoder,
			agent: Agent{
				Model:     models.Claude37Sonnet,
				MaxTokens: 0,
			},
			expectError: false,
			expectWarn:  true,
		},
		{
			name:      "unsupported model",
			agentName: AgentCoder,
			agent: Agent{
				Model:     "unsupported-model",
				MaxTokens: 4000,
			},
			expectError: false,
			expectWarn:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the agent in config
			testConfig.Agents[tt.agentName] = tt.agent
			
			err := validateAgent(testConfig, tt.agentName, tt.agent)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			// If max tokens was invalid, it should have been fixed
			if tt.agent.MaxTokens <= 0 {
				updatedAgent := testConfig.Agents[tt.agentName]
				assert.Greater(t, updatedAgent.MaxTokens, int64(0))
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	// Create a test config with various configurations
	testConfig := &Config{
		Agents: map[AgentName]Agent{
			AgentCoder: {
				Model:     models.Claude37Sonnet,
				MaxTokens: 4000,
			},
		},
		Providers: map[models.ModelProvider]Provider{
			models.ProviderAnthropic: {
				APIKey:   "test-key",
				Disabled: false,
			},
			models.ProviderOpenAI: {
				APIKey:   "",
				Disabled: false,
			},
		},
		LSP: map[string]LSPConfig{
			"go": {
				Command:  "gopls",
				Args:     []string{"serve"},
				Disabled: false,
			},
			"python": {
				Command:  "",
				Disabled: false,
			},
		},
	}
	cfg = testConfig
	
	err := Validate()
	assert.NoError(t, err)
	
	// Verify that empty API key provider was disabled
	openaiProvider := testConfig.Providers[models.ProviderOpenAI]
	assert.True(t, openaiProvider.Disabled)
	
	// Verify that empty command LSP was disabled
	pythonLSP := testConfig.LSP["python"]
	assert.True(t, pythonLSP.Disabled)
}

func TestGetProviderAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider models.ModelProvider
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "anthropic provider",
			provider: models.ProviderAnthropic,
			envVar:   "ANTHROPIC_API_KEY",
			envValue: "test-anthropic-key",
			expected: "test-anthropic-key",
		},
		{
			name:     "openai provider",
			provider: models.ProviderOpenAI,
			envVar:   "OPENAI_API_KEY",
			envValue: "test-openai-key",
			expected: "test-openai-key",
		},
		{
			name:     "gemini provider",
			provider: models.ProviderGemini,
			envVar:   "GEMINI_API_KEY",
			envValue: "test-gemini-key",
			expected: "test-gemini-key",
		},
		{
			name:     "groq provider",
			provider: models.ProviderGROQ,
			envVar:   "GROQ_API_KEY",
			envValue: "test-groq-key",
			expected: "test-groq-key",
		},
		{
			name:     "azure provider",
			provider: models.ProviderAzure,
			envVar:   "AZURE_OPENAI_API_KEY",
			envValue: "test-azure-key",
			expected: "test-azure-key",
		},
		{
			name:     "openrouter provider",
			provider: models.ProviderOpenRouter,
			envVar:   "OPENROUTER_API_KEY",
			envValue: "test-openrouter-key",
			expected: "test-openrouter-key",
		},
		{
			name:     "provider without env var",
			provider: models.ProviderAnthropic,
			envVar:   "",
			envValue: "",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}
			
			result := getProviderAPIKey(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateAgentModel(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Set up environment for testing
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	// Load initial config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Test updating agent model
	err = UpdateAgentModel(AgentCoder, models.Claude37Sonnet)
	assert.NoError(t, err)
	
	// Verify the agent was updated
	updatedAgent := config.Agents[AgentCoder]
	assert.Equal(t, models.Claude37Sonnet, updatedAgent.Model)
	
	// Test updating with unsupported model
	err = UpdateAgentModel(AgentCoder, "unsupported-model")
	assert.Error(t, err)
}

func TestUpdateTheme(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Load initial config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Test updating theme
	err = UpdateTheme("dark")
	assert.NoError(t, err)
	
	// Verify the theme was updated
	assert.Equal(t, "dark", config.TUI.Theme)
}

func TestGet(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Load config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Test Get function
	retrievedConfig := Get()
	assert.Equal(t, config, retrievedConfig)
}

func TestWorkingDirectory(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Load config
	_, err = Load(tmpDir, false)
	require.NoError(t, err)
	
	// Test WorkingDirectory function
	workingDir := WorkingDirectory()
	assert.Equal(t, tmpDir, workingDir)
}

func TestWorkingDirectoryPanic(t *testing.T) {
	resetConfig()
	
	// Test that WorkingDirectory panics when config is not loaded
	assert.Panics(t, func() {
		WorkingDirectory()
	})
}

func TestMCPServerDefaults(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create config with MCP server without type
	localConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "test-command",
				"args":    []string{"arg1", "arg2"},
			},
		},
	}
	
	localConfigData, err := json.Marshal(localConfig)
	require.NoError(t, err)
	
	localConfigPath := filepath.Join(tmpDir, ".opencode.json")
	err = os.WriteFile(localConfigPath, localConfigData, 0644)
	require.NoError(t, err)
	
	// Load config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify MCP server got default type
	mcpServer, ok := config.MCPServers["test-server"]
	assert.True(t, ok)
	assert.Equal(t, MCPStdio, mcpServer.Type)
	assert.Equal(t, "test-command", mcpServer.Command)
	assert.Equal(t, []string{"arg1", "arg2"}, mcpServer.Args)
}

func TestLoadGitHubToken(t *testing.T) {
	// Test with environment variable
	os.Setenv("GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("GITHUB_TOKEN")
	
	token, err := LoadGitHubToken()
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)
	
	// Test without environment variable
	os.Unsetenv("GITHUB_TOKEN")
	
	// Create a temporary config directory
	tmpDir, err := os.MkdirTemp("", "github-copilot-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create copilot config directory
	copilotDir := filepath.Join(tmpDir, "github-copilot")
	err = os.MkdirAll(copilotDir, 0755)
	require.NoError(t, err)
	
	// Create hosts.json file
	hostsConfig := map[string]map[string]interface{}{
		"github.com": {
			"oauth_token": "test-oauth-token",
		},
	}
	
	hostsData, err := json.Marshal(hostsConfig)
	require.NoError(t, err)
	
	hostsPath := filepath.Join(copilotDir, "hosts.json")
	err = os.WriteFile(hostsPath, hostsData, 0644)
	require.NoError(t, err)
	
	// Set XDG_CONFIG_HOME to our test directory
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	
	token, err = LoadGitHubToken()
	assert.NoError(t, err)
	assert.Equal(t, "test-oauth-token", token)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, ".opencode", defaultDataDirectory)
	assert.Equal(t, "info", defaultLogLevel)
	assert.Equal(t, "opencode", appName)
	assert.Equal(t, int64(4096), MaxTokensFallbackDefault)
	
	// Test default context paths
	assert.Contains(t, defaultContextPaths, "CLAUDE.md")
	assert.Contains(t, defaultContextPaths, ".cursorrules")
	assert.Contains(t, defaultContextPaths, "opencode.md")
	assert.Contains(t, defaultContextPaths, ".github/copilot-instructions.md")
}

func TestDataDirectory(t *testing.T) {
	resetConfig()
	
	// Test with OPENCODE_DEV_DEBUG environment variable
	os.Setenv("OPENCODE_DEV_DEBUG", "true")
	defer os.Unsetenv("OPENCODE_DEV_DEBUG")
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create config with custom data directory
	localConfig := map[string]interface{}{
		"data": map[string]interface{}{
			"directory": filepath.Join(tmpDir, "custom-data"),
		},
	}
	
	localConfigData, err := json.Marshal(localConfig)
	require.NoError(t, err)
	
	localConfigPath := filepath.Join(tmpDir, ".opencode.json")
	err = os.WriteFile(localConfigPath, localConfigData, 0644)
	require.NoError(t, err)
	
	// Load config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify custom data directory
	expectedDataDir := filepath.Join(tmpDir, "custom-data")
	assert.Equal(t, expectedDataDir, config.Data.Directory)
	
	// Verify debug.log and messages directory were created
	assert.FileExists(t, filepath.Join(expectedDataDir, "debug.log"))
	assert.DirExists(t, filepath.Join(expectedDataDir, "messages"))
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid config",
			config: &Config{
				Agents: map[AgentName]Agent{
					AgentCoder: {
						Model:     models.Claude37Sonnet,
						MaxTokens: 4000,
					},
				},
				Providers: map[models.ModelProvider]Provider{
					models.ProviderAnthropic: {
						APIKey:   "test-key",
						Disabled: false,
					},
				},
				LSP: map[string]LSPConfig{
					"go": {
						Command:  "gopls",
						Disabled: false,
					},
				},
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = tt.config
			
			err := Validate()
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConcurrentConfigAccess(t *testing.T) {
	resetConfig()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "opencode-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Load config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Test concurrent access to Get()
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				retrievedConfig := Get()
				assert.Equal(t, config, retrievedConfig)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}