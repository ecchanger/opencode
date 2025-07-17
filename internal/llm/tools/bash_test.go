package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPermissionService struct {
	requestCalls []permission.CreatePermissionRequest
	shouldAllow  bool
}

func (m *mockPermissionService) Request(req permission.CreatePermissionRequest) bool {
	m.requestCalls = append(m.requestCalls, req)
	return m.shouldAllow
}

func TestNewBashTool(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	assert.NotNil(t, tool)
	assert.IsType(t, &bashTool{}, tool)
}

func TestBashTool_Info(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	info := tool.Info()
	
	assert.Equal(t, BashToolName, info.Name)
	assert.Equal(t, "bash", info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Description, "Executes a given bash command")
	assert.Contains(t, info.Description, "banned commands")
	assert.Contains(t, info.Parameters, "command")
	assert.Contains(t, info.Parameters, "timeout")
	assert.Contains(t, info.Required, "command")
}

func TestBashTool_Run_ValidCommand(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "echo hello world",
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sessionID", "test-session")
	ctx = context.WithValue(ctx, "messageID", "test-message")
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	response, err := tool.Run(ctx, call)
	require.NoError(t, err)
	
	assert.Equal(t, "text", response.Type)
	assert.Contains(t, response.Text, "hello world")
	assert.NotNil(t, response.Metadata)
	
	// Verify permission was requested
	assert.Len(t, mockPerm.requestCalls, 1)
	assert.Equal(t, "test-session", mockPerm.requestCalls[0].SessionID)
	assert.Equal(t, BashToolName, mockPerm.requestCalls[0].ToolName)
	assert.Equal(t, "execute", mockPerm.requestCalls[0].Action)
}

func TestBashTool_Run_SafeReadOnlyCommand(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "pwd",
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sessionID", "test-session")
	ctx = context.WithValue(ctx, "messageID", "test-message")
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	response, err := tool.Run(ctx, call)
	require.NoError(t, err)
	
	assert.Equal(t, "text", response.Type)
	assert.NotEmpty(t, response.Text)
	
	// Verify permission was NOT requested for safe read-only commands
	assert.Len(t, mockPerm.requestCalls, 0)
}

func TestBashTool_Run_BannedCommand(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	bannedCommands := []string{"curl", "wget", "nc", "telnet"}
	
	for _, cmd := range bannedCommands {
		t.Run(fmt.Sprintf("banned_%s", cmd), func(t *testing.T) {
			params := BashParams{
				Command: fmt.Sprintf("%s https://example.com", cmd),
				Timeout: 5000,
			}
			
			paramsJSON, err := json.Marshal(params)
			require.NoError(t, err)
			
			ctx := context.Background()
			ctx = context.WithValue(ctx, "sessionID", "test-session")
			ctx = context.WithValue(ctx, "messageID", "test-message")
			
			call := ToolCall{
				ID:    "test-call",
				Input: string(paramsJSON),
			}
			
			response, err := tool.Run(ctx, call)
			require.NoError(t, err)
			
			assert.Equal(t, "error", response.Type)
			assert.Contains(t, response.Text, "not allowed")
			assert.Contains(t, response.Text, cmd)
		})
	}
}

func TestBashTool_Run_PermissionDenied(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: false}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "rm -rf /tmp/test",
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sessionID", "test-session")
	ctx = context.WithValue(ctx, "messageID", "test-message")
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	response, err := tool.Run(ctx, call)
	require.Error(t, err)
	assert.Equal(t, permission.ErrorPermissionDenied, err)
	
	// Verify permission was requested
	assert.Len(t, mockPerm.requestCalls, 1)
}

func TestBashTool_Run_InvalidParameters(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid json",
			input: "invalid json",
		},
		{
			name:  "empty command",
			input: `{"command": "", "timeout": 5000}`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "sessionID", "test-session")
			ctx = context.WithValue(ctx, "messageID", "test-message")
			
			call := ToolCall{
				ID:    "test-call",
				Input: tt.input,
			}
			
			response, err := tool.Run(ctx, call)
			require.NoError(t, err)
			
			assert.Equal(t, "error", response.Type)
			assert.NotEmpty(t, response.Text)
		})
	}
}

func TestBashTool_Run_TimeoutHandling(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	tests := []struct {
		name            string
		timeout         int
		expectedTimeout int
	}{
		{
			name:            "default timeout",
			timeout:         0,
			expectedTimeout: DefaultTimeout,
		},
		{
			name:            "custom timeout",
			timeout:         5000,
			expectedTimeout: 5000,
		},
		{
			name:            "max timeout exceeded",
			timeout:         MaxTimeout + 1000,
			expectedTimeout: MaxTimeout,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := BashParams{
				Command: "echo test",
				Timeout: tt.timeout,
			}
			
			paramsJSON, err := json.Marshal(params)
			require.NoError(t, err)
			
			ctx := context.Background()
			ctx = context.WithValue(ctx, "sessionID", "test-session")
			ctx = context.WithValue(ctx, "messageID", "test-message")
			
			call := ToolCall{
				ID:    "test-call",
				Input: string(paramsJSON),
			}
			
			response, err := tool.Run(ctx, call)
			require.NoError(t, err)
			
			assert.Equal(t, "text", response.Type)
			assert.Contains(t, response.Text, "test")
		})
	}
}

func TestBashTool_Run_ErrorHandling(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "false", // Command that always exits with code 1
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sessionID", "test-session")
	ctx = context.WithValue(ctx, "messageID", "test-message")
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	response, err := tool.Run(ctx, call)
	require.NoError(t, err)
	
	assert.Equal(t, "text", response.Type)
	assert.Contains(t, response.Text, "Exit code 1")
}

func TestBashTool_Run_MissingContext(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "rm -rf /tmp/test",
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	// Context without sessionID and messageID
	ctx := context.Background()
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	response, err := tool.Run(ctx, call)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session ID and message ID are required")
}

func TestBashTool_Run_NoOutput(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "true", // Command that produces no output
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sessionID", "test-session")
	ctx = context.WithValue(ctx, "messageID", "test-message")
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	response, err := tool.Run(ctx, call)
	require.NoError(t, err)
	
	assert.Equal(t, "text", response.Type)
	assert.Equal(t, "no output", response.Text)
}

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short output",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty output",
			input:    "",
			expected: "",
		},
		{
			name:  "long output",
			input: strings.Repeat("a", MaxOutputLength+1000),
			expected: func() string {
				halfLength := MaxOutputLength / 2
				start := strings.Repeat("a", halfLength)
				end := strings.Repeat("a", halfLength)
				return fmt.Sprintf("%s\n\n... [1 lines truncated] ...\n\n%s", start, end)
			}(),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "single line",
			input:    "hello world",
			expected: 1,
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: 3,
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBashResponseMetadata(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	params := BashParams{
		Command: "echo test",
		Timeout: 5000,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sessionID", "test-session")
	ctx = context.WithValue(ctx, "messageID", "test-message")
	
	call := ToolCall{
		ID:    "test-call",
		Input: string(paramsJSON),
	}
	
	startTime := time.Now()
	response, err := tool.Run(ctx, call)
	endTime := time.Now()
	
	require.NoError(t, err)
	require.NotNil(t, response.Metadata)
	
	// Parse metadata
	var metadata BashResponseMetadata
	err = json.Unmarshal(response.Metadata, &metadata)
	require.NoError(t, err)
	
	// Verify timing
	assert.GreaterOrEqual(t, metadata.StartTime, startTime.UnixMilli())
	assert.LessOrEqual(t, metadata.EndTime, endTime.UnixMilli())
	assert.GreaterOrEqual(t, metadata.EndTime, metadata.StartTime)
}

func TestSafeReadOnlyCommands(t *testing.T) {
	mockPerm := &mockPermissionService{shouldAllow: true}
	tool := NewBashTool(mockPerm)
	
	safeCommands := []string{
		"ls",
		"ls -la",
		"pwd",
		"echo hello",
		"date",
		"whoami",
		"git status",
		"git log --oneline",
		"git diff",
		"go version",
		"go help",
	}
	
	for _, cmd := range safeCommands {
		t.Run(fmt.Sprintf("safe_%s", strings.Replace(cmd, " ", "_", -1)), func(t *testing.T) {
			params := BashParams{
				Command: cmd,
				Timeout: 5000,
			}
			
			paramsJSON, err := json.Marshal(params)
			require.NoError(t, err)
			
			ctx := context.Background()
			ctx = context.WithValue(ctx, "sessionID", "test-session")
			ctx = context.WithValue(ctx, "messageID", "test-message")
			
			call := ToolCall{
				ID:    "test-call",
				Input: string(paramsJSON),
			}
			
			// Reset mock calls
			mockPerm.requestCalls = nil
			
			response, err := tool.Run(ctx, call)
			require.NoError(t, err)
			
			assert.Equal(t, "text", response.Type)
			
			// Verify permission was NOT requested for safe commands
			assert.Len(t, mockPerm.requestCalls, 0)
		})
	}
}

func TestBashConstants(t *testing.T) {
	assert.Equal(t, "bash", BashToolName)
	assert.Equal(t, 1*60*1000, DefaultTimeout)
	assert.Equal(t, 10*60*1000, MaxTimeout)
	assert.Equal(t, 30000, MaxOutputLength)
	
	// Test that banned commands list is not empty
	assert.NotEmpty(t, bannedCommands)
	assert.Contains(t, bannedCommands, "curl")
	assert.Contains(t, bannedCommands, "wget")
	
	// Test that safe commands list is not empty
	assert.NotEmpty(t, safeReadOnlyCommands)
	assert.Contains(t, safeReadOnlyCommands, "ls")
	assert.Contains(t, safeReadOnlyCommands, "pwd")
}

func TestBashDescription(t *testing.T) {
	description := bashDescription()
	
	assert.NotEmpty(t, description)
	assert.Contains(t, description, "Executes a given bash command")
	assert.Contains(t, description, "banned commands")
	assert.Contains(t, description, "curl")
	assert.Contains(t, description, "wget")
	assert.Contains(t, description, fmt.Sprintf("%d characters", MaxOutputLength))
	assert.Contains(t, description, "git commit")
	assert.Contains(t, description, "pull request")
}

func TestBashParamsStruct(t *testing.T) {
	params := BashParams{
		Command: "echo test",
		Timeout: 5000,
	}
	
	assert.Equal(t, "echo test", params.Command)
	assert.Equal(t, 5000, params.Timeout)
	
	// Test JSON marshaling
	data, err := json.Marshal(params)
	require.NoError(t, err)
	
	var unmarshaled BashParams
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, params.Command, unmarshaled.Command)
	assert.Equal(t, params.Timeout, unmarshaled.Timeout)
}

func TestBashPermissionsParamsStruct(t *testing.T) {
	params := BashPermissionsParams{
		Command: "echo test",
		Timeout: 5000,
	}
	
	assert.Equal(t, "echo test", params.Command)
	assert.Equal(t, 5000, params.Timeout)
	
	// Test JSON marshaling
	data, err := json.Marshal(params)
	require.NoError(t, err)
	
	var unmarshaled BashPermissionsParams
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, params.Command, unmarshaled.Command)
	assert.Equal(t, params.Timeout, unmarshaled.Timeout)
}