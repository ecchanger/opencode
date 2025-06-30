package diff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionType(t *testing.T) {
	assert.Equal(t, "add", string(ActionAdd))
	assert.Equal(t, "delete", string(ActionDelete))
	assert.Equal(t, "update", string(ActionUpdate))
}

func TestNewDiffError(t *testing.T) {
	err := NewDiffError("test error message")
	assert.Equal(t, "test error message", err.Error())
	assert.Equal(t, "test error message", err.message)
}

func TestFileError(t *testing.T) {
	err := fileError("Update", "Missing File", "/path/to/file")
	assert.Equal(t, "Update File Error: Missing File: /path/to/file", err.Error())
}

func TestContextError(t *testing.T) {
	err := contextError(5, "context content", false)
	assert.Equal(t, "Invalid Context 5:\ncontext content", err.Error())

	err = contextError(10, "eof context", true)
	assert.Equal(t, "Invalid EOF Context 10:\neof context", err.Error())
}

func TestNewParser(t *testing.T) {
	currentFiles := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	lines := []string{"line1", "line2", "line3"}

	parser := NewParser(currentFiles, lines)

	assert.NotNil(t, parser)
	assert.Equal(t, currentFiles, parser.currentFiles)
	assert.Equal(t, lines, parser.lines)
	assert.Equal(t, 0, parser.index)
	assert.Equal(t, 0, parser.fuzz)
	assert.NotNil(t, parser.patch.Actions)
}

func TestParser_IsDone(t *testing.T) {
	lines := []string{
		"some line",
		"*** End Patch",
		"another line",
	}
	parser := NewParser(nil, lines)

	// Not done at start
	parser.index = 0
	assert.False(t, parser.isDone([]string{"*** End Patch"}))

	// Done when at End Patch
	parser.index = 1
	assert.True(t, parser.isDone([]string{"*** End Patch"}))

	// Done when past end
	parser.index = 10
	assert.True(t, parser.isDone([]string{"*** End Patch"}))
}

func TestParser_StartsWith(t *testing.T) {
	lines := []string{"*** Update File: test.txt", "other line"}
	parser := NewParser(nil, lines)

	// Test with string
	assert.True(t, parser.startsWith("*** Update File:"))
	assert.False(t, parser.startsWith("*** Delete File:"))

	// Test with slice
	assert.True(t, parser.startsWith([]string{"*** Update File:", "*** Delete File:"}))
	assert.False(t, parser.startsWith([]string{"*** Delete File:", "*** Add File:"}))
}

func TestParser_ReadStr(t *testing.T) {
	lines := []string{
		"*** Update File: test.txt",
		"some other line",
	}
	parser := NewParser(nil, lines)

	// Read with prefix match
	result := parser.readStr("*** Update File: ", false)
	assert.Equal(t, "test.txt", result)
	assert.Equal(t, 1, parser.index) // Should advance index

	// Read without prefix match
	result = parser.readStr("*** Delete File: ", false)
	assert.Equal(t, "", result)
	assert.Equal(t, 1, parser.index) // Should not advance index

	// Read everything
	result = parser.readStr("some other", true)
	assert.Equal(t, "some other line", result)
	assert.Equal(t, 2, parser.index) // Should advance index
}

func TestIdentifyFilesNeeded(t *testing.T) {
	patchText := `*** Begin Patch
*** Update File: file1.txt
@@ line1
@@ line2
*** Delete File: file2.txt
*** Update File: file3.txt
@@ line3
*** End Patch`

	files := IdentifyFilesNeeded(patchText)

	assert.Len(t, files, 3)
	assert.Contains(t, files, "file1.txt")
	assert.Contains(t, files, "file2.txt")
	assert.Contains(t, files, "file3.txt")
}

func TestIdentifyFilesAdded(t *testing.T) {
	patchText := `*** Begin Patch
*** Add File: new1.txt
+content1
*** Update File: existing.txt
@@ modification
*** Add File: new2.txt
+content2
*** End Patch`

	files := IdentifyFilesAdded(patchText)

	assert.Len(t, files, 2)
	assert.Contains(t, files, "new1.txt")
	assert.Contains(t, files, "new2.txt")
}

func TestTextToPatch_InvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "missing begin",
			text: "some content\n*** End Patch",
		},
		{
			name: "missing end",
			text: "*** Begin Patch\nsome content",
		},
		{
			name: "empty",
			text: "",
		},
		{
			name: "single line",
			text: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := TextToPatch(tt.text, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid patch text")
		})
	}
}

func TestTextToPatch_ValidPatch(t *testing.T) {
	patchText := `*** Begin Patch
*** Delete File: delete_me.txt
*** End Patch`

	origFiles := map[string]string{
		"delete_me.txt": "content to delete",
	}

	patch, fuzz, err := TextToPatch(patchText, origFiles)
	assert.NoError(t, err)
	assert.Equal(t, 0, fuzz)

	assert.Len(t, patch.Actions, 1)
	action, exists := patch.Actions["delete_me.txt"]
	assert.True(t, exists)
	assert.Equal(t, ActionDelete, action.Type)
}

func TestChunk(t *testing.T) {
	chunk := Chunk{
		OrigIndex: 5,
		DelLines:  []string{"line to delete"},
		InsLines:  []string{"line to insert"},
	}

	assert.Equal(t, 5, chunk.OrigIndex)
	assert.Len(t, chunk.DelLines, 1)
	assert.Len(t, chunk.InsLines, 1)
	assert.Equal(t, "line to delete", chunk.DelLines[0])
	assert.Equal(t, "line to insert", chunk.InsLines[0])
}

func TestPatchAction(t *testing.T) {
	content := "new file content"
	movePath := "/new/path"

	action := PatchAction{
		Type:     ActionAdd,
		NewFile:  &content,
		Chunks:   []Chunk{},
		MovePath: &movePath,
	}

	assert.Equal(t, ActionAdd, action.Type)
	assert.NotNil(t, action.NewFile)
	assert.Equal(t, content, *action.NewFile)
	assert.NotNil(t, action.MovePath)
	assert.Equal(t, movePath, *action.MovePath)
	assert.Empty(t, action.Chunks)
}

func TestFileChange(t *testing.T) {
	oldContent := "old content"
	newContent := "new content"
	movePath := "/new/location"

	change := FileChange{
		Type:       ActionUpdate,
		OldContent: &oldContent,
		NewContent: &newContent,
		MovePath:   &movePath,
	}

	assert.Equal(t, ActionUpdate, change.Type)
	assert.NotNil(t, change.OldContent)
	assert.Equal(t, oldContent, *change.OldContent)
	assert.NotNil(t, change.NewContent)
	assert.Equal(t, newContent, *change.NewContent)
	assert.NotNil(t, change.MovePath)
	assert.Equal(t, movePath, *change.MovePath)
}

func TestCommit(t *testing.T) {
	changes := map[string]FileChange{
		"file1.txt": {Type: ActionAdd},
		"file2.txt": {Type: ActionDelete},
	}

	commit := Commit{Changes: changes}

	assert.Len(t, commit.Changes, 2)
	assert.Equal(t, ActionAdd, commit.Changes["file1.txt"].Type)
	assert.Equal(t, ActionDelete, commit.Changes["file2.txt"].Type)
}

func TestLoadFiles(t *testing.T) {
	// Mock open function
	openFn := func(path string) (string, error) {
		switch path {
		case "file1.txt":
			return "content1", nil
		case "file2.txt":
			return "content2", nil
		default:
			return "", NewDiffError("file not found")
		}
	}

	// Test successful load
	files, err := LoadFiles([]string{"file1.txt", "file2.txt"}, openFn)
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "content1", files["file1.txt"])
	assert.Equal(t, "content2", files["file2.txt"])

	// Test with missing file
	_, err = LoadFiles([]string{"missing.txt"}, openFn)
	assert.Error(t, err)
}

func TestApplyCommit(t *testing.T) {
	writtenFiles := make(map[string]string)
	removedFiles := make([]string, 0)

	writeFn := func(path, content string) error {
		writtenFiles[path] = content
		return nil
	}

	removeFn := func(path string) error {
		removedFiles = append(removedFiles, path)
		return nil
	}

	newContent1 := "new content 1"
	newContent2 := "new content 2"

	commit := Commit{
		Changes: map[string]FileChange{
			"new_file.txt": {
				Type:       ActionAdd,
				NewContent: &newContent1,
			},
			"updated_file.txt": {
				Type:       ActionUpdate,
				NewContent: &newContent2,
			},
			"deleted_file.txt": {
				Type: ActionDelete,
			},
		},
	}

	err := ApplyCommit(commit, writeFn, removeFn)
	assert.NoError(t, err)

	// Check written files
	assert.Len(t, writtenFiles, 2)
	assert.Equal(t, "new content 1", writtenFiles["new_file.txt"])
	assert.Equal(t, "new content 2", writtenFiles["updated_file.txt"])

	// Check removed files
	assert.Len(t, removedFiles, 1)
	assert.Contains(t, removedFiles, "deleted_file.txt")
}

func TestAssembleChanges(t *testing.T) {
	orig := map[string]string{
		"file1.txt": "original content 1",
		"file2.txt": "original content 2",
		"file3.txt": "original content 3",
	}

	updated := map[string]string{
		"file1.txt": "updated content 1",  // Modified
		"file2.txt": "original content 2", // Unchanged
		"file4.txt": "new content 4",      // Added
		// file3.txt is missing (deleted)
	}

	commit := AssembleChanges(orig, updated)

	assert.Len(t, commit.Changes, 3) // Should have changes for file1, file3, and file4

	// Check updated file
	change1, exists := commit.Changes["file1.txt"]
	assert.True(t, exists)
	assert.Equal(t, ActionUpdate, change1.Type)
	assert.Equal(t, "updated content 1", *change1.NewContent)

	// Check deleted file
	change3, exists := commit.Changes["file3.txt"]
	assert.True(t, exists)
	assert.Equal(t, ActionDelete, change3.Type)

	// Check added file
	change4, exists := commit.Changes["file4.txt"]
	assert.True(t, exists)
	assert.Equal(t, ActionAdd, change4.Type)
	assert.Equal(t, "new content 4", *change4.NewContent)

	// Unchanged file should not be in changes
	_, exists = commit.Changes["file2.txt"]
	assert.False(t, exists)
}

func TestTryFindMatch(t *testing.T) {
	lines := []string{
		"line 1",
		"line 2  ", // with trailing spaces
		"  line 3", // with leading spaces
		"line 4",
	}

	context := []string{"line 2", "line 3"}

	// Test exact match (should fail due to whitespace)
	idx, fuzz := tryFindMatch(lines, context, 0, func(a, b string) bool {
		return a == b
	})
	assert.Equal(t, -1, idx)
	assert.Equal(t, 0, fuzz)

	// Test trimmed match
	idx, fuzz = tryFindMatch(lines, context, 0, func(a, b string) bool {
		return strings.TrimSpace(a) == strings.TrimSpace(b)
	})
	assert.Equal(t, 1, idx)
	assert.NotEqual(t, 0, fuzz) // Should have some fuzz for trimmed match
}

func TestFindContextCore(t *testing.T) {
	lines := []string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
	}

	// Test exact match
	context := []string{"line 2", "line 3"}
	idx, fuzz := findContextCore(lines, context, 0)
	assert.Equal(t, 1, idx)
	assert.Equal(t, 0, fuzz)

	// Test no match
	context = []string{"missing line"}
	idx, fuzz = findContextCore(lines, context, 0)
	assert.Equal(t, -1, idx)
	assert.Equal(t, 0, fuzz)

	// Test empty context
	idx, fuzz = findContextCore(lines, []string{}, 2)
	assert.Equal(t, 2, idx) // Should return start position
	assert.Equal(t, 0, fuzz)
}

func TestFindContext(t *testing.T) {
	lines := []string{"line 1", "line 2", "line 3"}
	context := []string{"line 2"}

	// Test normal context finding
	idx, fuzz := findContext(lines, context, 0, false)
	assert.Equal(t, 1, idx)
	assert.Equal(t, 0, fuzz)

	// Test EOF context finding
	idx, fuzz = findContext(lines, context, 0, true)
	assert.Equal(t, 1, idx)
	// Fuzz might be different for EOF context
}

// Test edge cases and error conditions
func TestParser_EdgeCases(t *testing.T) {
	// Test parser with empty lines
	parser := NewParser(nil, []string{})
	assert.True(t, parser.isDone([]string{"any"}))

	// Test readStr beyond bounds
	result := parser.readStr("prefix", false)
	assert.Equal(t, "", result)
}

func TestPeekNextSection(t *testing.T) {
	lines := []string{
		" context line 1",
		"-deleted line",
		"+added line",
		" context line 2",
		"@@", // End marker
	}

	old, chunks, endIndex, eof := peekNextSection(lines, 0)

	assert.False(t, eof)
	assert.Equal(t, 4, endIndex) // Should stop at @@
	assert.Len(t, old, 3)        // context + deleted lines
	assert.Len(t, chunks, 1)     // One chunk with the deletion/addition

	chunk := chunks[0]
	assert.Equal(t, 1, chunk.OrigIndex) // After first context line
	assert.Len(t, chunk.DelLines, 1)
	assert.Len(t, chunk.InsLines, 1)
	assert.Equal(t, "deleted line", chunk.DelLines[0])
	assert.Equal(t, "added line", chunk.InsLines[0])
}

func TestDiffError_Interface(t *testing.T) {
	// Test that DiffError implements error interface
	var err error = NewDiffError("test")
	assert.NotNil(t, err)
	assert.Equal(t, "test", err.Error())
}
