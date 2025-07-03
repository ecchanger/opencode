package logging

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestLogData_Add(t *testing.T) {
	t.Parallel()

	logData := &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	msg := LogMessage{
		ID:      "test-id",
		Time:    time.Now(),
		Level:   "info",
		Message: "test message",
	}

	logData.Add(msg)

	messages := logData.List()
	assert.Len(t, messages, 1)
	assert.Equal(t, msg.ID, messages[0].ID)
	assert.Equal(t, msg.Message, messages[0].Message)
}

func TestLogData_List(t *testing.T) {
	t.Parallel()

	logData := &LogData{
		messages: []LogMessage{
			{ID: "1", Message: "message 1"},
			{ID: "2", Message: "message 2"},
		},
		Broker: pubsub.NewBroker[LogMessage](),
	}

	messages := logData.List()
	assert.Len(t, messages, 2)
	assert.Equal(t, "1", messages[0].ID)
	assert.Equal(t, "2", messages[1].ID)
}

func TestLogData_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试")
	}

	logData := &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	// 并发添加消息
	const numGoroutines = 10
	const messagesPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := LogMessage{
					ID:      fmt.Sprintf("goroutine-%d-msg-%d", goroutineID, j),
					Message: fmt.Sprintf("Message from goroutine %d, msg %d", goroutineID, j),
					Level:   "info",
					Time:    time.Now(),
				}
				logData.Add(msg)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证消息数量
	messages := logData.List()
	assert.Len(t, messages, numGoroutines*messagesPerGoroutine)
}

func TestNewWriter(t *testing.T) {
	t.Parallel()

	writer := NewWriter()
	assert.NotNil(t, writer)
}

func TestWriter_Write_BasicLogEntry(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 模拟slog的logfmt输出
	logEntry := `time=2023-11-20T10:30:00Z level=INFO msg="test message" key=value`

	n, err := writer.Write([]byte(logEntry))

	assert.NoError(t, err)
	assert.Equal(t, len(logEntry), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, "test message", msg.Message)
	assert.Equal(t, "info", msg.Level)
	assert.Contains(t, msg.Attributes, Attr{Key: "key", Value: "value"})
	
	// 验证时间解析
	expectedTime, _ := time.Parse(time.RFC3339, "2023-11-20T10:30:00Z")
	assert.Equal(t, expectedTime.Unix(), msg.Time.Unix())
}

func TestWriter_Write_WithPersistFlag(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 包含persist标志的日志条目
	logEntry := fmt.Sprintf(`time=2023-11-20T10:30:00Z level=WARN msg="persist message" %s=true`, persistKeyArg)

	n, err := writer.Write([]byte(logEntry))

	assert.NoError(t, err)
	assert.Equal(t, len(logEntry), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, "persist message", msg.Message)
	assert.Equal(t, "warn", msg.Level)
	assert.True(t, msg.Persist)
}

func TestWriter_Write_WithPersistTime(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 包含persist time的日志条目
	logEntry := fmt.Sprintf(`time=2023-11-20T10:30:00Z level=ERROR msg="timed persist message" %s=5s`, PersistTimeArg)

	n, err := writer.Write([]byte(logEntry))

	assert.NoError(t, err)
	assert.Equal(t, len(logEntry), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, "timed persist message", msg.Message)
	assert.Equal(t, "error", msg.Level)
	assert.Equal(t, 5*time.Second, msg.PersistTime)
}

func TestWriter_Write_MultipleAttributes(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 包含多个属性的日志条目
	logEntry := `time=2023-11-20T10:30:00Z level=DEBUG msg="debug message" user_id=123 action=login ip=192.168.1.1`

	n, err := writer.Write([]byte(logEntry))

	assert.NoError(t, err)
	assert.Equal(t, len(logEntry), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, "debug message", msg.Message)
	assert.Equal(t, "debug", msg.Level)
	assert.Len(t, msg.Attributes, 3)

	// 验证属性
	expectedAttrs := map[string]string{
		"user_id": "123",
		"action":  "login",
		"ip":      "192.168.1.1",
	}

	for _, attr := range msg.Attributes {
		expectedValue, exists := expectedAttrs[attr.Key]
		assert.True(t, exists, "意外的属性: %s", attr.Key)
		assert.Equal(t, expectedValue, attr.Value)
	}
}

func TestWriter_Write_MultipleRecords(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 包含多条记录的日志数据
	logEntry := "time=2023-11-20T10:30:00Z level=INFO msg=\"first message\" key1=value1\n" +
		"time=2023-11-20T10:31:00Z level=WARN msg=\"second message\" key2=value2\n"

	n, err := writer.Write([]byte(logEntry))

	assert.NoError(t, err)
	assert.Equal(t, len(logEntry), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 2)

	// 验证第一条消息
	assert.Equal(t, "first message", messages[0].Message)
	assert.Equal(t, "info", messages[0].Level)
	assert.Contains(t, messages[0].Attributes, Attr{Key: "key1", Value: "value1"})

	// 验证第二条消息
	assert.Equal(t, "second message", messages[1].Message)
	assert.Equal(t, "warn", messages[1].Level)
	assert.Contains(t, messages[1].Attributes, Attr{Key: "key2", Value: "value2"})
}

func TestWriter_Write_InvalidTimeFormat(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 包含无效时间格式的日志条目
	logEntry := `time=invalid-time level=ERROR msg="error message"`

	_, err := writer.Write([]byte(logEntry))

	// 应该返回错误，因为时间解析失败
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing time")
}

func TestWriter_Write_InvalidPersistTime(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 包含无效persist time的日志条目
	logEntry := fmt.Sprintf(`time=2023-11-20T10:30:00Z level=INFO msg="message with invalid persist time" %s=invalid-duration`, PersistTimeArg)

	n, err := writer.Write([]byte(logEntry))

	// 不应该返回错误，无效的persist time应该被忽略
	assert.NoError(t, err)
	assert.Equal(t, len(logEntry), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, time.Duration(0), msg.PersistTime)
}

func TestWriter_Write_EmptyInput(t *testing.T) {
	writer := NewWriter()

	_, err := writer.Write([]byte(""))

	assert.NoError(t, err)
}

func TestWriter_Write_MalformedLogfmt(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 格式错误的logfmt数据
	logEntry := `time=2023-11-20T10:30:00Z level=INFO msg=unclosed quote"`

	_, err := writer.Write([]byte(logEntry))

	// logfmt解析器可能会处理格式错误的数据，具体行为取决于实现
	// 这里主要测试不会panic
	_ = err // 允许任何结果，主要测试不会panic
}

func TestSubscribe(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	eventChan := Subscribe(ctx)

	// 添加消息
	msg := LogMessage{
		ID:      "test-event",
		Message: "test event message",
	}

	defaultLogData.Add(msg)

	// 等待事件
	select {
	case event := <-eventChan:
		assert.Equal(t, pubsub.CreatedEvent, event.Type)
		assert.Equal(t, msg.ID, event.Payload.ID)
		assert.Equal(t, msg.Message, event.Payload.Message)
	case <-ctx.Done():
		t.Fatal("超时等待事件")
	}
}

func TestList(t *testing.T) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	messages := []LogMessage{
		{ID: "1", Message: "message 1"},
		{ID: "2", Message: "message 2"},
	}

	defaultLogData = &LogData{
		messages: messages,
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	result := List()
	assert.Len(t, result, 2)
	assert.Equal(t, messages, result)
}

// 基准测试
func BenchmarkWriter_Write(b *testing.B) {
	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()
	logEntry := []byte(`time=2023-11-20T10:30:00Z level=INFO msg="benchmark message" key=value`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = writer.Write(logEntry)
	}
}

func BenchmarkLogData_Add(b *testing.B) {
	logData := &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	msg := LogMessage{
		ID:      "benchmark-id",
		Time:    time.Now(),
		Level:   "info",
		Message: "benchmark message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logData.Add(msg)
	}
}

func BenchmarkLogData_List(b *testing.B) {
	// 准备测试数据
	messages := make([]LogMessage, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = LogMessage{
			ID:      fmt.Sprintf("msg-%d", i),
			Message: fmt.Sprintf("Message %d", i),
		}
	}

	logData := &LogData{
		messages: messages,
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logData.List()
	}
}

// 边界条件测试
func TestWriter_Write_LargeBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大批量测试")
	}

	// 备份原始defaultLogData
	originalLogData := defaultLogData
	defer func() { defaultLogData = originalLogData }()

	// 创建新的LogData用于测试
	defaultLogData = &LogData{
		messages: make([]LogMessage, 0),
		Broker:   pubsub.NewBroker[LogMessage](),
	}

	writer := NewWriter()

	// 创建大量日志条目
	var logBuffer bytes.Buffer
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&logBuffer, "time=2023-11-20T10:30:00Z level=INFO msg=\"message %d\" index=%d\n", i, i)
	}

	n, err := writer.Write(logBuffer.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, logBuffer.Len(), n)

	messages := defaultLogData.List()
	assert.Len(t, messages, 1000)
}