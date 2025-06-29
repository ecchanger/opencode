package message

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/opencode-ai/opencode/internal/db"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockQuerier implements db.Querier interface for testing
type MockQuerier struct {
	// Store test data
	messages map[string]db.Message
	sessions map[string]db.Session
	files    map[string]db.File
}

func NewMockQuerier() *MockQuerier {
	return &MockQuerier{
		messages: make(map[string]db.Message),
		sessions: make(map[string]db.Session),
		files:    make(map[string]db.File),
	}
}

// Message methods
func (m *MockQuerier) CreateMessage(ctx context.Context, params db.CreateMessageParams) (db.Message, error) {
	msg := db.Message{
		ID:        params.ID,
		SessionID: params.SessionID,
		Role:      params.Role,
		Parts:     params.Parts,
		Model:     params.Model,
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}
	m.messages[params.ID] = msg
	return msg, nil
}

func (m *MockQuerier) UpdateMessage(ctx context.Context, params db.UpdateMessageParams) error {
	if msg, exists := m.messages[params.ID]; exists {
		msg.Parts = params.Parts
		msg.FinishedAt = params.FinishedAt
		m.messages[params.ID] = msg
	}
	return nil
}

func (m *MockQuerier) GetMessage(ctx context.Context, id string) (db.Message, error) {
	if msg, exists := m.messages[id]; exists {
		return msg, nil
	}
	return db.Message{}, sql.ErrNoRows
}

func (m *MockQuerier) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.Message, error) {
	var result []db.Message
	for _, msg := range m.messages {
		if msg.SessionID == sessionID {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m *MockQuerier) DeleteMessage(ctx context.Context, id string) error {
	delete(m.messages, id)
	return nil
}

func (m *MockQuerier) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	for id, msg := range m.messages {
		if msg.SessionID == sessionID {
			delete(m.messages, id)
		}
	}
	return nil
}

// Session methods (stub implementations)
func (m *MockQuerier) CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
	session := db.Session{
		ID:        params.ID,
		Title:     params.Title,
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}
	m.sessions[params.ID] = session
	return session, nil
}

func (m *MockQuerier) GetSessionByID(ctx context.Context, id string) (db.Session, error) {
	if session, exists := m.sessions[id]; exists {
		return session, nil
	}
	return db.Session{}, sql.ErrNoRows
}

func (m *MockQuerier) ListSessions(ctx context.Context) ([]db.Session, error) {
	var result []db.Session
	for _, session := range m.sessions {
		result = append(result, session)
	}
	return result, nil
}

func (m *MockQuerier) UpdateSession(ctx context.Context, params db.UpdateSessionParams) (db.Session, error) {
	if session, exists := m.sessions[params.ID]; exists {
		session.Title = params.Title
		session.SummaryMessageID = params.SummaryMessageID
		m.sessions[params.ID] = session
		return session, nil
	}
	return db.Session{}, sql.ErrNoRows
}

func (m *MockQuerier) DeleteSession(ctx context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

// File methods (stub implementations)
func (m *MockQuerier) CreateFile(ctx context.Context, params db.CreateFileParams) (db.File, error) {
	file := db.File{
		ID:        params.ID,
		SessionID: params.SessionID,
		Path:      params.Path,
		Content:   params.Content,
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}
	m.files[params.ID] = file
	return file, nil
}

func (m *MockQuerier) GetFile(ctx context.Context, id string) (db.File, error) {
	if file, exists := m.files[id]; exists {
		return file, nil
	}
	return db.File{}, sql.ErrNoRows
}

func (m *MockQuerier) GetFileByPathAndSession(ctx context.Context, params db.GetFileByPathAndSessionParams) (db.File, error) {
	for _, file := range m.files {
		if file.Path == params.Path && file.SessionID == params.SessionID {
			return file, nil
		}
	}
	return db.File{}, sql.ErrNoRows
}

func (m *MockQuerier) UpdateFile(ctx context.Context, params db.UpdateFileParams) (db.File, error) {
	if file, exists := m.files[params.ID]; exists {
		file.Content = params.Content
		m.files[params.ID] = file
		return file, nil
	}
	return db.File{}, sql.ErrNoRows
}

func (m *MockQuerier) DeleteFile(ctx context.Context, id string) error {
	delete(m.files, id)
	return nil
}

func (m *MockQuerier) DeleteSessionFiles(ctx context.Context, sessionID string) error {
	for id, file := range m.files {
		if file.SessionID == sessionID {
			delete(m.files, id)
		}
	}
	return nil
}

func (m *MockQuerier) ListFilesBySession(ctx context.Context, sessionID string) ([]db.File, error) {
	var result []db.File
	for _, file := range m.files {
		if file.SessionID == sessionID {
			result = append(result, file)
		}
	}
	return result, nil
}

func (m *MockQuerier) ListFilesByPath(ctx context.Context, path string) ([]db.File, error) {
	var result []db.File
	for _, file := range m.files {
		if file.Path == path {
			result = append(result, file)
		}
	}
	return result, nil
}

func (m *MockQuerier) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]db.File, error) {
	return m.ListFilesBySession(ctx, sessionID)
}

func (m *MockQuerier) ListNewFiles(ctx context.Context) ([]db.File, error) {
	var result []db.File
	for _, file := range m.files {
		result = append(result, file)
	}
	return result, nil
}

func TestNewService(t *testing.T) {
	mockQuerier := NewMockQuerier()
	service := NewService(mockQuerier)

	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
}

func TestServiceCreate(t *testing.T) {
	mockQuerier := NewMockQuerier()
	service := NewService(mockQuerier)

	ctx := context.Background()
	sessionID := "test-session-id"

	params := CreateMessageParams{
		Role: User,
		Parts: []ContentPart{
			TextContent{Text: "Hello, world!"},
		},
		Model: models.Claude4Sonnet,
	}

	result, err := service.Create(ctx, sessionID, params)

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, sessionID, result.SessionID)
	assert.Equal(t, User, result.Role)
	assert.Equal(t, models.Claude4Sonnet, result.Model)
	assert.Len(t, result.Parts, 2) // Text + Finish
}

func TestServiceUpdate(t *testing.T) {
	mockQuerier := NewMockQuerier()
	service := NewService(mockQuerier)

	ctx := context.Background()

	// First create the message
	createParams := CreateMessageParams{
		Role: User,
		Parts: []ContentPart{
			TextContent{Text: "Original message"},
		},
		Model: models.Claude4Sonnet,
	}

	createdMsg, err := service.Create(ctx, "test-session-id", createParams)
	require.NoError(t, err)

	// Now update it
	createdMsg.Parts = []ContentPart{
		TextContent{Text: "Updated message"},
		Finish{Reason: "stop", Time: 1234567890},
	}

	err = service.Update(ctx, createdMsg)
	require.NoError(t, err)
}

func TestServiceGet(t *testing.T) {
	mockQuerier := NewMockQuerier()
	service := NewService(mockQuerier)

	ctx := context.Background()

	// First create a message
	params := CreateMessageParams{
		Role: User,
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
		},
		Model: models.Claude4Sonnet,
	}

	createdMsg, err := service.Create(ctx, "test-session-id", params)
	require.NoError(t, err)

	// Now get it
	result, err := service.Get(ctx, createdMsg.ID)

	require.NoError(t, err)
	assert.Equal(t, createdMsg.ID, result.ID)
	assert.Equal(t, "test-session-id", result.SessionID)
	assert.Equal(t, User, result.Role)
	assert.Equal(t, models.Claude4Sonnet, result.Model)
}

func TestServiceList(t *testing.T) {
	mockQuerier := &MockQuerier{}
	service := NewService(mockQuerier)

	ctx := context.Background()
	sessionID := "test-session-id"

	dbMessages := []db.Message{
		{
			ID:        "msg-1",
			SessionID: sessionID,
			Role:      string(User),
			Parts:     `[{"type":"text","data":{"text":"First message"}}]`,
			Model:     sql.NullString{String: string(models.Claude4Sonnet), Valid: true},
			CreatedAt: 1234567890,
			UpdatedAt: 1234567890,
		},
		{
			ID:        "msg-2",
			SessionID: sessionID,
			Role:      string(Assistant),
			Parts:     `[{"type":"text","data":{"text":"Second message"}}]`,
			Model:     sql.NullString{String: string(models.Claude4Sonnet), Valid: true},
			CreatedAt: 1234567891,
			UpdatedAt: 1234567891,
		},
	}

	mockQuerier.On("ListMessagesBySession", ctx, sessionID).Return(dbMessages, nil)

	results, err := service.List(ctx, sessionID)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "msg-1", results[0].ID)
	assert.Equal(t, "msg-2", results[1].ID)

	mockQuerier.AssertExpectations(t)
}

func TestServiceDelete(t *testing.T) {
	mockQuerier := &MockQuerier{}
	service := NewService(mockQuerier)

	ctx := context.Background()
	messageID := "test-message-id"

	dbMessage := db.Message{
		ID:        messageID,
		SessionID: "test-session-id",
		Role:      string(User),
		Parts:     `[{"type":"text","data":{"text":"Hello"}}]`,
		Model:     sql.NullString{String: string(models.Claude4Sonnet), Valid: true},
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}

	mockQuerier.On("GetMessage", ctx, messageID).Return(dbMessage, nil)
	mockQuerier.On("DeleteMessage", ctx, messageID).Return(nil)

	err := service.Delete(ctx, messageID)

	require.NoError(t, err)
	mockQuerier.AssertExpectations(t)
}

func TestServiceDeleteSessionMessages(t *testing.T) {
	mockQuerier := &MockQuerier{}
	service := NewService(mockQuerier)

	ctx := context.Background()
	sessionID := "test-session-id"

	dbMessages := []db.Message{
		{
			ID:        "msg-1",
			SessionID: sessionID,
			Role:      string(User),
			Parts:     `[{"type":"text","data":{"text":"First message"}}]`,
			Model:     sql.NullString{String: string(models.Claude4Sonnet), Valid: true},
		},
		{
			ID:        "msg-2",
			SessionID: sessionID,
			Role:      string(Assistant),
			Parts:     `[{"type":"text","data":{"text":"Second message"}}]`,
			Model:     sql.NullString{String: string(models.Claude4Sonnet), Valid: true},
		},
	}

	mockQuerier.On("ListMessagesBySession", ctx, sessionID).Return(dbMessages, nil)
	mockQuerier.On("GetMessage", ctx, "msg-1").Return(dbMessages[0], nil)
	mockQuerier.On("DeleteMessage", ctx, "msg-1").Return(nil)
	mockQuerier.On("GetMessage", ctx, "msg-2").Return(dbMessages[1], nil)
	mockQuerier.On("DeleteMessage", ctx, "msg-2").Return(nil)

	err := service.DeleteSessionMessages(ctx, sessionID)

	require.NoError(t, err)
	mockQuerier.AssertExpectations(t)
}

func TestMarshallParts(t *testing.T) {
	t.Run("marshalls various content parts", func(t *testing.T) {
		parts := []ContentPart{
			TextContent{Text: "Hello, world!"},
			ImageURLContent{URL: "https://example.com/image.jpg"},
			ToolCall{ID: "call-1", Name: "test-tool", Input: `{"param": "value"}`},
			ToolResult{ToolCallID: "call-1", Content: "Result content"},
			Finish{Reason: "stop", Time: 1234567890},
		}

		data, err := marshallParts(parts)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Should be valid JSON
		var result []map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Len(t, result, 5)
		assert.Equal(t, "text", result[0]["type"])
		assert.Equal(t, "image_url", result[1]["type"])
		assert.Equal(t, "tool_call", result[2]["type"])
		assert.Equal(t, "tool_result", result[3]["type"])
		assert.Equal(t, "finish", result[4]["type"])
	})

	t.Run("handles unknown part type", func(t *testing.T) {
		// This would be a custom type that doesn't match any known types
		parts := []ContentPart{
			&unknownContentPart{Content: "unknown"},
		}

		_, err := marshallParts(parts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown part type")
	})
}

func TestUnmarshallParts(t *testing.T) {
	t.Run("unmarshalls various content parts", func(t *testing.T) {
		jsonData := `[
			{"type": "text", "data": {"text": "Hello, world!"}},
			{"type": "image_url", "data": {"url": "https://example.com/image.jpg"}},
			{"type": "tool_call", "data": {"id": "call-1", "name": "test-tool", "input": "{\"param\": \"value\"}"}},
			{"type": "tool_result", "data": {"call_id": "call-1", "content": "Result content"}},
			{"type": "finish", "data": {"reason": "stop", "time": 1234567890}}
		]`

		parts, err := unmarshallParts([]byte(jsonData))
		require.NoError(t, err)
		assert.Len(t, parts, 5)

		// Check text content
		textPart, ok := parts[0].(TextContent)
		assert.True(t, ok)
		assert.Equal(t, "Hello, world!", textPart.Text)

		// Check image URL content
		imagePart, ok := parts[1].(ImageURLContent)
		assert.True(t, ok)
		assert.Equal(t, "https://example.com/image.jpg", imagePart.URL)

		// Check tool call
		toolCallPart, ok := parts[2].(ToolCall)
		assert.True(t, ok)
		assert.Equal(t, "call-1", toolCallPart.ID)
		assert.Equal(t, "test-tool", toolCallPart.Name)

		// Check tool result
		toolResultPart, ok := parts[3].(ToolResult)
		assert.True(t, ok)
		assert.Equal(t, "call-1", toolResultPart.ToolCallID)
		assert.Equal(t, "Result content", toolResultPart.Content)

		// Check finish
		finishPart, ok := parts[4].(Finish)
		assert.True(t, ok)
		assert.Equal(t, "stop", finishPart.Reason)
		assert.Equal(t, int64(1234567890), finishPart.Time)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		invalidJSON := `[{"type": "text", "data": }]`

		_, err := unmarshallParts([]byte(invalidJSON))
		assert.Error(t, err)
	})

	t.Run("handles unknown part type", func(t *testing.T) {
		jsonData := `[{"type": "unknown_type", "data": {"content": "unknown"}}]`

		_, err := unmarshallParts([]byte(jsonData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown part type")
	})
}

func TestRoundTripSerialization(t *testing.T) {
	originalParts := []ContentPart{
		TextContent{Text: "Hello, world!"},
		ImageURLContent{URL: "https://example.com/image.jpg", Detail: "high"},
		BinaryContent{Data: []byte("binary data"), MimeType: "image/png"},
		ToolCall{ID: "call-1", Name: "test-tool", Input: `{"param": "value"}`},
		ToolResult{ToolCallID: "call-1", Content: "Result content", IsError: false},
		Finish{Reason: "stop", Time: 1234567890},
	}

	// Marshall to JSON
	data, err := marshallParts(originalParts)
	require.NoError(t, err)

	// Unmarshall back
	parsedParts, err := unmarshallParts(data)
	require.NoError(t, err)

	// Compare lengths
	assert.Len(t, parsedParts, len(originalParts))

	// Compare each part
	for i, original := range originalParts {
		parsed := parsedParts[i]
		assert.IsType(t, original, parsed)

		switch orig := original.(type) {
		case TextContent:
			parsedText := parsed.(TextContent)
			assert.Equal(t, orig.Text, parsedText.Text)
		case ImageURLContent:
			parsedImage := parsed.(ImageURLContent)
			assert.Equal(t, orig.URL, parsedImage.URL)
			assert.Equal(t, orig.Detail, parsedImage.Detail)
		case BinaryContent:
			parsedBinary := parsed.(BinaryContent)
			assert.Equal(t, orig.Data, parsedBinary.Data)
			assert.Equal(t, orig.MimeType, parsedBinary.MimeType)
		case ToolCall:
			parsedToolCall := parsed.(ToolCall)
			assert.Equal(t, orig.ID, parsedToolCall.ID)
			assert.Equal(t, orig.Name, parsedToolCall.Name)
			assert.Equal(t, orig.Input, parsedToolCall.Input)
		case ToolResult:
			parsedToolResult := parsed.(ToolResult)
			assert.Equal(t, orig.ToolCallID, parsedToolResult.ToolCallID)
			assert.Equal(t, orig.Content, parsedToolResult.Content)
			assert.Equal(t, orig.IsError, parsedToolResult.IsError)
		case Finish:
			parsedFinish := parsed.(Finish)
			assert.Equal(t, orig.Reason, parsedFinish.Reason)
			assert.Equal(t, orig.Time, parsedFinish.Time)
		}
	}
}

// unknownContentPart is a test type that doesn't implement ContentPart interface
type unknownContentPart struct {
	Content string `json:"content"`
}

func (u *unknownContentPart) isPart() {}

func TestMessageFinishPart(t *testing.T) {
	t.Run("finds finish part", func(t *testing.T) {
		message := Message{
			Parts: []ContentPart{
				TextContent{Text: "Hello"},
				Finish{Reason: "stop", Time: 1234567890},
			},
		}

		finishPart := message.FinishPart()
		assert.NotNil(t, finishPart)
		assert.Equal(t, "stop", finishPart.Reason)
		assert.Equal(t, int64(1234567890), finishPart.Time)
	})

	t.Run("returns nil when no finish part", func(t *testing.T) {
		message := Message{
			Parts: []ContentPart{
				TextContent{Text: "Hello"},
			},
		}

		finishPart := message.FinishPart()
		assert.Nil(t, finishPart)
	})
}
