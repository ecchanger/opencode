package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, MCPType("stdio"), MCPStdio)
	assert.Equal(t, MCPType("sse"), MCPSse)
}

func TestAgentName_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, AgentName("coder"), AgentCoder)
	assert.Equal(t, AgentName("summarizer"), AgentSummarizer)
	assert.Equal(t, AgentName("task"), AgentTask)
	assert.Equal(t, AgentName("title"), AgentTitle)
}

func TestApplicationConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ".opencode", defaultDataDirectory)
	assert.Equal(t, "info", defaultLogLevel)
	assert.Equal(t, "opencode", appName)
	assert.Equal(t, int64(4096), int64(MaxTokensFallbackDefault))
	assert.NotEmpty(t, defaultContextPaths)
	assert.Contains(t, defaultContextPaths, ".cursorrules")
	assert.Contains(t, defaultContextPaths, "opencode.md")
}

func TestMCPServer_Struct(t *testing.T) {
	t.Parallel()

	server := MCPServer{
		Command: "python",
		Env:     []string{"PATH=/usr/bin"},
		Args:    []string{"--help"},
		Type:    MCPStdio,
		URL:     "http://localhost:8080",
		Headers: map[string]string{"Content-Type": "application/json"},
	}

	assert.Equal(t, "python", server.Command)
	assert.Equal(t, []string{"PATH=/usr/bin"}, server.Env)
	assert.Equal(t, []string{"--help"}, server.Args)
	assert.Equal(t, MCPStdio, server.Type)
	assert.Equal(t, "http://localhost:8080", server.URL)
	assert.Equal(t, "application/json", server.Headers["Content-Type"])
}

func TestAgent_Struct(t *testing.T) {
	t.Parallel()

	agent := Agent{
		Model:           models.Claude37Sonnet,
		MaxTokens:       4000,
		ReasoningEffort: "medium",
	}

	assert.Equal(t, models.Claude37Sonnet, agent.Model)
	assert.Equal(t, int64(4000), agent.MaxTokens)
	assert.Equal(t, "medium", agent.ReasoningEffort)
}

func TestProvider_Struct(t *testing.T) {
	t.Parallel()

	provider := Provider{
		APIKey:   "test-api-key",
		Disabled: false,
	}

	assert.Equal(t, "test-api-key", provider.APIKey)
	assert.False(t, provider.Disabled)
}

func TestConfig_Struct(t *testing.T) {
	t.Parallel()

	config := Config{
		Data: Data{
			Directory: "/tmp/test",
		},
		WorkingDir:   "/work",
		MCPServers:   make(map[string]MCPServer),
		Providers:    make(map[models.ModelProvider]Provider),
		LSP:          make(map[string]LSPConfig),
		Agents:       make(map[AgentName]Agent),
		Debug:        true,
		DebugLSP:     false,
		ContextPaths: []string{".cursorrules"},
		TUI: TUIConfig{
			Theme: "dark",
		},
		Shell: ShellConfig{
			Path: "/bin/bash",
			Args: []string{"-l"},
		},
		AutoCompact: true,
	}

	assert.Equal(t, "/tmp/test", config.Data.Directory)
	assert.Equal(t, "/work", config.WorkingDir)
	assert.True(t, config.Debug)
	assert.False(t, config.DebugLSP)
	assert.True(t, config.AutoCompact)
	assert.Equal(t, "dark", config.TUI.Theme)
	assert.Equal(t, "/bin/bash", config.Shell.Path)
	assert.Equal(t, []string{"-l"}, config.Shell.Args)
}

func TestHasAWSCredentials(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setup    func()
		cleanup  func()
		expected bool
	}{
		{
			name: "有AWS访问密钥",
			setup: func() {
				os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
				os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
			},
			cleanup: func() {
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			},
			expected: true,
		},
		{
			name: "有AWS配置文件",
			setup: func() {
				os.Setenv("AWS_PROFILE", "test-profile")
			},
			cleanup: func() {
				os.Unsetenv("AWS_PROFILE")
			},
			expected: true,
		},
		{
			name: "有AWS区域",
			setup: func() {
				os.Setenv("AWS_REGION", "us-east-1")
			},
			cleanup: func() {
				os.Unsetenv("AWS_REGION")
			},
			expected: true,
		},
		{
			name: "有容器凭证",
			setup: func() {
				os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "/v2/credentials/test")
			},
			cleanup: func() {
				os.Unsetenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
			},
			expected: true,
		},
		{
			name: "无AWS凭证",
			setup: func() {
				// 确保清理所有AWS环境变量
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				os.Unsetenv("AWS_PROFILE")
				os.Unsetenv("AWS_DEFAULT_PROFILE")
				os.Unsetenv("AWS_REGION")
				os.Unsetenv("AWS_DEFAULT_REGION")
				os.Unsetenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
				os.Unsetenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
			},
			cleanup:  func() {},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			defer tc.cleanup()

			result := hasAWSCredentials()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHasVertexAICredentials(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setup    func()
		cleanup  func()
		expected bool
	}{
		{
			name: "有VertexAI凭证",
			setup: func() {
				os.Setenv("VERTEXAI_PROJECT", "test-project")
				os.Setenv("VERTEXAI_LOCATION", "us-central1")
			},
			cleanup: func() {
				os.Unsetenv("VERTEXAI_PROJECT")
				os.Unsetenv("VERTEXAI_LOCATION")
			},
			expected: true,
		},
		{
			name: "有Google Cloud凭证",
			setup: func() {
				os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
				os.Setenv("GOOGLE_CLOUD_REGION", "us-central1")
			},
			cleanup: func() {
				os.Unsetenv("GOOGLE_CLOUD_PROJECT")
				os.Unsetenv("GOOGLE_CLOUD_REGION")
			},
			expected: true,
		},
		{
			name: "无VertexAI凭证",
			setup: func() {
				os.Unsetenv("VERTEXAI_PROJECT")
				os.Unsetenv("VERTEXAI_LOCATION")
				os.Unsetenv("GOOGLE_CLOUD_PROJECT")
				os.Unsetenv("GOOGLE_CLOUD_REGION")
				os.Unsetenv("GOOGLE_CLOUD_LOCATION")
			},
			cleanup:  func() {},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			defer tc.cleanup()

			result := hasVertexAICredentials()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetProviderAPIKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		provider models.ModelProvider
		envKey   string
		envValue string
		expected string
	}{
		{
			name:     "Anthropic API Key",
			provider: models.ProviderAnthropic,
			envKey:   "ANTHROPIC_API_KEY",
			envValue: "test-anthropic-key",
			expected: "test-anthropic-key",
		},
		{
			name:     "OpenAI API Key",
			provider: models.ProviderOpenAI,
			envKey:   "OPENAI_API_KEY",
			envValue: "test-openai-key",
			expected: "test-openai-key",
		},
		{
			name:     "Gemini API Key",
			provider: models.ProviderGemini,
			envKey:   "GEMINI_API_KEY",
			envValue: "test-gemini-key",
			expected: "test-gemini-key",
		},
		{
			name:     "Groq API Key",
			provider: models.ProviderGROQ,
			envKey:   "GROQ_API_KEY",
			envValue: "test-groq-key",
			expected: "test-groq-key",
		},
		{
			name:     "Azure API Key",
			provider: models.ProviderAzure,
			envKey:   "AZURE_OPENAI_API_KEY",
			envValue: "test-azure-key",
			expected: "test-azure-key",
		},
		{
			name:     "OpenRouter API Key",
			provider: models.ProviderOpenRouter,
			envKey:   "OPENROUTER_API_KEY",
			envValue: "test-openrouter-key",
			expected: "test-openrouter-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置环境变量
			os.Setenv(tc.envKey, tc.envValue)
			defer os.Unsetenv(tc.envKey)

			result := getProviderAPIKey(tc.provider)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetProviderAPIKey_NoEnvVar(t *testing.T) {
	t.Parallel()

	// 确保环境变量未设置
	os.Unsetenv("ANTHROPIC_API_KEY")

	result := getProviderAPIKey(models.ProviderAnthropic)
	assert.Empty(t, result)
}

func TestWorkingDirectory_WithConfig(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置测试配置
	cfg = &Config{
		WorkingDir: "/test/dir",
	}

	result := WorkingDirectory()
	assert.Equal(t, "/test/dir", result)
}

func TestWorkingDirectory_WithoutConfig(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { 
		cfg = originalCfg
		// 从panic中恢复
		if r := recover(); r != nil {
			assert.Contains(t, r.(string), "config not loaded")
		}
	}()

	// 设置nil配置
	cfg = nil

	// 这应该panic
	assert.Panics(t, func() {
		WorkingDirectory()
	})
}

func TestGet_ReturnsConfig(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置测试配置
	testCfg := &Config{
		WorkingDir: "/test",
	}
	cfg = testCfg

	result := Get()
	assert.Equal(t, testCfg, result)
}

func TestGet_NilConfig(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置nil配置
	cfg = nil

	result := Get()
	assert.Nil(t, result)
}

func TestApplyDefaultValues(t *testing.T) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// 设置测试配置
	cfg = &Config{
		MCPServers: map[string]MCPServer{
			"test-server": {
				Command: "test",
				// Type 故意留空来测试默认值
			},
		},
	}

	applyDefaultValues()

	assert.Equal(t, MCPStdio, cfg.MCPServers["test-server"].Type)
}

// 测试LoadGitHubToken函数（如果环境变量存在）
func TestLoadGitHubToken_FromEnv(t *testing.T) {
	t.Parallel()

	// 设置环境变量
	os.Setenv("GITHUB_TOKEN", "test-github-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	token, err := LoadGitHubToken()
	assert.NoError(t, err)
	assert.Equal(t, "test-github-token", token)
}

func TestLoadGitHubToken_NoToken(t *testing.T) {
	t.Parallel()

	// 确保环境变量未设置
	os.Unsetenv("GITHUB_TOKEN")

	token, err := LoadGitHubToken()
	// 如果没有token文件，应该返回错误
	assert.Error(t, err)
	assert.Empty(t, token)
}

// 基准测试
func BenchmarkGet(b *testing.B) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	cfg = &Config{
		WorkingDir: "/test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get()
	}
}

func BenchmarkWorkingDirectory(b *testing.B) {
	// 备份原始配置
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	cfg = &Config{
		WorkingDir: "/test/directory",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WorkingDirectory()
	}
}

func BenchmarkHasAWSCredentials(b *testing.B) {
	// 设置测试环境
	os.Setenv("AWS_REGION", "us-east-1")
	defer os.Unsetenv("AWS_REGION")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasAWSCredentials()
	}
}

func BenchmarkGetProviderAPIKey(b *testing.B) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getProviderAPIKey(models.ProviderAnthropic)
	}
}

// 结构体序列化测试
func TestConfigSerialization(t *testing.T) {
	t.Parallel()

	config := Config{
		Data: Data{
			Directory: "/test/data",
		},
		WorkingDir: "/test/work",
		Debug:      true,
		TUI: TUIConfig{
			Theme: "dark",
		},
		Shell: ShellConfig{
			Path: "/bin/bash",
			Args: []string{"-l"},
		},
	}

	// 测试JSON序列化
	data, err := json.Marshal(config)
	require.NoError(t, err)
	assert.Contains(t, string(data), "/test/data")
	assert.Contains(t, string(data), "dark")

	// 测试JSON反序列化
	var deserializedConfig Config
	err = json.Unmarshal(data, &deserializedConfig)
	require.NoError(t, err)
	assert.Equal(t, config.Data.Directory, deserializedConfig.Data.Directory)
	assert.Equal(t, config.TUI.Theme, deserializedConfig.TUI.Theme)
	assert.Equal(t, config.Debug, deserializedConfig.Debug)
}