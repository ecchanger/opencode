package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("loads config with defaults", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		
		// Reset global config for clean test
		cfg = nil
		viper.Reset()
		
		config, err := Load(tempDir, false)
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, tempDir, config.WorkingDir)
		assert.Equal(t, defaultDataDirectory, config.Data.Directory)
		assert.Equal(t, "opencode", config.TUI.Theme)
		assert.True(t, config.AutoCompact)
	})

	t.Run("loads config with debug enabled", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg = nil
		viper.Reset()
		
		config, err := Load(tempDir, true)
		require.NoError(t, err)
		assert.True(t, config.Debug)
	})

	t.Run("returns existing config on second call", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg = nil
		viper.Reset()
		
		config1, err := Load(tempDir, false)
		require.NoError(t, err)
		
		config2, err := Load(tempDir, false)
		require.NoError(t, err)
		assert.Same(t, config1, config2)
	})

	t.Run("loads local config when available", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg = nil
		viper.Reset()
		
		// Create a local config file
		localConfig := `{
			"tui": {
				"theme": "gruvbox"
			},
			"debug": true
		}`
		
		configPath := filepath.Join(tempDir, ".opencode")
		err := os.WriteFile(configPath, []byte(localConfig), 0644)
		require.NoError(t, err)
		
		config, err := Load(tempDir, false)
		require.NoError(t, err)
		assert.Equal(t, "gruvbox", config.TUI.Theme)
		assert.True(t, config.Debug)
	})
}

func TestValidate(t *testing.T) {
	t.Run("validates valid config", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg = nil
		viper.Reset()
		
		// Set up a valid configuration
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")
		
		_, err := Load(tempDir, false)
		require.NoError(t, err)
		
		err = Validate()
		assert.NoError(t, err)
	})
}

func TestWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	viper.Reset()
	
	_, err := Load(tempDir, false)
	require.NoError(t, err)
	
	assert.Equal(t, tempDir, WorkingDirectory())
}

func TestGet(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	viper.Reset()
	
	config, err := Load(tempDir, false)
	require.NoError(t, err)
	
	assert.Same(t, config, Get())
}

func TestUpdateAgentModel(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	viper.Reset()
	
	_, err := Load(tempDir, false)
	require.NoError(t, err)
	
	err = UpdateAgentModel(AgentCoder, models.Claude4Sonnet)
	assert.NoError(t, err)
	
	config := Get()
	assert.Equal(t, models.Claude4Sonnet, config.Agents[AgentCoder].Model)
}

func TestUpdateTheme(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	viper.Reset()
	
	_, err := Load(tempDir, false)
	require.NoError(t, err)
	
	err = UpdateTheme("gruvbox")
	assert.NoError(t, err)
	
	config := Get()
	assert.Equal(t, "gruvbox", config.TUI.Theme)
}

func TestHasAWSCredentials(t *testing.T) {
	t.Run("detects AWS access key credentials", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
		defer func() {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}()
		
		assert.True(t, hasAWSCredentials())
	})

	t.Run("detects AWS profile", func(t *testing.T) {
		os.Setenv("AWS_PROFILE", "test-profile")
		defer os.Unsetenv("AWS_PROFILE")
		
		assert.True(t, hasAWSCredentials())
	})

	t.Run("detects AWS region", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		defer os.Unsetenv("AWS_REGION")
		
		assert.True(t, hasAWSCredentials())
	})

	t.Run("returns false when no credentials", func(t *testing.T) {
		// Ensure no AWS env vars are set
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_PROFILE")
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_DEFAULT_REGION")
		os.Unsetenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
		os.Unsetenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
		
		assert.False(t, hasAWSCredentials())
	})
}

func TestHasVertexAICredentials(t *testing.T) {
	t.Run("detects VertexAI credentials", func(t *testing.T) {
		os.Setenv("VERTEXAI_PROJECT", "test-project")
		os.Setenv("VERTEXAI_LOCATION", "us-central1")
		defer func() {
			os.Unsetenv("VERTEXAI_PROJECT")
			os.Unsetenv("VERTEXAI_LOCATION")
		}()
		
		assert.True(t, hasVertexAICredentials())
	})

	t.Run("detects Google Cloud credentials", func(t *testing.T) {
		os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
		os.Setenv("GOOGLE_CLOUD_REGION", "us-central1")
		defer func() {
			os.Unsetenv("GOOGLE_CLOUD_PROJECT")
			os.Unsetenv("GOOGLE_CLOUD_REGION")
		}()
		
		assert.True(t, hasVertexAICredentials())
	})

	t.Run("returns false when no credentials", func(t *testing.T) {
		os.Unsetenv("VERTEXAI_PROJECT")
		os.Unsetenv("VERTEXAI_LOCATION")
		os.Unsetenv("GOOGLE_CLOUD_PROJECT")
		os.Unsetenv("GOOGLE_CLOUD_REGION")
		os.Unsetenv("GOOGLE_CLOUD_LOCATION")
		
		assert.False(t, hasVertexAICredentials())
	})
}

func TestSetProviderDefaults(t *testing.T) {
	t.Run("sets Anthropic as default when API key available", func(t *testing.T) {
		cfg = nil
		viper.Reset()
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")
		
		tempDir := t.TempDir()
		config, err := Load(tempDir, false)
		require.NoError(t, err)
		
		assert.Equal(t, models.Claude4Sonnet, config.Agents[AgentCoder].Model)
	})

	t.Run("sets OpenAI as default when only OpenAI key available", func(t *testing.T) {
		cfg = nil
		viper.Reset()
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("OPENAI_API_KEY")
		
		tempDir := t.TempDir()
		config, err := Load(tempDir, false)
		require.NoError(t, err)
		
		assert.Equal(t, models.GPT41, config.Agents[AgentCoder].Model)
	})
}

func TestMCPServerDefaults(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	viper.Reset()
	
	config, err := Load(tempDir, false)
	require.NoError(t, err)
	
	// Test that MCP servers get default type
	testServer := MCPServer{
		Command: "test-command",
	}
	config.MCPServers["test"] = testServer
	applyDefaultValues()
	
	assert.Equal(t, MCPStdio, config.MCPServers["test"].Type)
}

func TestAgentDefaults(t *testing.T) {
	tempDir := t.TempDir()
	cfg = nil
	viper.Reset()
	
	config, err := Load(tempDir, false)
	require.NoError(t, err)
	
	// Check that title agent has reduced max tokens
	assert.Equal(t, int64(80), config.Agents[AgentTitle].MaxTokens)
}

func TestConfigureViper(t *testing.T) {
	viper.Reset()
	configureViper()
	
	// Verify that viper is configured correctly
	assert.Equal(t, ".opencode", viper.GetString("configname"))
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	setDefaults(false)
	
	assert.Equal(t, defaultDataDirectory, viper.GetString("data.directory"))
	assert.Equal(t, defaultContextPaths, viper.GetStringSlice("contextPaths"))
	assert.Equal(t, "opencode", viper.GetString("tui.theme"))
	assert.True(t, viper.GetBool("autoCompact"))
	assert.False(t, viper.GetBool("debug"))
	
	viper.Reset()
	setDefaults(true)
	assert.True(t, viper.GetBool("debug"))
}