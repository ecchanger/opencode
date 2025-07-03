package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCaller(t *testing.T) {
	t.Parallel()

	caller := getCaller()
	assert.NotEmpty(t, caller)
	assert.Contains(t, caller, ".go")
	assert.Contains(t, caller, ":")
}

func TestLoggingFunctions(t *testing.T) {
	// 这些测试需要捕获日志输出来验证
	// 由于slog使用全局状态，我们需要小心处理

	// 创建临时的日志输出缓冲区
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 设置临时的日志handler
	originalHandler := slog.Default()
	defer slog.SetDefault(originalHandler)

	handler := slog.NewTextHandler(tmpFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	testCases := []struct {
		name     string
		logFunc  func(string, ...any)
		message  string
		args     []any
		level    string
	}{
		{
			name:     "Info日志",
			logFunc:  Info,
			message:  "test info message",
			args:     []any{"key", "value"},
			level:    "INFO",
		},
		{
			name:     "Debug日志",
			logFunc:  Debug,
			message:  "test debug message",
			args:     []any{"debug_key", "debug_value"},
			level:    "DEBUG",
		},
		{
			name:     "Warn日志",
			logFunc:  Warn,
			message:  "test warn message",
			args:     []any{"warn_key", "warn_value"},
			level:    "WARN",
		},
		{
			name:     "Error日志",
			logFunc:  Error,
			message:  "test error message",
			args:     []any{"error_key", "error_value"},
			level:    "ERROR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 记录当前文件位置
			currentPos, err := tmpFile.Seek(0, io.SeekCurrent)
			require.NoError(t, err)

			// 调用日志函数
			tc.logFunc(tc.message, tc.args...)

			// 读取新写入的内容
			tmpFile.Sync() // 确保写入磁盘
			_, err = tmpFile.Seek(currentPos, io.SeekStart)
			require.NoError(t, err)

			logOutput, err := io.ReadAll(tmpFile)
			require.NoError(t, err)

			logStr := string(logOutput)
			assert.Contains(t, logStr, tc.message)
			assert.Contains(t, logStr, tc.level)
			assert.Contains(t, logStr, "source")
		})
	}
}

func TestPersistLoggingFunctions(t *testing.T) {
	// 测试带有persist标志的日志函数
	tmpFile, err := os.CreateTemp("", "test-persist-log-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 设置临时的日志handler
	originalHandler := slog.Default()
	defer slog.SetDefault(originalHandler)

	handler := slog.NewTextHandler(tmpFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	testCases := []struct {
		name      string
		logFunc   func(string, ...any)
		message   string
		level     string
		hasPersist bool
	}{
		{
			name:      "InfoPersist日志",
			logFunc:   InfoPersist,
			message:   "persist info message",
			level:     "INFO",
			hasPersist: true,
		},
		{
			name:      "DebugPersist日志",
			logFunc:   DebugPersist,
			message:   "persist debug message",
			level:     "DEBUG",
			hasPersist: true,
		},
		{
			name:      "WarnPersist日志",
			logFunc:   WarnPersist,
			message:   "persist warn message",
			level:     "WARN",
			hasPersist: true,
		},
		{
			name:      "ErrorPersist日志",
			logFunc:   ErrorPersist,
			message:   "persist error message",
			level:     "ERROR",
			hasPersist: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 记录当前文件位置
			currentPos, err := tmpFile.Seek(0, io.SeekCurrent)
			require.NoError(t, err)

			// 调用日志函数
			tc.logFunc(tc.message, "test_key", "test_value")

			// 读取新写入的内容
			tmpFile.Sync()
			_, err = tmpFile.Seek(currentPos, io.SeekStart)
			require.NoError(t, err)

			logOutput, err := io.ReadAll(tmpFile)
			require.NoError(t, err)

			logStr := string(logOutput)
			assert.Contains(t, logStr, tc.message)
			assert.Contains(t, logStr, tc.level)
			if tc.hasPersist {
				assert.Contains(t, logStr, persistKeyArg)
			}
		})
	}
}

func TestRecoverPanic_WithPanic(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir, err := os.MkdirTemp("", "panic-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 切换到临时目录
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// 设置临时日志
	tmpFile, err := os.CreateTemp(tmpDir, "test-panic-log-*.log")
	require.NoError(t, err)
	defer tmpFile.Close()

	originalHandler := slog.Default()
	defer slog.SetDefault(originalHandler)

	handler := slog.NewTextHandler(tmpFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// 测试cleanup函数是否被调用
	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}

	// 模拟panic
	func() {
		defer RecoverPanic("test-function", cleanup)
		panic("test panic message")
	}()

	// 验证cleanup被调用
	assert.True(t, cleanupCalled)

	// 验证panic日志文件被创建
	files, err := filepath.Glob("opencode-panic-test-function-*.log")
	require.NoError(t, err)
	assert.NotEmpty(t, files, "应该创建panic日志文件")

	if len(files) > 0 {
		// 验证panic日志文件内容
		content, err := os.ReadFile(files[0])
		require.NoError(t, err)
		
		contentStr := string(content)
		assert.Contains(t, contentStr, "Panic in test-function")
		assert.Contains(t, contentStr, "test panic message")
		assert.Contains(t, contentStr, "Stack Trace:")
		assert.Contains(t, contentStr, "Time:")
	}
}

func TestRecoverPanic_NoPanic(t *testing.T) {
	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}

	// 正常执行，没有panic
	func() {
		defer RecoverPanic("test-function", cleanup)
		// 正常代码
	}()

	// cleanup不应该被调用
	assert.False(t, cleanupCalled)
}

func TestRecoverPanic_NilCleanup(t *testing.T) {
	// 测试cleanup为nil的情况
	func() {
		defer RecoverPanic("test-function", nil)
		panic("test panic")
	}()

	// 应该不会崩溃，即使cleanup为nil
}

func TestGetSessionPrefix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		sessionID string
		expected  string
	}{
		{
			sessionID: "12345678901234567890",
			expected:  "12345678",
		},
		{
			sessionID: "abcdefgh",
			expected:  "abcdefgh",
		},
		{
			sessionID: "short",
			expected:  "short",
		},
		{
			sessionID: "",
			expected:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("SessionID_%s", tc.sessionID), func(t *testing.T) {
			result := GetSessionPrefix(tc.sessionID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMessageDirFunctions_EmptyMessageDir(t *testing.T) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 设置空的MessageDir
	MessageDir = ""

	sessionID := "test-session-123"
	requestSeqID := 1

	// 测试所有需要MessageDir的函数
	result := AppendToSessionLogFile(sessionID, "test.log", "test content")
	assert.Empty(t, result)

	result = WriteRequestMessage(sessionID, requestSeqID, "test message")
	assert.Empty(t, result)

	result = AppendToStreamSessionLog(sessionID, requestSeqID, "test chunk")
	assert.Empty(t, result)

	result = WriteChatResponseJson(sessionID, requestSeqID, map[string]string{"test": "data"})
	assert.Empty(t, result)

	result = WriteToolResultsJson(sessionID, requestSeqID, []string{"result1", "result2"})
	assert.Empty(t, result)
}

func TestMessageDirFunctions_EmptySessionID(t *testing.T) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "message-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	MessageDir = tmpDir

	requestSeqID := 1

	// 测试空sessionID
	result := AppendToSessionLogFile("", "test.log", "test content")
	assert.Empty(t, result)

	result = WriteRequestMessage("", requestSeqID, "test message")
	assert.Empty(t, result)

	result = AppendToStreamSessionLog("", requestSeqID, "test chunk")
	assert.Empty(t, result)
}

func TestMessageDirFunctions_InvalidRequestSeqID(t *testing.T) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "message-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	MessageDir = tmpDir
	sessionID := "test-session-123"

	// 测试无效的requestSeqID
	result := WriteRequestMessage(sessionID, 0, "test message")
	assert.Empty(t, result)

	result = WriteRequestMessage(sessionID, -1, "test message")
	assert.Empty(t, result)
}

func TestAppendToSessionLogFile_Success(t *testing.T) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "message-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	MessageDir = tmpDir
	sessionID := "test-session-12345678"
	filename := "test.log"
	content := "test content line 1\ntest content line 2"

	result := AppendToSessionLogFile(sessionID, filename, content)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, GetSessionPrefix(sessionID))
	assert.Contains(t, result, filename)

	// 验证文件内容
	fileContent, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Equal(t, content, string(fileContent))
}

func TestWriteRequestMessageJson_Success(t *testing.T) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "message-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	MessageDir = tmpDir
	sessionID := "test-session-12345678"
	requestSeqID := 1
	message := map[string]interface{}{
		"type":    "request",
		"content": "test message",
		"id":      123,
	}

	result := WriteRequestMessageJson(sessionID, requestSeqID, message)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "1_request.json")

	// 验证文件内容是有效的JSON
	fileContent, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Contains(t, string(fileContent), "test message")
	assert.Contains(t, string(fileContent), "request")
}

func TestWriteRequestMessageJson_MarshalError(t *testing.T) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "message-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	MessageDir = tmpDir
	sessionID := "test-session-12345678"
	requestSeqID := 1

	// 创建无法序列化的对象
	unmarshalableMessage := map[string]interface{}{
		"func": func() {}, // 函数无法序列化为JSON
	}

	result := WriteRequestMessageJson(sessionID, requestSeqID, unmarshalableMessage)
	assert.Empty(t, result) // 应该返回空字符串
}

// 基准测试
func BenchmarkGetCaller(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getCaller()
	}
}

func BenchmarkGetSessionPrefix(b *testing.B) {
	sessionID := "test-session-1234567890123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetSessionPrefix(sessionID)
	}
}

func BenchmarkInfo(b *testing.B) {
	// 设置一个丢弃输出的logger来避免I/O开销
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message", "key", "value", "number", i)
	}
}

func BenchmarkAppendToSessionLogFile(b *testing.B) {
	// 备份原始MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	MessageDir = tmpDir
	sessionID := "benchmark-session-12345678"
	filename := "benchmark.log"
	content := "benchmark content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AppendToSessionLogFile(sessionID, filename, content)
	}
}

// 测试持久化参数常量
func TestPersistConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "$_persist", persistKeyArg)
	assert.Equal(t, "$_persist_time", PersistTimeArg)
}