package logging

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogMessage_Struct(t *testing.T) {
	t.Parallel()

	now := time.Now()
	attrs := []Attr{
		{Key: "user_id", Value: "123"},
		{Key: "action", Value: "login"},
	}

	msg := LogMessage{
		ID:          "log-msg-123",
		Time:        now,
		Level:       "info",
		Persist:     true,
		PersistTime: 5 * time.Second,
		Message:     "User logged in successfully",
		Attributes:  attrs,
	}

	assert.Equal(t, "log-msg-123", msg.ID)
	assert.Equal(t, now, msg.Time)
	assert.Equal(t, "info", msg.Level)
	assert.True(t, msg.Persist)
	assert.Equal(t, 5*time.Second, msg.PersistTime)
	assert.Equal(t, "User logged in successfully", msg.Message)
	assert.Len(t, msg.Attributes, 2)
	assert.Equal(t, attrs, msg.Attributes)
}

func TestLogMessage_Defaults(t *testing.T) {
	t.Parallel()

	msg := LogMessage{}

	assert.Empty(t, msg.ID)
	assert.True(t, msg.Time.IsZero())
	assert.Empty(t, msg.Level)
	assert.False(t, msg.Persist)
	assert.Equal(t, time.Duration(0), msg.PersistTime)
	assert.Empty(t, msg.Message)
	assert.Nil(t, msg.Attributes)
}

func TestAttr_Struct(t *testing.T) {
	t.Parallel()

	attr := Attr{
		Key:   "test_key",
		Value: "test_value",
	}

	assert.Equal(t, "test_key", attr.Key)
	assert.Equal(t, "test_value", attr.Value)
}

func TestAttr_EmptyValues(t *testing.T) {
	t.Parallel()

	attr := Attr{
		Key:   "",
		Value: "",
	}

	assert.Empty(t, attr.Key)
	assert.Empty(t, attr.Value)
}

func TestLogMessage_JSONSerialization(t *testing.T) {
	t.Parallel()

	now := time.Now()
	attrs := []Attr{
		{Key: "user_id", Value: "123"},
		{Key: "ip_address", Value: "192.168.1.1"},
	}

	originalMsg := LogMessage{
		ID:          "json-test-123",
		Time:        now,
		Level:       "warn",
		Persist:     true,
		PersistTime: 10 * time.Second,
		Message:     "Test JSON serialization",
		Attributes:  attrs,
	}

	// 序列化
	jsonData, err := json.Marshal(originalMsg)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// 验证JSON包含预期字段
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "json-test-123")
	assert.Contains(t, jsonStr, "warn")
	assert.Contains(t, jsonStr, "Test JSON serialization")
	assert.Contains(t, jsonStr, "user_id")
	assert.Contains(t, jsonStr, "123")

	// 反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	// 验证反序列化结果
	assert.Equal(t, originalMsg.ID, deserializedMsg.ID)
	assert.Equal(t, originalMsg.Level, deserializedMsg.Level)
	assert.Equal(t, originalMsg.Persist, deserializedMsg.Persist)
	assert.Equal(t, originalMsg.PersistTime, deserializedMsg.PersistTime)
	assert.Equal(t, originalMsg.Message, deserializedMsg.Message)
	assert.Len(t, deserializedMsg.Attributes, 2)
	assert.Equal(t, originalMsg.Attributes, deserializedMsg.Attributes)

	// 时间字段需要特殊处理，因为JSON序列化可能会有精度差异
	assert.True(t, originalMsg.Time.Equal(deserializedMsg.Time))
}

func TestAttr_JSONSerialization(t *testing.T) {
	t.Parallel()

	originalAttr := Attr{
		Key:   "json_key",
		Value: "json_value",
	}

	// 序列化
	jsonData, err := json.Marshal(originalAttr)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// 验证JSON内容
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "json_key")
	assert.Contains(t, jsonStr, "json_value")

	// 反序列化
	var deserializedAttr Attr
	err = json.Unmarshal(jsonData, &deserializedAttr)
	require.NoError(t, err)

	assert.Equal(t, originalAttr.Key, deserializedAttr.Key)
	assert.Equal(t, originalAttr.Value, deserializedAttr.Value)
}

func TestLogMessage_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	msg := LogMessage{
		ID:      "special-chars-123",
		Level:   "error",
		Message: "Message with special chars: \"quotes\", 'apostrophes', \n newlines, \t tabs",
		Attributes: []Attr{
			{Key: "unicode", Value: "测试中文字符"},
			{Key: "emoji", Value: "🚀🔥💯"},
			{Key: "json_chars", Value: `{"nested": "json"}`},
		},
	}

	// 测试JSON序列化不会失败
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// 测试反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.Message, deserializedMsg.Message)
	assert.Equal(t, msg.Attributes, deserializedMsg.Attributes)
}

func TestLogMessage_EmptyAttributes(t *testing.T) {
	t.Parallel()

	msg := LogMessage{
		ID:         "empty-attrs-123",
		Level:      "debug",
		Message:    "Message without attributes",
		Attributes: []Attr{},
	}

	// 序列化
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)

	// 反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, deserializedMsg.ID)
	assert.Equal(t, msg.Message, deserializedMsg.Message)
	assert.NotNil(t, deserializedMsg.Attributes)
	assert.Len(t, deserializedMsg.Attributes, 0)
}

func TestLogMessage_NilAttributes(t *testing.T) {
	t.Parallel()

	msg := LogMessage{
		ID:         "nil-attrs-123",
		Level:      "debug",
		Message:    "Message with nil attributes",
		Attributes: nil,
	}

	// 序列化
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)

	// 反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, deserializedMsg.ID)
	assert.Equal(t, msg.Message, deserializedMsg.Message)
	// JSON反序列化nil slice会变成nil，这是正确的行为
	assert.Nil(t, deserializedMsg.Attributes)
}

func TestLogMessage_ZeroTime(t *testing.T) {
	t.Parallel()

	msg := LogMessage{
		ID:      "zero-time-123",
		Time:    time.Time{}, // 零值时间
		Level:   "info",
		Message: "Message with zero time",
	}

	// 序列化
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)

	// 反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, deserializedMsg.ID)
	assert.Equal(t, msg.Message, deserializedMsg.Message)
	assert.True(t, deserializedMsg.Time.IsZero())
}

func TestLogMessage_LongMessage(t *testing.T) {
	t.Parallel()

	// 创建很长的消息
	longMessage := ""
	for i := 0; i < 1000; i++ {
		longMessage += "This is a very long message. "
	}

	msg := LogMessage{
		ID:      "long-msg-123",
		Level:   "info",
		Message: longMessage,
	}

	// 测试序列化大消息
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// 测试反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.Message, deserializedMsg.Message)
	assert.Equal(t, len(longMessage), len(deserializedMsg.Message))
}

func TestLogMessage_ManyAttributes(t *testing.T) {
	t.Parallel()

	// 创建大量属性
	attrs := make([]Attr, 100)
	for i := 0; i < 100; i++ {
		attrs[i] = Attr{
			Key:   fmt.Sprintf("key_%d", i),
			Value: fmt.Sprintf("value_%d", i),
		}
	}

	msg := LogMessage{
		ID:         "many-attrs-123",
		Level:      "debug",
		Message:    "Message with many attributes",
		Attributes: attrs,
	}

	// 测试序列化
	jsonData, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// 测试反序列化
	var deserializedMsg LogMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, deserializedMsg.ID)
	assert.Equal(t, msg.Message, deserializedMsg.Message)
	assert.Len(t, deserializedMsg.Attributes, 100)
	assert.Equal(t, attrs, deserializedMsg.Attributes)
}

// 基准测试
func BenchmarkLogMessage_JSONMarshal(b *testing.B) {
	msg := LogMessage{
		ID:      "benchmark-marshal",
		Time:    time.Now(),
		Level:   "info",
		Message: "Benchmark message for JSON marshaling",
		Attributes: []Attr{
			{Key: "user_id", Value: "12345"},
			{Key: "action", Value: "test"},
			{Key: "timestamp", Value: "2023-11-20T10:30:00Z"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(msg)
	}
}

func BenchmarkLogMessage_JSONUnmarshal(b *testing.B) {
	msg := LogMessage{
		ID:      "benchmark-unmarshal",
		Time:    time.Now(),
		Level:   "info",
		Message: "Benchmark message for JSON unmarshaling",
		Attributes: []Attr{
			{Key: "user_id", Value: "12345"},
			{Key: "action", Value: "test"},
		},
	}

	jsonData, err := json.Marshal(msg)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var deserializedMsg LogMessage
		_ = json.Unmarshal(jsonData, &deserializedMsg)
	}
}

func BenchmarkAttr_JSONMarshal(b *testing.B) {
	attr := Attr{
		Key:   "benchmark_key",
		Value: "benchmark_value_with_some_longer_content",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(attr)
	}
}