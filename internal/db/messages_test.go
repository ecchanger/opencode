package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDBWithMessages(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	
	// Create the messages table
	_, err = db.Exec(`
		CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			parts TEXT NOT NULL,
			model TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			finished_at INTEGER
		)
	`)
	require.NoError(t, err)
	
	return db
}

func TestCreateMessage(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	tests := []struct {
		name   string
		params CreateMessageParams
		want   func(t *testing.T, result Message)
	}{
		{
			name: "successful message creation with model",
			params: CreateMessageParams{
				ID:        "msg-1",
				SessionID: "session-1",
				Role:      "user",
				Parts:     "Hello, world!",
				Model:     sql.NullString{String: "gpt-4", Valid: true},
			},
			want: func(t *testing.T, result Message) {
				assert.Equal(t, "msg-1", result.ID)
				assert.Equal(t, "session-1", result.SessionID)
				assert.Equal(t, "user", result.Role)
				assert.Equal(t, "Hello, world!", result.Parts)
				assert.True(t, result.Model.Valid)
				assert.Equal(t, "gpt-4", result.Model.String)
				assert.Greater(t, result.CreatedAt, int64(0))
				assert.Greater(t, result.UpdatedAt, int64(0))
				assert.False(t, result.FinishedAt.Valid)
			},
		},
		{
			name: "successful message creation without model",
			params: CreateMessageParams{
				ID:        "msg-2",
				SessionID: "session-1",
				Role:      "assistant",
				Parts:     "Hi there!",
				Model:     sql.NullString{String: "", Valid: false},
			},
			want: func(t *testing.T, result Message) {
				assert.Equal(t, "msg-2", result.ID)
				assert.Equal(t, "session-1", result.SessionID)
				assert.Equal(t, "assistant", result.Role)
				assert.Equal(t, "Hi there!", result.Parts)
				assert.False(t, result.Model.Valid)
				assert.Greater(t, result.CreatedAt, int64(0))
				assert.Greater(t, result.UpdatedAt, int64(0))
				assert.False(t, result.FinishedAt.Valid)
			},
		},
		{
			name: "message with complex parts content",
			params: CreateMessageParams{
				ID:        "msg-3",
				SessionID: "session-2",
				Role:      "user",
				Parts:     `{"type": "text", "content": "This is a complex message with JSON content"}`,
				Model:     sql.NullString{String: "claude-3", Valid: true},
			},
			want: func(t *testing.T, result Message) {
				assert.Equal(t, "msg-3", result.ID)
				assert.Equal(t, "session-2", result.SessionID)
				assert.Equal(t, "user", result.Role)
				assert.Contains(t, result.Parts, "complex message")
				assert.Contains(t, result.Parts, "JSON content")
				assert.True(t, result.Model.Valid)
				assert.Equal(t, "claude-3", result.Model.String)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := queries.CreateMessage(ctx, tt.params)
			require.NoError(t, err)
			tt.want(t, result)
		})
	}
}

func TestCreateMessageDuplicate(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	params := CreateMessageParams{
		ID:        "duplicate-msg",
		SessionID: "session-1",
		Role:      "user",
		Parts:     "Hello!",
		Model:     sql.NullString{String: "gpt-4", Valid: true},
	}
	
	// Create first message
	_, err := queries.CreateMessage(ctx, params)
	require.NoError(t, err)
	
	// Try to create duplicate message
	_, err = queries.CreateMessage(ctx, params)
	assert.Error(t, err)
}

func TestGetMessage(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a message first
	createParams := CreateMessageParams{
		ID:        "test-msg-get",
		SessionID: "session-1",
		Role:      "user",
		Parts:     "Test message for get",
		Model:     sql.NullString{String: "gpt-4", Valid: true},
	}
	
	created, err := queries.CreateMessage(ctx, createParams)
	require.NoError(t, err)
	
	// Get the message by ID
	result, err := queries.GetMessage(ctx, "test-msg-get")
	require.NoError(t, err)
	
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, created.SessionID, result.SessionID)
	assert.Equal(t, created.Role, result.Role)
	assert.Equal(t, created.Parts, result.Parts)
	assert.Equal(t, created.Model, result.Model)
	assert.Equal(t, created.CreatedAt, result.CreatedAt)
	assert.Equal(t, created.UpdatedAt, result.UpdatedAt)
	assert.Equal(t, created.FinishedAt, result.FinishedAt)
}

func TestGetMessageNotFound(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	_, err := queries.GetMessage(ctx, "nonexistent-message")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestListMessagesBySession(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create messages in different sessions
	messages := []CreateMessageParams{
		{
			ID:        "msg-1",
			SessionID: "session-1",
			Role:      "user",
			Parts:     "First message",
			Model:     sql.NullString{String: "gpt-4", Valid: true},
		},
		{
			ID:        "msg-2",
			SessionID: "session-1",
			Role:      "assistant",
			Parts:     "Second message",
			Model:     sql.NullString{String: "gpt-4", Valid: true},
		},
		{
			ID:        "msg-3",
			SessionID: "session-2",
			Role:      "user",
			Parts:     "Different session message",
			Model:     sql.NullString{String: "claude-3", Valid: true},
		},
		{
			ID:        "msg-4",
			SessionID: "session-1",
			Role:      "user",
			Parts:     "Third message",
			Model:     sql.NullString{String: "gpt-4", Valid: true},
		},
	}
	
	// Create messages with slight delay to test ordering
	for i, msg := range messages {
		_, err := queries.CreateMessage(ctx, msg)
		require.NoError(t, err)
		if i < len(messages)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}
	}
	
	// List messages for session-1
	result, err := queries.ListMessagesBySession(ctx, "session-1")
	require.NoError(t, err)
	
	// Should have 3 messages for session-1, ordered by created_at ASC
	assert.Len(t, result, 3)
	assert.Equal(t, "msg-1", result[0].ID) // Created first
	assert.Equal(t, "msg-2", result[1].ID) // Created second
	assert.Equal(t, "msg-4", result[2].ID) // Created last
	
	// Verify all messages belong to session-1
	for _, msg := range result {
		assert.Equal(t, "session-1", msg.SessionID)
	}
	
	// List messages for session-2
	result2, err := queries.ListMessagesBySession(ctx, "session-2")
	require.NoError(t, err)
	
	// Should have 1 message for session-2
	assert.Len(t, result2, 1)
	assert.Equal(t, "msg-3", result2[0].ID)
	assert.Equal(t, "session-2", result2[0].SessionID)
}

func TestListMessagesBySessionEmpty(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	result, err := queries.ListMessagesBySession(ctx, "nonexistent-session")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestUpdateMessage(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a message first
	createParams := CreateMessageParams{
		ID:        "test-msg-update",
		SessionID: "session-1",
		Role:      "assistant",
		Parts:     "Original content",
		Model:     sql.NullString{String: "gpt-4", Valid: true},
	}
	
	created, err := queries.CreateMessage(ctx, createParams)
	require.NoError(t, err)
	
	// Update the message
	finishedAt := time.Now().Unix()
	updateParams := UpdateMessageParams{
		ID:         "test-msg-update",
		Parts:      "Updated content with more information",
		FinishedAt: sql.NullInt64{Int64: finishedAt, Valid: true},
	}
	
	err = queries.UpdateMessage(ctx, updateParams)
	require.NoError(t, err)
	
	// Get the updated message
	updated, err := queries.GetMessage(ctx, "test-msg-update")
	require.NoError(t, err)
	
	// Verify updates
	assert.Equal(t, "test-msg-update", updated.ID)
	assert.Equal(t, "Updated content with more information", updated.Parts)
	assert.True(t, updated.FinishedAt.Valid)
	assert.Equal(t, finishedAt, updated.FinishedAt.Int64)
	
	// Verify unchanged fields
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, created.SessionID, updated.SessionID)
	assert.Equal(t, created.Role, updated.Role)
	assert.Equal(t, created.Model, updated.Model)
	assert.Equal(t, created.CreatedAt, updated.CreatedAt)
	
	// Verify updated_at was changed
	assert.Greater(t, updated.UpdatedAt, created.UpdatedAt)
}

func TestUpdateMessageNotFound(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	updateParams := UpdateMessageParams{
		ID:         "nonexistent-message",
		Parts:      "Updated content",
		FinishedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
	}
	
	err := queries.UpdateMessage(ctx, updateParams)
	assert.NoError(t, err) // SQL UPDATE doesn't error on no rows affected
}

func TestDeleteMessage(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a message first
	createParams := CreateMessageParams{
		ID:        "test-msg-delete",
		SessionID: "session-1",
		Role:      "user",
		Parts:     "Message to delete",
		Model:     sql.NullString{String: "gpt-4", Valid: true},
	}
	
	_, err := queries.CreateMessage(ctx, createParams)
	require.NoError(t, err)
	
	// Verify message exists
	_, err = queries.GetMessage(ctx, "test-msg-delete")
	require.NoError(t, err)
	
	// Delete the message
	err = queries.DeleteMessage(ctx, "test-msg-delete")
	require.NoError(t, err)
	
	// Verify message is deleted
	_, err = queries.GetMessage(ctx, "test-msg-delete")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteMessageNotFound(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Delete non-existent message should not error
	err := queries.DeleteMessage(ctx, "nonexistent-message")
	assert.NoError(t, err)
}

func TestDeleteSessionMessages(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create messages in different sessions
	messages := []CreateMessageParams{
		{
			ID:        "msg-1",
			SessionID: "session-1",
			Role:      "user",
			Parts:     "Message 1",
			Model:     sql.NullString{String: "gpt-4", Valid: true},
		},
		{
			ID:        "msg-2",
			SessionID: "session-1",
			Role:      "assistant",
			Parts:     "Message 2",
			Model:     sql.NullString{String: "gpt-4", Valid: true},
		},
		{
			ID:        "msg-3",
			SessionID: "session-2",
			Role:      "user",
			Parts:     "Message 3",
			Model:     sql.NullString{String: "claude-3", Valid: true},
		},
	}
	
	for _, msg := range messages {
		_, err := queries.CreateMessage(ctx, msg)
		require.NoError(t, err)
	}
	
	// Verify all messages exist
	result1, err := queries.ListMessagesBySession(ctx, "session-1")
	require.NoError(t, err)
	assert.Len(t, result1, 2)
	
	result2, err := queries.ListMessagesBySession(ctx, "session-2")
	require.NoError(t, err)
	assert.Len(t, result2, 1)
	
	// Delete all messages for session-1
	err = queries.DeleteSessionMessages(ctx, "session-1")
	require.NoError(t, err)
	
	// Verify session-1 messages are deleted
	result1After, err := queries.ListMessagesBySession(ctx, "session-1")
	require.NoError(t, err)
	assert.Empty(t, result1After)
	
	// Verify session-2 messages are still there
	result2After, err := queries.ListMessagesBySession(ctx, "session-2")
	require.NoError(t, err)
	assert.Len(t, result2After, 1)
	assert.Equal(t, "msg-3", result2After[0].ID)
}

func TestDeleteSessionMessagesNotFound(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Delete messages for non-existent session should not error
	err := queries.DeleteSessionMessages(ctx, "nonexistent-session")
	assert.NoError(t, err)
}

func TestMessageWithTransaction(t *testing.T) {
	db := setupTestDBWithMessages(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Start transaction
	tx, err := db.Begin()
	require.NoError(t, err)
	
	txQueries := queries.WithTx(tx)
	
	// Create message in transaction
	createParams := CreateMessageParams{
		ID:        "tx-message",
		SessionID: "session-1",
		Role:      "user",
		Parts:     "Transaction message",
		Model:     sql.NullString{String: "gpt-4", Valid: true},
	}
	
	created, err := txQueries.CreateMessage(ctx, createParams)
	require.NoError(t, err)
	assert.Equal(t, "tx-message", created.ID)
	
	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err)
	
	// Verify message doesn't exist after rollback
	_, err = queries.GetMessage(ctx, "tx-message")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}