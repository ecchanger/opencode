package diff

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestLineType(t *testing.T) {
	assert.Equal(t, 0, int(LineContext))
	assert.Equal(t, 1, int(LineAdded))
	assert.Equal(t, 2, int(LineRemoved))
}

func TestParseUnifiedDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected DiffResult
		wantErr  bool
	}{
		{
			name: "simple diff",
			diff: `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line1
-old line
+new line
 line3`,
			expected: DiffResult{
				OldFile: "test.txt",
				NewFile: "test.txt",
				Hunks: []Hunk{
					{
						Header: "@@ -1,3 +1,3 @@",
						Lines: []DiffLine{
							{OldLineNo: 1, NewLineNo: 1, Kind: LineContext, Content: "line1"},
							{OldLineNo: 2, NewLineNo: 0, Kind: LineRemoved, Content: "old line"},
							{OldLineNo: 0, NewLineNo: 2, Kind: LineAdded, Content: "new line"},
							{OldLineNo: 3, NewLineNo: 3, Kind: LineContext, Content: "line3"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple hunks",
			diff: `--- a/multi.txt
+++ b/multi.txt
@@ -1,2 +1,2 @@
-old1
+new1
 context
@@ -5,2 +5,2 @@
 context2
-old2
+new2`,
			expected: DiffResult{
				OldFile: "multi.txt",
				NewFile: "multi.txt",
				Hunks: []Hunk{
					{
						Header: "@@ -1,2 +1,2 @@",
						Lines: []DiffLine{
							{OldLineNo: 1, NewLineNo: 0, Kind: LineRemoved, Content: "old1"},
							{OldLineNo: 0, NewLineNo: 1, Kind: LineAdded, Content: "new1"},
							{OldLineNo: 2, NewLineNo: 2, Kind: LineContext, Content: "context"},
						},
					},
					{
						Header: "@@ -5,2 +5,2 @@",
						Lines: []DiffLine{
							{OldLineNo: 5, NewLineNo: 5, Kind: LineContext, Content: "context2"},
							{OldLineNo: 6, NewLineNo: 0, Kind: LineRemoved, Content: "old2"},
							{OldLineNo: 0, NewLineNo: 6, Kind: LineAdded, Content: "new2"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty diff",
			diff: "",
			expected: DiffResult{
				Hunks: []Hunk{},
			},
			wantErr: false,
		},
		{
			name: "diff with empty lines",
			diff: `--- a/empty.txt
+++ b/empty.txt
@@ -1,3 +1,3 @@
 line1

 line3`,
			expected: DiffResult{
				OldFile: "empty.txt",
				NewFile: "empty.txt",
				Hunks: []Hunk{
					{
						Header: "@@ -1,3 +1,3 @@",
						Lines: []DiffLine{
							{OldLineNo: 1, NewLineNo: 1, Kind: LineContext, Content: "line1"},
							{OldLineNo: 2, NewLineNo: 2, Kind: LineContext, Content: ""},
							{OldLineNo: 3, NewLineNo: 3, Kind: LineContext, Content: "line3"},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseUnifiedDiff(tt.diff)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.OldFile, result.OldFile)
				assert.Equal(t, tt.expected.NewFile, result.NewFile)
				assert.Len(t, result.Hunks, len(tt.expected.Hunks))

				for i, expectedHunk := range tt.expected.Hunks {
					if i < len(result.Hunks) {
						assert.Equal(t, expectedHunk.Header, result.Hunks[i].Header)
						assert.Len(t, result.Hunks[i].Lines, len(expectedHunk.Lines))

						for j, expectedLine := range expectedHunk.Lines {
							if j < len(result.Hunks[i].Lines) {
								line := result.Hunks[i].Lines[j]
								assert.Equal(t, expectedLine.OldLineNo, line.OldLineNo, "OldLineNo mismatch at hunk %d, line %d", i, j)
								assert.Equal(t, expectedLine.NewLineNo, line.NewLineNo, "NewLineNo mismatch at hunk %d, line %d", i, j)
								assert.Equal(t, expectedLine.Kind, line.Kind, "Kind mismatch at hunk %d, line %d", i, j)
								assert.Equal(t, expectedLine.Content, line.Content, "Content mismatch at hunk %d, line %d", i, j)
							}
						}
					}
				}
			}
		})
	}
}

func TestHighlightIntralineChanges(t *testing.T) {
	hunk := &Hunk{
		Lines: []DiffLine{
			{Kind: LineRemoved, Content: "Hello world"},
			{Kind: LineAdded, Content: "Hello Go"},
			{Kind: LineContext, Content: "unchanged"},
		},
	}

	HighlightIntralineChanges(hunk)

	// After highlighting, removed and added lines should have segments
	assert.NotEmpty(t, hunk.Lines[0].Segments)
	assert.NotEmpty(t, hunk.Lines[1].Segments)
	// Context lines should not have segments (or empty segments)
	assert.Empty(t, hunk.Lines[2].Segments)
}

func TestPairLines(t *testing.T) {
	lines := []DiffLine{
		{Kind: LineRemoved, Content: "removed1"},
		{Kind: LineAdded, Content: "added1"},
		{Kind: LineRemoved, Content: "removed2"},
		{Kind: LineContext, Content: "context"},
		{Kind: LineAdded, Content: "added2"},
	}

	pairs := pairLines(lines)

	assert.Len(t, pairs, 4)

	// First pair: removed + added
	assert.NotNil(t, pairs[0].left)
	assert.NotNil(t, pairs[0].right)
	assert.Equal(t, "removed1", pairs[0].left.Content)
	assert.Equal(t, "added1", pairs[0].right.Content)

	// Second pair: removed only
	assert.NotNil(t, pairs[1].left)
	assert.Nil(t, pairs[1].right)
	assert.Equal(t, "removed2", pairs[1].left.Content)

	// Third pair: context (both sides same)
	assert.NotNil(t, pairs[2].left)
	assert.NotNil(t, pairs[2].right)
	assert.Equal(t, "context", pairs[2].left.Content)
	assert.Equal(t, "context", pairs[2].right.Content)

	// Fourth pair: added only
	assert.Nil(t, pairs[3].left)
	assert.NotNil(t, pairs[3].right)
	assert.Equal(t, "added2", pairs[3].right.Content)
}

func TestSyntaxHighlight(t *testing.T) {
	var buf bytes.Buffer
	source := `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`

	err := SyntaxHighlight(&buf, source, "main.go", "terminal", lipgloss.Color("#000000"))
	assert.NoError(t, err)

	// Output should not be empty
	assert.NotEmpty(t, buf.String())

	// Should contain some ANSI color codes (basic check)
	output := buf.String()
	assert.True(t, strings.Contains(output, "package") || strings.Contains(output, "main"))
}

func TestSyntaxHighlight_InvalidFormatter(t *testing.T) {
	var buf bytes.Buffer
	source := "test content"

	// Should use fallback formatter and not error
	err := SyntaxHighlight(&buf, source, "test.txt", "invalid-formatter", lipgloss.Color("#000000"))
	assert.NoError(t, err)
}

func TestGenerateDiff(t *testing.T) {
	beforeContent := `line1
line2
line3`

	afterContent := `line1
modified line2
line3
new line4`

	diff, oldLines, newLines := GenerateDiff(beforeContent, afterContent, "test.txt")

	assert.NotEmpty(t, diff)
	assert.Equal(t, 3, oldLines) // 3 lines in before
	assert.Equal(t, 4, newLines) // 4 lines in after

	// Should contain unified diff markers
	assert.Contains(t, diff, "@@")
	assert.Contains(t, diff, "-line2")
	assert.Contains(t, diff, "+modified line2")
	assert.Contains(t, diff, "+new line4")
}

func TestFormatDiff(t *testing.T) {
	diffText := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line1
-old line
+new line
 line3`

	formatted, err := FormatDiff(diffText)
	assert.NoError(t, err)
	assert.NotEmpty(t, formatted)
}

func TestFormatDiff_InvalidDiff(t *testing.T) {
	invalidDiff := "this is not a valid diff"

	_, err := FormatDiff(invalidDiff)
	// Should handle invalid diff gracefully
	assert.NoError(t, err) // The function might not error on invalid diff, just return empty
}

func TestDiffLine(t *testing.T) {
	line := DiffLine{
		OldLineNo: 10,
		NewLineNo: 15,
		Kind:      LineAdded,
		Content:   "test content",
		Segments: []Segment{
			{Start: 0, End: 4, Type: LineAdded, Text: "test"},
		},
	}

	assert.Equal(t, 10, line.OldLineNo)
	assert.Equal(t, 15, line.NewLineNo)
	assert.Equal(t, LineAdded, line.Kind)
	assert.Equal(t, "test content", line.Content)
	assert.Len(t, line.Segments, 1)
	assert.Equal(t, "test", line.Segments[0].Text)
}

func TestSegment(t *testing.T) {
	segment := Segment{
		Start: 5,
		End:   10,
		Type:  LineRemoved,
		Text:  "hello",
	}

	assert.Equal(t, 5, segment.Start)
	assert.Equal(t, 10, segment.End)
	assert.Equal(t, LineRemoved, segment.Type)
	assert.Equal(t, "hello", segment.Text)
}

func TestHunk(t *testing.T) {
	hunk := Hunk{
		Header: "@@ -1,3 +1,3 @@",
		Lines: []DiffLine{
			{Kind: LineContext, Content: "context"},
			{Kind: LineAdded, Content: "added"},
		},
	}

	assert.Equal(t, "@@ -1,3 +1,3 @@", hunk.Header)
	assert.Len(t, hunk.Lines, 2)
	assert.Equal(t, LineContext, hunk.Lines[0].Kind)
	assert.Equal(t, LineAdded, hunk.Lines[1].Kind)
}

func TestParseConfig(t *testing.T) {
	config := ParseConfig{ContextSize: 5}

	// Test WithContextSize option
	opt := WithContextSize(10)
	opt(&config)
	assert.Equal(t, 10, config.ContextSize)

	// Test WithContextSize with negative value (should not change)
	opt = WithContextSize(-1)
	opt(&config)
	assert.Equal(t, 10, config.ContextSize) // Should remain 10
}

func TestSideBySideConfig(t *testing.T) {
	// Test default config
	config := NewSideBySideConfig()
	assert.Equal(t, 160, config.TotalWidth)

	// Test with custom width
	config = NewSideBySideConfig(WithTotalWidth(200))
	assert.Equal(t, 200, config.TotalWidth)

	// Test with invalid width (should not change)
	config = NewSideBySideConfig(WithTotalWidth(-10))
	assert.Equal(t, 160, config.TotalWidth) // Should remain default
}

func TestRenderSideBySideHunk(t *testing.T) {
	hunk := Hunk{
		Header: "@@ -1,3 +1,3 @@",
		Lines: []DiffLine{
			{OldLineNo: 1, NewLineNo: 1, Kind: LineContext, Content: "context line"},
			{OldLineNo: 2, NewLineNo: 0, Kind: LineRemoved, Content: "removed line"},
			{OldLineNo: 0, NewLineNo: 2, Kind: LineAdded, Content: "added line"},
		},
	}

	result := RenderSideBySideHunk("test.txt", hunk)
	assert.NotEmpty(t, result)

	// Should contain the hunk header
	assert.Contains(t, result, "@@ -1,3 +1,3 @@")
}

func TestRenderSideBySideHunk_WithOptions(t *testing.T) {
	hunk := Hunk{
		Header: "@@ -1,2 +1,2 @@",
		Lines: []DiffLine{
			{Kind: LineContext, Content: "test"},
		},
	}

	result := RenderSideBySideHunk("test.txt", hunk, WithTotalWidth(100))
	assert.NotEmpty(t, result)
}

// Test edge cases
func TestParseUnifiedDiff_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		diff string
	}{
		{
			name: "diff with no newline marker",
			diff: `--- a/test.txt
+++ b/test.txt
@@ -1,1 +1,1 @@
-old
+new
\ No newline at end of file`,
		},
		{
			name: "diff without file headers",
			diff: `@@ -1,1 +1,1 @@
-old
+new`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result, err := ParseUnifiedDiff(tt.diff)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}
