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

func setupTestDBWithFiles(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	
	// Create the files table
	_, err = db.Exec(`
		CREATE TABLE files (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			path TEXT NOT NULL,
			content TEXT NOT NULL,
			version TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			is_new INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)
	
	return db
}

func TestCreateFile(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	tests := []struct {
		name   string
		params CreateFileParams
		want   func(t *testing.T, result File)
	}{
		{
			name: "successful file creation",
			params: CreateFileParams{
				ID:        "file-1",
				SessionID: "session-1",
				Path:      "/path/to/file1.txt",
				Content:   "Hello, world!",
				Version:   "v1",
			},
			want: func(t *testing.T, result File) {
				assert.Equal(t, "file-1", result.ID)
				assert.Equal(t, "session-1", result.SessionID)
				assert.Equal(t, "/path/to/file1.txt", result.Path)
				assert.Equal(t, "Hello, world!", result.Content)
				assert.Equal(t, "v1", result.Version)
				assert.Greater(t, result.CreatedAt, int64(0))
				assert.Greater(t, result.UpdatedAt, int64(0))
			},
		},
		{
			name: "file with complex content",
			params: CreateFileParams{
				ID:        "file-2",
				SessionID: "session-1",
				Path:      "/src/main.go",
				Content:   "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, Go!\")\n}",
				Version:   "v2.1.0",
			},
			want: func(t *testing.T, result File) {
				assert.Equal(t, "file-2", result.ID)
				assert.Equal(t, "session-1", result.SessionID)
				assert.Equal(t, "/src/main.go", result.Path)
				assert.Contains(t, result.Content, "package main")
				assert.Contains(t, result.Content, "fmt.Println")
				assert.Equal(t, "v2.1.0", result.Version)
			},
		},
		{
			name: "file with empty content",
			params: CreateFileParams{
				ID:        "file-3",
				SessionID: "session-2",
				Path:      "/empty.txt",
				Content:   "",
				Version:   "v1.0",
			},
			want: func(t *testing.T, result File) {
				assert.Equal(t, "file-3", result.ID)
				assert.Equal(t, "session-2", result.SessionID)
				assert.Equal(t, "/empty.txt", result.Path)
				assert.Equal(t, "", result.Content)
				assert.Equal(t, "v1.0", result.Version)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := queries.CreateFile(ctx, tt.params)
			require.NoError(t, err)
			tt.want(t, result)
		})
	}
}

func TestCreateFileDuplicate(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	params := CreateFileParams{
		ID:        "duplicate-file",
		SessionID: "session-1",
		Path:      "/path/to/file.txt",
		Content:   "content",
		Version:   "v1",
	}
	
	// Create first file
	_, err := queries.CreateFile(ctx, params)
	require.NoError(t, err)
	
	// Try to create duplicate file
	_, err = queries.CreateFile(ctx, params)
	assert.Error(t, err)
}

func TestGetFile(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a file first
	createParams := CreateFileParams{
		ID:        "test-file-get",
		SessionID: "session-1",
		Path:      "/test/file.txt",
		Content:   "Test content for get",
		Version:   "v1.0",
	}
	
	created, err := queries.CreateFile(ctx, createParams)
	require.NoError(t, err)
	
	// Get the file by ID
	result, err := queries.GetFile(ctx, "test-file-get")
	require.NoError(t, err)
	
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, created.SessionID, result.SessionID)
	assert.Equal(t, created.Path, result.Path)
	assert.Equal(t, created.Content, result.Content)
	assert.Equal(t, created.Version, result.Version)
	assert.Equal(t, created.CreatedAt, result.CreatedAt)
	assert.Equal(t, created.UpdatedAt, result.UpdatedAt)
}

func TestGetFileNotFound(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	_, err := queries.GetFile(ctx, "nonexistent-file")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetFileByPathAndSession(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create multiple files with same path but different sessions
	files := []CreateFileParams{
		{
			ID:        "file-1",
			SessionID: "session-1",
			Path:      "/shared/file.txt",
			Content:   "Content from session 1",
			Version:   "v1",
		},
		{
			ID:        "file-2",
			SessionID: "session-2",
			Path:      "/shared/file.txt",
			Content:   "Content from session 2",
			Version:   "v2",
		},
		{
			ID:        "file-3",
			SessionID: "session-1",
			Path:      "/shared/file.txt",
			Content:   "Updated content from session 1",
			Version:   "v3",
		},
	}
	
	// Create files with slight delay to test ordering
	for i, file := range files {
		_, err := queries.CreateFile(ctx, file)
		require.NoError(t, err)
		if i < len(files)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}
	}
	
	// Get file by path and session (should return latest)
	result, err := queries.GetFileByPathAndSession(ctx, GetFileByPathAndSessionParams{
		Path:      "/shared/file.txt",
		SessionID: "session-1",
	})
	require.NoError(t, err)
	
	// Should return the latest file for session-1
	assert.Equal(t, "file-3", result.ID)
	assert.Equal(t, "session-1", result.SessionID)
	assert.Equal(t, "/shared/file.txt", result.Path)
	assert.Equal(t, "Updated content from session 1", result.Content)
	assert.Equal(t, "v3", result.Version)
}

func TestGetFileByPathAndSessionNotFound(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	_, err := queries.GetFileByPathAndSession(ctx, GetFileByPathAndSessionParams{
		Path:      "/nonexistent/file.txt",
		SessionID: "session-1",
	})
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestListFilesByPath(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create multiple files with same path
	files := []CreateFileParams{
		{
			ID:        "file-1",
			SessionID: "session-1",
			Path:      "/shared/file.txt",
			Content:   "Version 1",
			Version:   "v1",
		},
		{
			ID:        "file-2",
			SessionID: "session-2",
			Path:      "/shared/file.txt",
			Content:   "Version 2",
			Version:   "v2",
		},
		{
			ID:        "file-3",
			SessionID: "session-1",
			Path:      "/different/file.txt",
			Content:   "Different path",
			Version:   "v1",
		},
		{
			ID:        "file-4",
			SessionID: "session-1",
			Path:      "/shared/file.txt",
			Content:   "Version 3",
			Version:   "v3",
		},
	}
	
	// Create files with slight delay to test ordering
	for i, file := range files {
		_, err := queries.CreateFile(ctx, file)
		require.NoError(t, err)
		if i < len(files)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}
	}
	
	// List files by path
	result, err := queries.ListFilesByPath(ctx, "/shared/file.txt")
	require.NoError(t, err)
	
	// Should have 3 files with same path, ordered by created_at DESC
	assert.Len(t, result, 3)
	assert.Equal(t, "file-4", result[0].ID) // Created last
	assert.Equal(t, "file-2", result[1].ID) // Created second
	assert.Equal(t, "file-1", result[2].ID) // Created first
	
	// Verify all files have same path
	for _, file := range result {
		assert.Equal(t, "/shared/file.txt", file.Path)
	}
}

func TestListFilesByPathEmpty(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	result, err := queries.ListFilesByPath(ctx, "/nonexistent/file.txt")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListFilesBySession(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create files in different sessions
	files := []CreateFileParams{
		{
			ID:        "file-1",
			SessionID: "session-1",
			Path:      "/file1.txt",
			Content:   "Content 1",
			Version:   "v1",
		},
		{
			ID:        "file-2",
			SessionID: "session-1",
			Path:      "/file2.txt",
			Content:   "Content 2",
			Version:   "v1",
		},
		{
			ID:        "file-3",
			SessionID: "session-2",
			Path:      "/file3.txt",
			Content:   "Content 3",
			Version:   "v1",
		},
		{
			ID:        "file-4",
			SessionID: "session-1",
			Path:      "/file4.txt",
			Content:   "Content 4",
			Version:   "v1",
		},
	}
	
	// Create files with slight delay to test ordering
	for i, file := range files {
		_, err := queries.CreateFile(ctx, file)
		require.NoError(t, err)
		if i < len(files)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}
	}
	
	// List files for session-1
	result, err := queries.ListFilesBySession(ctx, "session-1")
	require.NoError(t, err)
	
	// Should have 3 files for session-1, ordered by created_at ASC
	assert.Len(t, result, 3)
	assert.Equal(t, "file-1", result[0].ID) // Created first
	assert.Equal(t, "file-2", result[1].ID) // Created second
	assert.Equal(t, "file-4", result[2].ID) // Created last
	
	// Verify all files belong to session-1
	for _, file := range result {
		assert.Equal(t, "session-1", file.SessionID)
	}
}

func TestListFilesBySessionEmpty(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	result, err := queries.ListFilesBySession(ctx, "nonexistent-session")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListLatestSessionFiles(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create files with multiple versions of same paths
	files := []CreateFileParams{
		{
			ID:        "file-1",
			SessionID: "session-1",
			Path:      "/file1.txt",
			Content:   "Version 1 of file1",
			Version:   "v1",
		},
		{
			ID:        "file-2",
			SessionID: "session-1",
			Path:      "/file2.txt",
			Content:   "Version 1 of file2",
			Version:   "v1",
		},
		{
			ID:        "file-3",
			SessionID: "session-2",
			Path:      "/file1.txt",
			Content:   "Version 2 of file1 (different session)",
			Version:   "v2",
		},
		{
			ID:        "file-4",
			SessionID: "session-1",
			Path:      "/file1.txt",
			Content:   "Version 3 of file1",
			Version:   "v3",
		},
	}
	
	// Create files with slight delay to test ordering
	for i, file := range files {
		_, err := queries.CreateFile(ctx, file)
		require.NoError(t, err)
		if i < len(files)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}
	}
	
	// List latest session files for session-1
	result, err := queries.ListLatestSessionFiles(ctx, "session-1")
	require.NoError(t, err)
	
	// Should have 2 files (latest version of each path), ordered by path
	assert.Len(t, result, 2)
	
	// First file should be /file1.txt (latest version)
	assert.Equal(t, "file-4", result[0].ID)
	assert.Equal(t, "/file1.txt", result[0].Path)
	assert.Equal(t, "Version 3 of file1", result[0].Content)
	
	// Second file should be /file2.txt
	assert.Equal(t, "file-2", result[1].ID)
	assert.Equal(t, "/file2.txt", result[1].Path)
	assert.Equal(t, "Version 1 of file2", result[1].Content)
}

func TestListLatestSessionFilesEmpty(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	result, err := queries.ListLatestSessionFiles(ctx, "nonexistent-session")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListNewFiles(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create files with is_new flag
	files := []CreateFileParams{
		{
			ID:        "file-1",
			SessionID: "session-1",
			Path:      "/file1.txt",
			Content:   "Content 1",
			Version:   "v1",
		},
		{
			ID:        "file-2",
			SessionID: "session-1",
			Path:      "/file2.txt",
			Content:   "Content 2",
			Version:   "v1",
		},
	}
	
	// Create files
	for _, file := range files {
		_, err := queries.CreateFile(ctx, file)
		require.NoError(t, err)
	}
	
	// Manually set is_new flag for file-2
	_, err := db.Exec("UPDATE files SET is_new = 1 WHERE id = ?", "file-2")
	require.NoError(t, err)
	
	// List new files
	result, err := queries.ListNewFiles(ctx)
	require.NoError(t, err)
	
	// Should have 1 new file
	assert.Len(t, result, 1)
	assert.Equal(t, "file-2", result[0].ID)
	assert.Equal(t, "/file2.txt", result[0].Path)
}

func TestListNewFilesEmpty(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	result, err := queries.ListNewFiles(ctx)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestUpdateFile(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a file first
	createParams := CreateFileParams{
		ID:        "test-file-update",
		SessionID: "session-1",
		Path:      "/test/file.txt",
		Content:   "Original content",
		Version:   "v1.0",
	}
	
	created, err := queries.CreateFile(ctx, createParams)
	require.NoError(t, err)
	
	// Update the file
	updateParams := UpdateFileParams{
		ID:      "test-file-update",
		Content: "Updated content with more information",
		Version: "v2.0",
	}
	
	updated, err := queries.UpdateFile(ctx, updateParams)
	require.NoError(t, err)
	
	// Verify updates
	assert.Equal(t, "test-file-update", updated.ID)
	assert.Equal(t, "Updated content with more information", updated.Content)
	assert.Equal(t, "v2.0", updated.Version)
	
	// Verify unchanged fields
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, created.SessionID, updated.SessionID)
	assert.Equal(t, created.Path, updated.Path)
	assert.Equal(t, created.CreatedAt, updated.CreatedAt)
	
	// Verify updated_at was changed
	assert.Greater(t, updated.UpdatedAt, created.UpdatedAt)
}

func TestUpdateFileNotFound(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	updateParams := UpdateFileParams{
		ID:      "nonexistent-file",
		Content: "Updated content",
		Version: "v2.0",
	}
	
	_, err := queries.UpdateFile(ctx, updateParams)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteFile(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create a file first
	createParams := CreateFileParams{
		ID:        "test-file-delete",
		SessionID: "session-1",
		Path:      "/test/file.txt",
		Content:   "File to delete",
		Version:   "v1.0",
	}
	
	_, err := queries.CreateFile(ctx, createParams)
	require.NoError(t, err)
	
	// Verify file exists
	_, err = queries.GetFile(ctx, "test-file-delete")
	require.NoError(t, err)
	
	// Delete the file
	err = queries.DeleteFile(ctx, "test-file-delete")
	require.NoError(t, err)
	
	// Verify file is deleted
	_, err = queries.GetFile(ctx, "test-file-delete")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteFileNotFound(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Delete non-existent file should not error
	err := queries.DeleteFile(ctx, "nonexistent-file")
	assert.NoError(t, err)
}

func TestDeleteSessionFiles(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Create files in different sessions
	files := []CreateFileParams{
		{
			ID:        "file-1",
			SessionID: "session-1",
			Path:      "/file1.txt",
			Content:   "Content 1",
			Version:   "v1",
		},
		{
			ID:        "file-2",
			SessionID: "session-1",
			Path:      "/file2.txt",
			Content:   "Content 2",
			Version:   "v1",
		},
		{
			ID:        "file-3",
			SessionID: "session-2",
			Path:      "/file3.txt",
			Content:   "Content 3",
			Version:   "v1",
		},
	}
	
	for _, file := range files {
		_, err := queries.CreateFile(ctx, file)
		require.NoError(t, err)
	}
	
	// Verify all files exist
	result1, err := queries.ListFilesBySession(ctx, "session-1")
	require.NoError(t, err)
	assert.Len(t, result1, 2)
	
	result2, err := queries.ListFilesBySession(ctx, "session-2")
	require.NoError(t, err)
	assert.Len(t, result2, 1)
	
	// Delete all files for session-1
	err = queries.DeleteSessionFiles(ctx, "session-1")
	require.NoError(t, err)
	
	// Verify session-1 files are deleted
	result1After, err := queries.ListFilesBySession(ctx, "session-1")
	require.NoError(t, err)
	assert.Empty(t, result1After)
	
	// Verify session-2 files are still there
	result2After, err := queries.ListFilesBySession(ctx, "session-2")
	require.NoError(t, err)
	assert.Len(t, result2After, 1)
	assert.Equal(t, "file-3", result2After[0].ID)
}

func TestDeleteSessionFilesNotFound(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Delete files for non-existent session should not error
	err := queries.DeleteSessionFiles(ctx, "nonexistent-session")
	assert.NoError(t, err)
}

func TestFileWithTransaction(t *testing.T) {
	db := setupTestDBWithFiles(t)
	defer db.Close()
	
	queries := New(db)
	ctx := context.Background()
	
	// Start transaction
	tx, err := db.Begin()
	require.NoError(t, err)
	
	txQueries := queries.WithTx(tx)
	
	// Create file in transaction
	createParams := CreateFileParams{
		ID:        "tx-file",
		SessionID: "session-1",
		Path:      "/tx/file.txt",
		Content:   "Transaction file",
		Version:   "v1.0",
	}
	
	created, err := txQueries.CreateFile(ctx, createParams)
	require.NoError(t, err)
	assert.Equal(t, "tx-file", created.ID)
	
	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err)
	
	// Verify file doesn't exist after rollback
	_, err = queries.GetFile(ctx, "tx-file")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}