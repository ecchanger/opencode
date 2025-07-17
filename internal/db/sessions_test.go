package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	
	// Create the sessions table
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			parent_session_id TEXT,
			title TEXT NOT NULL,
			message_count INTEGER NOT NULL DEFAULT 0,
			prompt_tokens INTEGER NOT NULL DEFAULT 0,
			completion_tokens INTEGER NOT NULL DEFAULT 0,
			cost REAL NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			summary_message_id TEXT
		)
	`)
	require.NoError(t, err)
	
	return db
}

func TestCreateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	tests := []struct {
		name   string
		params CreateSessionParams
		want   func(t *testing.T, result Session)
	}{
		{
			name: "successful session creation",
			params: CreateSessionParams{
				ID:               "test-session-1",
				ParentSessionID:  sql.NullString{String: "", Valid: false},
				Title:            "Test Session",
				MessageCount:     0,
				PromptTokens:     100,
				CompletionTokens: 50,
				Cost:             0.01,
			},
			want: func(t *testing.T, result Session) {
				assert.Equal(t, "test-session-1", result.ID)
				assert.Equal(t, "Test Session", result.Title)
				assert.Equal(t, int64(0), result.MessageCount)
				assert.Equal(t, int64(100), result.PromptTokens)
				assert.Equal(t, int64(50), result.CompletionTokens)
				assert.Equal(t, 0.01, result.Cost)
				assert.False(t, result.ParentSessionID.Valid)
				assert.False(t, result.SummaryMessageID.Valid)
				assert.Greater(t, result.CreatedAt, int64(0))
				assert.Greater(t, result.UpdatedAt, int64(0))
			},
		},
		{
			name: "session with parent session",
			params: CreateSessionParams{
				ID:               "test-session-2",
				ParentSessionID:  sql.NullString{String: "parent-session-1", Valid: true},
				Title:            "Child Session",
				MessageCount:     5,
				PromptTokens:     200,
				CompletionTokens: 150,
				Cost:             0.02,
			},
			want: func(t *testing.T, result Session) {
				assert.Equal(t, "test-session-2", result.ID)
				assert.Equal(t, "Child Session", result.Title)
				assert.Equal(t, int64(5), result.MessageCount)
				assert.Equal(t, int64(200), result.PromptTokens)
				assert.Equal(t, int64(150), result.CompletionTokens)
				assert.Equal(t, 0.02, result.Cost)
				assert.True(t, result.ParentSessionID.Valid)
				assert.Equal(t, "parent-session-1", result.ParentSessionID.String)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := queries.CreateSession(ctx, tt.params)
			require.NoError(t, err)
			tt.want(t, result)
		})
	}
}

func TestCreateSessionDuplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	params := CreateSessionParams{
		ID:               "duplicate-session",
		ParentSessionID:  sql.NullString{String: "", Valid: false},
		Title:            "Test Session",
		MessageCount:     0,
		PromptTokens:     100,
		CompletionTokens: 50,
		Cost:             0.01,
	}
	
	// Create first session
	_, err := queries.CreateSession(ctx, params)
	require.NoError(t, err)
	
	// Try to create duplicate session
	_, err = queries.CreateSession(ctx, params)
	assert.Error(t, err)
}

func TestGetSessionByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a session first
	createParams := CreateSessionParams{
		ID:               "test-session-get",
		ParentSessionID:  sql.NullString{String: "", Valid: false},
		Title:            "Test Session Get",
		MessageCount:     3,
		PromptTokens:     150,
		CompletionTokens: 75,
		Cost:             0.015,
	}
	
	created, err := queries.CreateSession(ctx, createParams)
	require.NoError(t, err)
	
	// Get the session by ID
	result, err := queries.GetSessionByID(ctx, "test-session-get")
	require.NoError(t, err)
	
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, created.Title, result.Title)
	assert.Equal(t, created.MessageCount, result.MessageCount)
	assert.Equal(t, created.PromptTokens, result.PromptTokens)
	assert.Equal(t, created.CompletionTokens, result.CompletionTokens)
	assert.Equal(t, created.Cost, result.Cost)
	assert.Equal(t, created.CreatedAt, result.CreatedAt)
	assert.Equal(t, created.UpdatedAt, result.UpdatedAt)
}

func TestGetSessionByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	_, err := queries.GetSessionByID(ctx, "nonexistent-session")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestListSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create multiple sessions (with parent sessions to test filtering)
	sessions := []CreateSessionParams{
		{
			ID:               "parent-session-1",
			ParentSessionID:  sql.NullString{String: "", Valid: false},
			Title:            "Parent Session 1",
			MessageCount:     1,
			PromptTokens:     100,
			CompletionTokens: 50,
			Cost:             0.01,
		},
		{
			ID:               "parent-session-2",
			ParentSessionID:  sql.NullString{String: "", Valid: false},
			Title:            "Parent Session 2",
			MessageCount:     2,
			PromptTokens:     200,
			CompletionTokens: 100,
			Cost:             0.02,
		},
		{
			ID:               "child-session-1",
			ParentSessionID:  sql.NullString{String: "parent-session-1", Valid: true},
			Title:            "Child Session 1",
			MessageCount:     3,
			PromptTokens:     150,
			CompletionTokens: 75,
			Cost:             0.015,
		},
	}
	
	// Create sessions with slight delay to test ordering
	for i, session := range sessions {
		_, err := queries.CreateSession(ctx, session)
		require.NoError(t, err)
		if i < len(sessions)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}
	}
	
	// List sessions (should only return parent sessions)
	result, err := queries.ListSessions(ctx)
	require.NoError(t, err)
	
	// Should only have 2 parent sessions, ordered by created_at DESC
	assert.Len(t, result, 2)
	assert.Equal(t, "parent-session-2", result[0].ID) // Created last
	assert.Equal(t, "parent-session-1", result[1].ID) // Created first
	
	// Verify all sessions are parent sessions
	for _, session := range result {
		assert.False(t, session.ParentSessionID.Valid)
	}
}

func TestListSessionsEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	result, err := queries.ListSessions(ctx)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestUpdateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a session first
	createParams := CreateSessionParams{
		ID:               "test-session-update",
		ParentSessionID:  sql.NullString{String: "", Valid: false},
		Title:            "Original Title",
		MessageCount:     1,
		PromptTokens:     100,
		CompletionTokens: 50,
		Cost:             0.01,
	}
	
	created, err := queries.CreateSession(ctx, createParams)
	require.NoError(t, err)
	
	// Update the session
	updateParams := UpdateSessionParams{
		ID:               "test-session-update",
		Title:            "Updated Title",
		PromptTokens:     300,
		CompletionTokens: 150,
		Cost:             0.03,
		SummaryMessageID: sql.NullString{String: "summary-msg-1", Valid: true},
	}
	
	updated, err := queries.UpdateSession(ctx, updateParams)
	require.NoError(t, err)
	
	// Verify updates
	assert.Equal(t, "test-session-update", updated.ID)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, int64(300), updated.PromptTokens)
	assert.Equal(t, int64(150), updated.CompletionTokens)
	assert.Equal(t, 0.03, updated.Cost)
	assert.True(t, updated.SummaryMessageID.Valid)
	assert.Equal(t, "summary-msg-1", updated.SummaryMessageID.String)
	
	// Verify unchanged fields
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, created.MessageCount, updated.MessageCount) // Not updated
	assert.Equal(t, created.ParentSessionID, updated.ParentSessionID) // Not updated
	assert.Equal(t, created.CreatedAt, updated.CreatedAt) // Not updated
}

func TestUpdateSessionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	updateParams := UpdateSessionParams{
		ID:               "nonexistent-session",
		Title:            "Updated Title",
		PromptTokens:     300,
		CompletionTokens: 150,
		Cost:             0.03,
		SummaryMessageID: sql.NullString{String: "", Valid: false},
	}
	
	_, err := queries.UpdateSession(ctx, updateParams)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a session first
	createParams := CreateSessionParams{
		ID:               "test-session-delete",
		ParentSessionID:  sql.NullString{String: "", Valid: false},
		Title:            "Session to Delete",
		MessageCount:     1,
		PromptTokens:     100,
		CompletionTokens: 50,
		Cost:             0.01,
	}
	
	_, err := queries.CreateSession(ctx, createParams)
	require.NoError(t, err)
	
	// Verify session exists
	_, err = queries.GetSessionByID(ctx, "test-session-delete")
	require.NoError(t, err)
	
	// Delete the session
	err = queries.DeleteSession(ctx, "test-session-delete")
	require.NoError(t, err)
	
	// Verify session is deleted
	_, err = queries.GetSessionByID(ctx, "test-session-delete")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteSessionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Delete non-existent session should not error
	err := queries.DeleteSession(ctx, "nonexistent-session")
	assert.NoError(t, err)
}

func TestSessionWithTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Start transaction
	tx, err := db.Begin()
	require.NoError(t, err)
	
	txQueries := queries.WithTx(tx)
	
	// Create session in transaction
	createParams := CreateSessionParams{
		ID:               "tx-session",
		ParentSessionID:  sql.NullString{String: "", Valid: false},
		Title:            "Transaction Session",
		MessageCount:     1,
		PromptTokens:     100,
		CompletionTokens: 50,
		Cost:             0.01,
	}
	
	created, err := txQueries.CreateSession(ctx, createParams)
	require.NoError(t, err)
	assert.Equal(t, "tx-session", created.ID)
	
	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err)
	
	// Verify session doesn't exist after rollback
	_, err = queries.GetSessionByID(ctx, "tx-session")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}