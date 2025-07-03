package session

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/db"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSession_Struct(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	session := Session{
		ID:               "session-123",
		ParentSessionID:  "parent-456",
		Title:            "Test Session",
		MessageCount:     10,
		PromptTokens:     1000,
		CompletionTokens: 500,
		SummaryMessageID: "msg-789",
		Cost:             0.05,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.Equal(t, "session-123", session.ID)
	assert.Equal(t, "parent-456", session.ParentSessionID)
	assert.Equal(t, "Test Session", session.Title)
	assert.Equal(t, int64(10), session.MessageCount)
	assert.Equal(t, int64(1000), session.PromptTokens)
	assert.Equal(t, int64(500), session.CompletionTokens)
	assert.Equal(t, "msg-789", session.SummaryMessageID)
	assert.Equal(t, 0.05, session.Cost)
	assert.Equal(t, now, session.CreatedAt)
	assert.Equal(t, now, session.UpdatedAt)
}

// Mock Querier for testing
type MockQuerier struct {
	mock.Mock
}

// Session methods
func (m *MockQuerier) CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Session), args.Error(1)
}

func (m *MockQuerier) GetSessionByID(ctx context.Context, id string) (db.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Session), args.Error(1)
}

func (m *MockQuerier) ListSessions(ctx context.Context) ([]db.Session, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Session), args.Error(1)
}

func (m *MockQuerier) UpdateSession(ctx context.Context, params db.UpdateSessionParams) (db.Session, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Session), args.Error(1)
}

func (m *MockQuerier) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// File methods (stub implementations - not used in session tests)
func (m *MockQuerier) CreateFile(ctx context.Context, params db.CreateFileParams) (db.File, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.File), args.Error(1)
}

func (m *MockQuerier) GetFile(ctx context.Context, id string) (db.File, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.File), args.Error(1)
}

func (m *MockQuerier) GetFileByPathAndSession(ctx context.Context, params db.GetFileByPathAndSessionParams) (db.File, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.File), args.Error(1)
}

func (m *MockQuerier) ListFilesByPath(ctx context.Context, path string) ([]db.File, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]db.File), args.Error(1)
}

func (m *MockQuerier) ListFilesBySession(ctx context.Context, sessionID string) ([]db.File, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]db.File), args.Error(1)
}

func (m *MockQuerier) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]db.File, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]db.File), args.Error(1)
}

func (m *MockQuerier) ListNewFiles(ctx context.Context) ([]db.File, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.File), args.Error(1)
}

func (m *MockQuerier) UpdateFile(ctx context.Context, params db.UpdateFileParams) (db.File, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.File), args.Error(1)
}

func (m *MockQuerier) DeleteFile(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) DeleteSessionFiles(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// Message methods (stub implementations - not used in session tests)
func (m *MockQuerier) CreateMessage(ctx context.Context, params db.CreateMessageParams) (db.Message, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Message), args.Error(1)
}

func (m *MockQuerier) GetMessage(ctx context.Context, id string) (db.Message, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Message), args.Error(1)
}

func (m *MockQuerier) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.Message, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]db.Message), args.Error(1)
}

func (m *MockQuerier) UpdateMessage(ctx context.Context, params db.UpdateMessageParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) DeleteMessage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	t.Parallel()

	mockQuerier := &MockQuerier{}
	service := NewService(mockQuerier)

	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
	assert.Implements(t, (*pubsub.Suscriber[Session])(nil), service)
}

func TestService_FromDBItem(t *testing.T) {
	t.Parallel()

	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier).(*service)

	now := time.Now().Unix()
	dbSession := db.Session{
		ID:               "session-123",
		ParentSessionID:  sql.NullString{String: "parent-456", Valid: true},
		Title:            "Test Session",
		MessageCount:     10,
		PromptTokens:     1000,
		CompletionTokens: 500,
		SummaryMessageID: sql.NullString{String: "msg-789", Valid: true},
		Cost:             0.05,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	session := svc.fromDBItem(dbSession)

	assert.Equal(t, "session-123", session.ID)
	assert.Equal(t, "parent-456", session.ParentSessionID)
	assert.Equal(t, "Test Session", session.Title)
	assert.Equal(t, int64(10), session.MessageCount)
	assert.Equal(t, int64(1000), session.PromptTokens)
	assert.Equal(t, int64(500), session.CompletionTokens)
	assert.Equal(t, "msg-789", session.SummaryMessageID)
	assert.Equal(t, 0.05, session.Cost)
	assert.Equal(t, now, session.CreatedAt)
	assert.Equal(t, now, session.UpdatedAt)
}

func TestService_FromDBItem_NullValues(t *testing.T) {
	t.Parallel()

	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier).(*service)

	now := time.Now().Unix()
	dbSession := db.Session{
		ID:               "session-123",
		ParentSessionID:  sql.NullString{}, // null parent
		Title:            "Test Session",
		MessageCount:     5,
		PromptTokens:     100,
		CompletionTokens: 50,
		SummaryMessageID: sql.NullString{}, // null summary
		Cost:             0.01,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	session := svc.fromDBItem(dbSession)

	assert.Equal(t, "session-123", session.ID)
	assert.Empty(t, session.ParentSessionID) // 应该是空字符串
	assert.Empty(t, session.SummaryMessageID) // 应该是空字符串
	assert.Equal(t, "Test Session", session.Title)
}

func TestService_Create(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	title := "Test Session"

	now := time.Now().Unix()
	expectedDBSession := db.Session{
		ID:               "generated-uuid",
		ParentSessionID:  sql.NullString{},
		Title:            title,
		MessageCount:     0,
		PromptTokens:     0,
		CompletionTokens: 0,
		SummaryMessageID: sql.NullString{},
		Cost:             0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockQuerier.On("CreateSession", ctx, mock.MatchedBy(func(params db.CreateSessionParams) bool {
		return params.Title == title
	})).Return(expectedDBSession, nil)

	session, err := svc.Create(ctx, title)

	assert.NoError(t, err)
	assert.Equal(t, "generated-uuid", session.ID)
	assert.Equal(t, title, session.Title)
	assert.Empty(t, session.ParentSessionID)
	mockQuerier.AssertExpectations(t)
}

func TestService_CreateTaskSession(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	toolCallID := "tool-123"
	parentSessionID := "parent-456"
	title := "Task Session"

	now := time.Now().Unix()
	expectedDBSession := db.Session{
		ID:               toolCallID,
		ParentSessionID:  sql.NullString{String: parentSessionID, Valid: true},
		Title:            title,
		MessageCount:     0,
		PromptTokens:     0,
		CompletionTokens: 0,
		SummaryMessageID: sql.NullString{},
		Cost:             0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockQuerier.On("CreateSession", ctx, mock.MatchedBy(func(params db.CreateSessionParams) bool {
		return params.ID == toolCallID && 
			   params.ParentSessionID.String == parentSessionID && 
			   params.ParentSessionID.Valid &&
			   params.Title == title
	})).Return(expectedDBSession, nil)

	session, err := svc.CreateTaskSession(ctx, toolCallID, parentSessionID, title)

	assert.NoError(t, err)
	assert.Equal(t, toolCallID, session.ID)
	assert.Equal(t, parentSessionID, session.ParentSessionID)
	assert.Equal(t, title, session.Title)
	mockQuerier.AssertExpectations(t)
}

func TestService_CreateTitleSession(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	parentSessionID := "parent-456"
	expectedID := "title-" + parentSessionID

	now := time.Now().Unix()
	expectedDBSession := db.Session{
		ID:               expectedID,
		ParentSessionID:  sql.NullString{String: parentSessionID, Valid: true},
		Title:            "Generate a title",
		MessageCount:     0,
		PromptTokens:     0,
		CompletionTokens: 0,
		SummaryMessageID: sql.NullString{},
		Cost:             0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockQuerier.On("CreateSession", ctx, mock.MatchedBy(func(params db.CreateSessionParams) bool {
		return params.ID == expectedID &&
			   params.ParentSessionID.String == parentSessionID &&
			   params.ParentSessionID.Valid &&
			   params.Title == "Generate a title"
	})).Return(expectedDBSession, nil)

	session, err := svc.CreateTitleSession(ctx, parentSessionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, session.ID)
	assert.Equal(t, parentSessionID, session.ParentSessionID)
	assert.Equal(t, "Generate a title", session.Title)
	mockQuerier.AssertExpectations(t)
}

func TestService_Get(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	sessionID := "session-123"

	now := time.Now().Unix()
	expectedDBSession := db.Session{
		ID:               sessionID,
		ParentSessionID:  sql.NullString{},
		Title:            "Test Session",
		MessageCount:     5,
		PromptTokens:     100,
		CompletionTokens: 50,
		SummaryMessageID: sql.NullString{},
		Cost:             0.01,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockQuerier.On("GetSessionByID", ctx, sessionID).Return(expectedDBSession, nil)

	session, err := svc.Get(ctx, sessionID)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, "Test Session", session.Title)
	mockQuerier.AssertExpectations(t)
}

func TestService_List(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()

	now := time.Now().Unix()
	expectedDBSessions := []db.Session{
		{
			ID:               "session-1",
			Title:            "Session 1",
			MessageCount:     5,
			PromptTokens:     100,
			CompletionTokens: 50,
			Cost:             0.01,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "session-2",
			Title:            "Session 2",
			MessageCount:     3,
			PromptTokens:     80,
			CompletionTokens: 40,
			Cost:             0.008,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	mockQuerier.On("ListSessions", ctx).Return(expectedDBSessions, nil)

	sessions, err := svc.List(ctx)

	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "session-1", sessions[0].ID)
	assert.Equal(t, "Session 1", sessions[0].Title)
	assert.Equal(t, "session-2", sessions[1].ID)
	assert.Equal(t, "Session 2", sessions[1].Title)
	mockQuerier.AssertExpectations(t)
}

func TestService_Save(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	now := time.Now().Unix()

	inputSession := Session{
		ID:               "session-123",
		Title:            "Updated Session",
		PromptTokens:     1000,
		CompletionTokens: 500,
		SummaryMessageID: "msg-789",
		Cost:             0.05,
	}

	expectedDBSession := db.Session{
		ID:               inputSession.ID,
		Title:            inputSession.Title,
		MessageCount:     10,
		PromptTokens:     inputSession.PromptTokens,
		CompletionTokens: inputSession.CompletionTokens,
		SummaryMessageID: sql.NullString{String: inputSession.SummaryMessageID, Valid: true},
		Cost:             inputSession.Cost,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockQuerier.On("UpdateSession", ctx, mock.MatchedBy(func(params db.UpdateSessionParams) bool {
		return params.ID == inputSession.ID &&
			   params.Title == inputSession.Title &&
			   params.PromptTokens == inputSession.PromptTokens &&
			   params.CompletionTokens == inputSession.CompletionTokens &&
			   params.SummaryMessageID.String == inputSession.SummaryMessageID &&
			   params.SummaryMessageID.Valid &&
			   params.Cost == inputSession.Cost
	})).Return(expectedDBSession, nil)

	session, err := svc.Save(ctx, inputSession)

	assert.NoError(t, err)
	assert.Equal(t, inputSession.ID, session.ID)
	assert.Equal(t, inputSession.Title, session.Title)
	assert.Equal(t, inputSession.PromptTokens, session.PromptTokens)
	assert.Equal(t, inputSession.CompletionTokens, session.CompletionTokens)
	assert.Equal(t, inputSession.SummaryMessageID, session.SummaryMessageID)
	assert.Equal(t, inputSession.Cost, session.Cost)
	mockQuerier.AssertExpectations(t)
}

func TestService_Save_EmptySummaryMessageID(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	now := time.Now().Unix()

	inputSession := Session{
		ID:               "session-123",
		Title:            "Updated Session",
		SummaryMessageID: "", // 空的summary message ID
		Cost:             0.05,
	}

	expectedDBSession := db.Session{
		ID:               inputSession.ID,
		Title:            inputSession.Title,
		SummaryMessageID: sql.NullString{}, // 应该是null
		Cost:             inputSession.Cost,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockQuerier.On("UpdateSession", ctx, mock.MatchedBy(func(params db.UpdateSessionParams) bool {
		return params.ID == inputSession.ID &&
			   params.SummaryMessageID.String == "" &&
			   !params.SummaryMessageID.Valid // 应该是无效的
	})).Return(expectedDBSession, nil)

	session, err := svc.Save(ctx, inputSession)

	assert.NoError(t, err)
	assert.Equal(t, inputSession.ID, session.ID)
	assert.Empty(t, session.SummaryMessageID)
	mockQuerier.AssertExpectations(t)
}

func TestService_Delete(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	sessionID := "session-123"

	now := time.Now().Unix()
	expectedDBSession := db.Session{
		ID:               sessionID,
		Title:            "Session to Delete",
		MessageCount:     5,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// 首先模拟Get调用
	mockQuerier.On("GetSessionByID", ctx, sessionID).Return(expectedDBSession, nil)
	// 然后模拟Delete调用
	mockQuerier.On("DeleteSession", ctx, sessionID).Return(nil)

	err := svc.Delete(ctx, sessionID)

	assert.NoError(t, err)
	mockQuerier.AssertExpectations(t)
}

// 错误场景测试
func TestService_Create_Error(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	title := "Test Session"

	expectedError := assert.AnError
	mockQuerier.On("CreateSession", ctx, mock.Anything).Return(db.Session{}, expectedError)

	session, err := svc.Create(ctx, title)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, session.ID)
	mockQuerier.AssertExpectations(t)
}

func TestService_Get_Error(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	sessionID := "session-123"

	expectedError := assert.AnError
	mockQuerier.On("GetSessionByID", ctx, sessionID).Return(db.Session{}, expectedError)

	session, err := svc.Get(ctx, sessionID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, session.ID)
	mockQuerier.AssertExpectations(t)
}

// 基准测试
func BenchmarkService_FromDBItem(b *testing.B) {
	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier).(*service)

	now := time.Now().Unix()
	dbSession := db.Session{
		ID:               "session-123",
		ParentSessionID:  sql.NullString{String: "parent-456", Valid: true},
		Title:            "Test Session",
		MessageCount:     10,
		PromptTokens:     1000,
		CompletionTokens: 500,
		SummaryMessageID: sql.NullString{String: "msg-789", Valid: true},
		Cost:             0.05,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.fromDBItem(dbSession)
	}
}

func BenchmarkNewService(b *testing.B) {
	mockQuerier := &MockQuerier{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewService(mockQuerier)
	}
}

// 并发测试
func TestService_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试")
	}

	mockQuerier := &MockQuerier{}
	svc := NewService(mockQuerier)

	ctx := context.Background()
	now := time.Now().Unix()

	// 设置mock期望（允许多次调用）
	mockQuerier.On("CreateSession", ctx, mock.Anything).Return(db.Session{
		ID:        "test-session",
		Title:     "Test Session",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil).Maybe()

	// 并发创建会话
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			title := "Concurrent Session"
			_, err := svc.Create(ctx, title)
			errors <- err
		}(i)
	}

	// 收集错误
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		if err != nil {
			t.Errorf("并发操作失败: %v", err)
		}
	}
}