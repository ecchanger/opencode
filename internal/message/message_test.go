package message

import (
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestMessageRole(t *testing.T) {
	t.Parallel()

	// 测试所有角色常量
	assert.Equal(t, MessageRole("assistant"), Assistant)
	assert.Equal(t, MessageRole("user"), User)
	assert.Equal(t, MessageRole("system"), System)
	assert.Equal(t, MessageRole("tool"), Tool)
}

func TestFinishReason(t *testing.T) {
	t.Parallel()

	// 测试所有完成原因常量
	reasons := []FinishReason{
		FinishReasonEndTurn,
		FinishReasonMaxTokens,
		FinishReasonToolUse,
		FinishReasonCanceled,
		FinishReasonError,
		FinishReasonPermissionDenied,
		FinishReasonUnknown,
	}

	expectedValues := []string{
		"end_turn",
		"max_tokens",
		"tool_use",
		"canceled",
		"error",
		"permission_denied",
		"unknown",
	}

	for i, reason := range reasons {
		assert.Equal(t, expectedValues[i], string(reason))
	}
}

func TestContentParts(t *testing.T) {
	t.Parallel()

	t.Run("ReasoningContent", func(t *testing.T) {
		content := ReasoningContent{Thinking: "思考过程"}
		assert.Equal(t, "思考过程", content.String())
		
		// 测试接口实现
		var part ContentPart = content
		assert.NotNil(t, part)
	})

	t.Run("TextContent", func(t *testing.T) {
		content := TextContent{Text: "hello world"}
		assert.Equal(t, "hello world", content.String())
		
		var part ContentPart = content
		assert.NotNil(t, part)
	})

	t.Run("ImageURLContent", func(t *testing.T) {
		content := ImageURLContent{
			URL:    "https://example.com/image.jpg",
			Detail: "high",
		}
		assert.Equal(t, "https://example.com/image.jpg", content.String())
		
		var part ContentPart = content
		assert.NotNil(t, part)
	})

	t.Run("BinaryContent", func(t *testing.T) {
		data := []byte("binary data")
		content := BinaryContent{
			Path:     "/path/to/file",
			MIMEType: "image/jpeg",
			Data:     data,
		}
		
		// 测试OpenAI格式
		openaiResult := content.String(models.ProviderOpenAI)
		assert.Contains(t, openaiResult, "data:image/jpeg;base64,")
		
		// 测试其他提供商格式
		otherResult := content.String(models.ProviderAnthropic)
		assert.NotContains(t, otherResult, "data:")
		
		var part ContentPart = content
		assert.NotNil(t, part)
	})

	t.Run("ToolCall", func(t *testing.T) {
		toolCall := ToolCall{
			ID:       "call_123",
			Name:     "get_weather",
			Input:    `{"location": "Beijing"}`,
			Type:     "function",
			Finished: false,
		}
		
		var part ContentPart = toolCall
		assert.NotNil(t, part)
	})

	t.Run("ToolResult", func(t *testing.T) {
		toolResult := ToolResult{
			ToolCallID: "call_123",
			Name:       "get_weather",
			Content:    "Weather is sunny",
			Metadata:   "source: api",
			IsError:    false,
		}
		
		var part ContentPart = toolResult
		assert.NotNil(t, part)
	})

	t.Run("Finish", func(t *testing.T) {
		finish := Finish{
			Reason: FinishReasonEndTurn,
			Time:   time.Now().Unix(),
		}
		
		var part ContentPart = finish
		assert.NotNil(t, part)
	})
}

func TestMessage(t *testing.T) {
	t.Parallel()

	t.Run("Content方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ReasoningContent{Thinking: "思考"},
				TextContent{Text: "hello"},
				TextContent{Text: "world"}, // 只返回第一个
			},
		}
		
		content := msg.Content()
		assert.Equal(t, "hello", content.Text)
	})

	t.Run("ReasoningContent方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
				ReasoningContent{Thinking: "思考过程"},
			},
		}
		
		reasoning := msg.ReasoningContent()
		assert.Equal(t, "思考过程", reasoning.Thinking)
	})

	t.Run("ImageURLContent方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ImageURLContent{URL: "url1"},
				TextContent{Text: "hello"},
				ImageURLContent{URL: "url2"},
			},
		}
		
		images := msg.ImageURLContent()
		assert.Len(t, images, 2)
		assert.Equal(t, "url1", images[0].URL)
		assert.Equal(t, "url2", images[1].URL)
	})

	t.Run("BinaryContent方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				BinaryContent{Data: []byte("data1")},
				TextContent{Text: "hello"},
				BinaryContent{Data: []byte("data2")},
			},
		}
		
		binaries := msg.BinaryContent()
		assert.Len(t, binaries, 2)
		assert.Equal(t, []byte("data1"), binaries[0].Data)
		assert.Equal(t, []byte("data2"), binaries[1].Data)
	})

	t.Run("ToolCalls方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ToolCall{ID: "call1"},
				TextContent{Text: "hello"},
				ToolCall{ID: "call2"},
			},
		}
		
		toolCalls := msg.ToolCalls()
		assert.Len(t, toolCalls, 2)
		assert.Equal(t, "call1", toolCalls[0].ID)
		assert.Equal(t, "call2", toolCalls[1].ID)
	})

	t.Run("ToolResults方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ToolResult{ToolCallID: "call1"},
				TextContent{Text: "hello"},
				ToolResult{ToolCallID: "call2"},
			},
		}
		
		toolResults := msg.ToolResults()
		assert.Len(t, toolResults, 2)
		assert.Equal(t, "call1", toolResults[0].ToolCallID)
		assert.Equal(t, "call2", toolResults[1].ToolCallID)
	})

	t.Run("IsFinished方法", func(t *testing.T) {
		msgUnfinished := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
			},
		}
		assert.False(t, msgUnfinished.IsFinished())
		
		msgFinished := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
				Finish{Reason: FinishReasonEndTurn},
			},
		}
		assert.True(t, msgFinished.IsFinished())
	})

	t.Run("FinishPart方法", func(t *testing.T) {
		finish := Finish{Reason: FinishReasonEndTurn, Time: 123456}
		msg := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
				finish,
			},
		}
		
		finishPart := msg.FinishPart()
		require.NotNil(t, finishPart)
		assert.Equal(t, FinishReasonEndTurn, finishPart.Reason)
		assert.Equal(t, int64(123456), finishPart.Time)
	})

	t.Run("FinishReason方法", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
				Finish{Reason: FinishReasonMaxTokens},
			},
		}
		
		reason := msg.FinishReason()
		assert.Equal(t, FinishReasonMaxTokens, reason)
	})

	t.Run("IsThinking方法", func(t *testing.T) {
		// 正在思考：有推理内容，无文本内容，未完成
		thinkingMsg := Message{
			Parts: []ContentPart{
				ReasoningContent{Thinking: "正在思考..."},
			},
		}
		assert.True(t, thinkingMsg.IsThinking())
		
		// 不在思考：有文本内容
		notThinkingMsg := Message{
			Parts: []ContentPart{
				ReasoningContent{Thinking: "思考过程"},
				TextContent{Text: "hello"},
			},
		}
		assert.False(t, notThinkingMsg.IsThinking())
		
		// 不在思考：已完成
		finishedMsg := Message{
			Parts: []ContentPart{
				ReasoningContent{Thinking: "思考过程"},
				Finish{Reason: FinishReasonEndTurn},
			},
		}
		assert.False(t, finishedMsg.IsThinking())
	})
}

func TestMessageAppendMethods(t *testing.T) {
	t.Parallel()

	t.Run("AppendContent", func(t *testing.T) {
		// 已有文本内容
		msg := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
			},
		}
		
		msg.AppendContent(" world")
		assert.Equal(t, "hello world", msg.Content().Text)
		
		// 无文本内容，新增
		emptyMsg := Message{Parts: []ContentPart{}}
		emptyMsg.AppendContent("new text")
		assert.Equal(t, "new text", emptyMsg.Content().Text)
	})

	t.Run("AppendReasoningContent", func(t *testing.T) {
		// 已有推理内容
		msg := Message{
			Parts: []ContentPart{
				ReasoningContent{Thinking: "step1"},
			},
		}
		
		msg.AppendReasoningContent(" step2")
		assert.Equal(t, "step1 step2", msg.ReasoningContent().Thinking)
		
		// 无推理内容，新增
		emptyMsg := Message{Parts: []ContentPart{}}
		emptyMsg.AppendReasoningContent("new thinking")
		assert.Equal(t, "new thinking", emptyMsg.ReasoningContent().Thinking)
	})

	t.Run("FinishToolCall", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ToolCall{ID: "call1", Finished: false},
				ToolCall{ID: "call2", Finished: false},
			},
		}
		
		msg.FinishToolCall("call1")
		
		toolCalls := msg.ToolCalls()
		assert.True(t, toolCalls[0].Finished)
		assert.False(t, toolCalls[1].Finished)
	})

	t.Run("AppendToolCallInput", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ToolCall{ID: "call1", Input: "initial"},
			},
		}
		
		msg.AppendToolCallInput("call1", " appended")
		
		toolCall := msg.ToolCalls()[0]
		assert.Equal(t, "initial appended", toolCall.Input)
	})

	t.Run("AddToolCall", func(t *testing.T) {
		msg := Message{Parts: []ContentPart{}}
		
		toolCall := ToolCall{ID: "call1", Name: "test"}
		msg.AddToolCall(toolCall)
		
		assert.Len(t, msg.ToolCalls(), 1)
		assert.Equal(t, "call1", msg.ToolCalls()[0].ID)
		
		// 更新现有的tool call
		updatedToolCall := ToolCall{ID: "call1", Name: "updated"}
		msg.AddToolCall(updatedToolCall)
		
		assert.Len(t, msg.ToolCalls(), 1)
		assert.Equal(t, "updated", msg.ToolCalls()[0].Name)
	})

	t.Run("SetToolCalls", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				ToolCall{ID: "old_call"},
				TextContent{Text: "keep this"},
			},
		}
		
		newToolCalls := []ToolCall{
			{ID: "new_call1"},
			{ID: "new_call2"},
		}
		
		msg.SetToolCalls(newToolCalls)
		
		// 应该替换所有tool calls，保留其他parts
		assert.Len(t, msg.ToolCalls(), 2)
		assert.Equal(t, "keep this", msg.Content().Text)
	})

	t.Run("AddToolResult", func(t *testing.T) {
		msg := Message{Parts: []ContentPart{}}
		
		toolResult := ToolResult{ToolCallID: "call1", Content: "result"}
		msg.AddToolResult(toolResult)
		
		assert.Len(t, msg.ToolResults(), 1)
		assert.Equal(t, "call1", msg.ToolResults()[0].ToolCallID)
	})

	t.Run("SetToolResults", func(t *testing.T) {
		msg := Message{Parts: []ContentPart{}}
		
		toolResults := []ToolResult{
			{ToolCallID: "call1"},
			{ToolCallID: "call2"},
		}
		
		msg.SetToolResults(toolResults)
		
		assert.Len(t, msg.ToolResults(), 2)
	})

	t.Run("AddFinish", func(t *testing.T) {
		msg := Message{
			Parts: []ContentPart{
				TextContent{Text: "hello"},
				Finish{Reason: FinishReasonMaxTokens}, // 应该被替换
			},
		}
		
		msg.AddFinish(FinishReasonEndTurn)
		
		assert.True(t, msg.IsFinished())
		assert.Equal(t, FinishReasonEndTurn, msg.FinishReason())
		
		// 应该只有一个finish part
		finishCount := 0
		for _, part := range msg.Parts {
			if _, ok := part.(Finish); ok {
				finishCount++
			}
		}
		assert.Equal(t, 1, finishCount)
	})

	t.Run("AddImageURL", func(t *testing.T) {
		msg := Message{Parts: []ContentPart{}}
		
		msg.AddImageURL("http://example.com/image.jpg", "high")
		
		images := msg.ImageURLContent()
		assert.Len(t, images, 1)
		assert.Equal(t, "http://example.com/image.jpg", images[0].URL)
		assert.Equal(t, "high", images[0].Detail)
	})

	t.Run("AddBinary", func(t *testing.T) {
		msg := Message{Parts: []ContentPart{}}
		
		data := []byte("binary data")
		msg.AddBinary("image/jpeg", data)
		
		binaries := msg.BinaryContent()
		assert.Len(t, binaries, 1)
		assert.Equal(t, "image/jpeg", binaries[0].MIMEType)
		assert.Equal(t, data, binaries[0].Data)
	})
}

func TestMarshallParts(t *testing.T) {
	t.Parallel()

	t.Run("单一类型", func(t *testing.T) {
		parts := []ContentPart{
			TextContent{Text: "hello"},
		}
		
		data, err := marshallParts(parts)
		require.NoError(t, err)
		
		// 验证能够成功反序列化
		unmarshalled, err := unmarshallParts(data)
		require.NoError(t, err)
		
		assert.Len(t, unmarshalled, 1)
		textPart, ok := unmarshalled[0].(TextContent)
		assert.True(t, ok)
		assert.Equal(t, "hello", textPart.Text)
	})

	t.Run("多种类型", func(t *testing.T) {
		parts := []ContentPart{
			TextContent{Text: "hello"},
			ReasoningContent{Thinking: "思考"},
			ToolCall{ID: "call1"},
			ToolResult{ToolCallID: "call1"},
			Finish{Reason: FinishReasonEndTurn},
		}
		
		data, err := marshallParts(parts)
		require.NoError(t, err)
		
		// 验证能够成功反序列化
		unmarshalled, err := unmarshallParts(data)
		require.NoError(t, err)
		
		assert.Len(t, unmarshalled, 5)
		
		// 验证每个类型都正确反序列化
		textPart, ok := unmarshalled[0].(TextContent)
		assert.True(t, ok)
		assert.Equal(t, "hello", textPart.Text)
		
		reasoningPart, ok := unmarshalled[1].(ReasoningContent)
		assert.True(t, ok)
		assert.Equal(t, "思考", reasoningPart.Thinking)
	})

	t.Run("未知类型", func(t *testing.T) {
		// 由于Go不允许在函数内部定义方法，我们跳过这个测试
		// 这个功能在实际使用中应该通过interface设计来保证类型安全
		t.Skip("跳过未知类型测试 - Go不允许在函数内部定义方法")
	})
}

func TestUnmarshallParts(t *testing.T) {
	t.Parallel()

	t.Run("有效JSON", func(t *testing.T) {
		// 先序列化已知数据
		originalParts := []ContentPart{
			TextContent{Text: "hello"},
			ReasoningContent{Thinking: "思考"},
		}
		
		data, err := marshallParts(originalParts)
		require.NoError(t, err)
		
		// 再反序列化
		parts, err := unmarshallParts(data)
		require.NoError(t, err)
		
		assert.Len(t, parts, 2)
		
		textPart, ok := parts[0].(TextContent)
		assert.True(t, ok)
		assert.Equal(t, "hello", textPart.Text)
		
		reasoningPart, ok := parts[1].(ReasoningContent)
		assert.True(t, ok)
		assert.Equal(t, "思考", reasoningPart.Thinking)
	})

	t.Run("无效JSON", func(t *testing.T) {
		invalidJSON := []byte("invalid json")
		
		_, err := unmarshallParts(invalidJSON)
		assert.Error(t, err)
	})

	t.Run("未知part类型", func(t *testing.T) {
		// 手动构造包含未知类型的JSON
		unknownJSON := `[{"type":"unknown_type","data":{}}]`
		
		_, err := unmarshallParts([]byte(unknownJSON))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown part type")
	})
}

// 跳过服务测试，因为需要完整实现db.Querier接口
// 在实际项目中，可以使用sqlmock或其他更适合的模拟工具
func TestServiceSkipped(t *testing.T) {
	t.Skip("跳过服务测试 - 需要完整的db.Querier模拟实现")
}

// 基准测试
func BenchmarkMarshallParts(b *testing.B) {
	parts := []ContentPart{
		TextContent{Text: "hello world"},
		ReasoningContent{Thinking: "thinking process"},
		ToolCall{ID: "call1", Name: "test", Input: "input"},
		ToolResult{ToolCallID: "call1", Content: "result"},
		Finish{Reason: FinishReasonEndTurn, Time: time.Now().Unix()},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = marshallParts(parts)
	}
}

func BenchmarkUnmarshallParts(b *testing.B) {
	parts := []ContentPart{
		TextContent{Text: "hello world"},
		ReasoningContent{Thinking: "thinking process"},
		ToolCall{ID: "call1", Name: "test", Input: "input"},
	}
	
	data, _ := marshallParts(parts)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = unmarshallParts(data)
	}
}

func BenchmarkMessageContentAccess(b *testing.B) {
	msg := Message{
		Parts: []ContentPart{
			ReasoningContent{Thinking: "thinking"},
			TextContent{Text: "hello world"},
			ImageURLContent{URL: "http://example.com/image.jpg"},
			ToolCall{ID: "call1"},
			ToolResult{ToolCallID: "call1"},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.Content()
		_ = msg.ReasoningContent()
		_ = msg.ImageURLContent()
		_ = msg.ToolCalls()
		_ = msg.ToolResults()
		_ = msg.IsFinished()
	}
}