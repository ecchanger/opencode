package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTextContent(t *testing.T) {
	t.Parallel()

	text := TextContent{
		Text: "Hello, world!",
	}

	assert.Equal(t, "Hello, world!", text.Text)
}

func TestReasoningContent(t *testing.T) {
	t.Parallel()

	reasoning := ReasoningContent{
		Thinking: "Let me think about this...",
	}

	assert.Equal(t, "Let me think about this...", reasoning.Thinking)
}

func TestImageURLContent(t *testing.T) {
	t.Parallel()

	imageURL := ImageURLContent{
		URL: "https://example.com/image.jpg",
	}

	assert.Equal(t, "https://example.com/image.jpg", imageURL.URL)
}

func TestBinaryContent(t *testing.T) {
	t.Parallel()

	data := []byte("binary data")
	binary := BinaryContent{
		MIMEType: "application/octet-stream",
		Data:     data,
	}

	assert.Equal(t, "application/octet-stream", binary.MIMEType)
	assert.Equal(t, data, binary.Data)
}

func TestToolCall(t *testing.T) {
	t.Parallel()

	toolCall := ToolCall{
		ID:   "tool-call-123",
		Name: "bash",
		Input: `{"command": "ls -la"}`,
	}

	assert.Equal(t, "tool-call-123", toolCall.ID)
	assert.Equal(t, "bash", toolCall.Name)
	assert.Equal(t, `{"command": "ls -la"}`, toolCall.Input)
}

func TestToolResult(t *testing.T) {
	t.Parallel()

	toolResult := ToolResult{
		ToolCallID: "tool-call-123",
		Content:    "file1.txt\nfile2.txt",
		IsError:    false,
	}

	assert.Equal(t, "tool-call-123", toolResult.ToolCallID)
	assert.Equal(t, "file1.txt\nfile2.txt", toolResult.Content)
	assert.False(t, toolResult.IsError)

	// 测试错误结果
	errorResult := ToolResult{
		ToolCallID: "tool-call-456",
		Content:    "command not found",
		IsError:    true,
	}

	assert.True(t, errorResult.IsError)
}

func TestFinish(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	finish := Finish{
		Reason: FinishReasonEndTurn,
		Time:   now,
	}

	assert.Equal(t, FinishReasonEndTurn, finish.Reason)
	assert.Equal(t, now, finish.Time)
}

func TestMessageRole(t *testing.T) {
	t.Parallel()

	// 测试消息角色常量
	assert.Equal(t, MessageRole("user"), User)
	assert.Equal(t, MessageRole("assistant"), Assistant)
}

func TestMessage_FinishPart(t *testing.T) {
	t.Parallel()

	// 测试没有完成部分的消息
	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			ToolCall{ID: "call-1", Name: "test", Input: "{}"},
		},
	}

	finishPart := msg.FinishPart()
	assert.Nil(t, finishPart)

	// 测试有完成部分的消息
	now := time.Now().Unix()
	finish := Finish{Reason: FinishReasonEndTurn, Time: now}
	msg.Parts = append(msg.Parts, finish)

	finishPart = msg.FinishPart()
	assert.NotNil(t, finishPart)
	assert.Equal(t, FinishReasonEndTurn, finishPart.Reason)
	assert.Equal(t, now, finishPart.Time)
}

func TestMessage_IsFinished(t *testing.T) {
	t.Parallel()

	// 测试未完成的消息
	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
		},
	}

	assert.False(t, msg.IsFinished())

	// 测试已完成的消息
	msg.Parts = append(msg.Parts, Finish{Reason: FinishReasonEndTurn, Time: time.Now().Unix()})
	assert.True(t, msg.IsFinished())
}

func TestMessage_Content(t *testing.T) {
	t.Parallel()

	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello, world!"},
			ToolCall{ID: "call-1", Name: "test", Input: "{}"},
			ToolResult{ToolCallID: "call-1", Content: "result", IsError: false},
		},
	}

	content := msg.Content()
	expected := TextContent{Text: "Hello, world!"}
	assert.Equal(t, expected, content)
}

func TestMessage_ToolResults(t *testing.T) {
	t.Parallel()

	toolResult1 := ToolResult{ToolCallID: "call-1", Content: "result1", IsError: false}
	toolResult2 := ToolResult{ToolCallID: "call-2", Content: "error", IsError: true}

	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			toolResult1,
			ToolCall{ID: "call-3", Name: "test", Input: "{}"},
			toolResult2,
		},
	}

	results := msg.ToolResults()
	assert.Len(t, results, 2)
	assert.Contains(t, results, toolResult1)
	assert.Contains(t, results, toolResult2)
}

func TestMessage_ToolCalls(t *testing.T) {
	t.Parallel()

	toolCall1 := ToolCall{ID: "call-1", Name: "bash", Input: `{"command": "ls"}`}
	toolCall2 := ToolCall{ID: "call-2", Name: "edit", Input: `{"file": "test.txt"}`}

	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			toolCall1,
			ToolResult{ToolCallID: "call-1", Content: "result", IsError: false},
			toolCall2,
		},
	}

	calls := msg.ToolCalls()
	assert.Len(t, calls, 2)
	assert.Contains(t, calls, toolCall1)
	assert.Contains(t, calls, toolCall2)
}

func TestMessage_ImageContent(t *testing.T) {
	t.Parallel()

	imageURL1 := ImageURLContent{URL: "https://example.com/image1.jpg"}
	imageURL2 := ImageURLContent{URL: "https://example.com/image2.png"}
	binary1 := BinaryContent{MIMEType: "image/jpeg", Data: []byte("fake image data")}

	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			imageURL1,
			binary1,
			imageURL2,
		},
	}

	imageURLs := msg.ImageURLContent()
	assert.Len(t, imageURLs, 2)
	assert.Contains(t, imageURLs, imageURL1)
	assert.Contains(t, imageURLs, imageURL2)

	binaries := msg.BinaryContent()
	assert.Len(t, binaries, 1)
	assert.Contains(t, binaries, binary1)
}

func TestMessage_AppendContent(t *testing.T) {
	t.Parallel()

	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
		},
	}

	// 添加文本内容
	msg.AppendContent(", world!")

	assert.Len(t, msg.Parts, 1) // AppendContent 会修改现有的TextContent
	content := msg.Content()
	assert.Equal(t, "Hello, world!", content.Text)
}

func TestMessage_AddToolCall(t *testing.T) {
	t.Parallel()

	msg := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
		},
	}

	// 添加工具调用
	toolCall := ToolCall{ID: "call-1", Name: "test", Input: "{}"}
	msg.AddToolCall(toolCall)

	assert.Len(t, msg.Parts, 2)
	calls := msg.ToolCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, toolCall, calls[0])
}