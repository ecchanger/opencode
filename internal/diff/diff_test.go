package diff

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/tui/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUnifiedDiff(t *testing.T) {
	t.Run("parses simple unified diff", func(t *testing.T) {
		diffText := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+line 2 modified
 line 3`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)

		assert.Equal(t, "test.txt", result.OldFile)
		assert.Equal(t, "test.txt", result.NewFile)
		assert.Len(t, result.Hunks, 1)

		hunk := result.Hunks[0]
		assert.Equal(t, "@@ -1,3 +1,3 @@", hunk.Header)
		assert.Len(t, hunk.Lines, 3)

		// Check first line (context)
		assert.Equal(t, LineContext, hunk.Lines[0].Kind)
		assert.Equal(t, "line 1", hunk.Lines[0].Content)
		assert.Equal(t, 1, hunk.Lines[0].OldLineNo)
		assert.Equal(t, 1, hunk.Lines[0].NewLineNo)

		// Check second line (removed)
		assert.Equal(t, LineRemoved, hunk.Lines[1].Kind)
		assert.Equal(t, "line 2", hunk.Lines[1].Content)
		assert.Equal(t, 2, hunk.Lines[1].OldLineNo)
		assert.Equal(t, 0, hunk.Lines[1].NewLineNo)

		// Check third line (added)
		assert.Equal(t, LineAdded, hunk.Lines[2].Kind)
		assert.Equal(t, "line 2 modified", hunk.Lines[2].Content)
		assert.Equal(t, 0, hunk.Lines[2].OldLineNo)
		assert.Equal(t, 2, hunk.Lines[2].NewLineNo)
	})

	t.Run("parses multiple hunks", func(t *testing.T) {
		diffText := `--- a/test.txt
+++ b/test.txt
@@ -1,2 +1,2 @@
-old line 1
+new line 1
 unchanged line
@@ -10,2 +10,2 @@
 another unchanged line
-old line 2
+new line 2`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)

		assert.Len(t, result.Hunks, 2)

		// First hunk
		assert.Equal(t, "@@ -1,2 +1,2 @@", result.Hunks[0].Header)
		assert.Len(t, result.Hunks[0].Lines, 2)

		// Second hunk
		assert.Equal(t, "@@ -10,2 +10,2 @@", result.Hunks[1].Header)
		assert.Len(t, result.Hunks[1].Lines, 2)
	})

	t.Run("handles empty lines", func(t *testing.T) {
		diffText := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 
-removed line
+added line`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)

		hunk := result.Hunks[0]
		assert.Len(t, hunk.Lines, 3)

		// Empty line should be context
		assert.Equal(t, LineContext, hunk.Lines[0].Kind)
		assert.Equal(t, "", hunk.Lines[0].Content)
	})

	t.Run("handles no newline at end of file", func(t *testing.T) {
		diffText := `--- a/test.txt
+++ b/test.txt
@@ -1 +1 @@
-old content
\ No newline at end of file
+new content
\ No newline at end of file`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)

		hunk := result.Hunks[0]
		assert.Len(t, hunk.Lines, 2)
		assert.Equal(t, "old content", hunk.Lines[0].Content)
		assert.Equal(t, "new content", hunk.Lines[1].Content)
	})

	t.Run("handles addition only diff", func(t *testing.T) {
		diffText := `--- /dev/null
+++ b/new_file.txt
@@ -0,0 +1,2 @@
+first line
+second line`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)

		assert.Equal(t, "/dev/null", result.OldFile)
		assert.Equal(t, "new_file.txt", result.NewFile)

		hunk := result.Hunks[0]
		assert.Len(t, hunk.Lines, 2)
		assert.Equal(t, LineAdded, hunk.Lines[0].Kind)
		assert.Equal(t, LineAdded, hunk.Lines[1].Kind)
	})

	t.Run("handles deletion only diff", func(t *testing.T) {
		diffText := `--- a/old_file.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-first line
-second line`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)

		assert.Equal(t, "old_file.txt", result.OldFile)
		assert.Equal(t, "/dev/null", result.NewFile)

		hunk := result.Hunks[0]
		assert.Len(t, hunk.Lines, 2)
		assert.Equal(t, LineRemoved, hunk.Lines[0].Kind)
		assert.Equal(t, LineRemoved, hunk.Lines[1].Kind)
	})
}

func TestHighlightIntralineChanges(t *testing.T) {
	t.Run("highlights character-level changes", func(t *testing.T) {
		hunk := &Hunk{
			Lines: []DiffLine{
				{
					Kind:    LineRemoved,
					Content: "Hello world",
				},
				{
					Kind:    LineAdded,
					Content: "Hello universe",
				},
			},
		}

		HighlightIntralineChanges(hunk)

		// Both lines should have segments
		assert.NotEmpty(t, hunk.Lines[0].Segments)
		assert.NotEmpty(t, hunk.Lines[1].Segments)
	})

	t.Run("preserves non-paired lines", func(t *testing.T) {
		hunk := &Hunk{
			Lines: []DiffLine{
				{
					Kind:    LineContext,
					Content: "unchanged line",
				},
				{
					Kind:    LineAdded,
					Content: "added line",
				},
			},
		}

		HighlightIntralineChanges(hunk)

		// Context line should remain unchanged
		assert.Empty(t, hunk.Lines[0].Segments)
		// Added line without corresponding removed line should remain unchanged
		assert.Empty(t, hunk.Lines[1].Segments)
	})
}

func TestPairLines(t *testing.T) {
	lines := []DiffLine{
		{Kind: LineContext, Content: "context line"},
		{Kind: LineRemoved, Content: "removed line"},
		{Kind: LineAdded, Content: "added line"},
		{Kind: LineRemoved, Content: "another removed"},
		{Kind: LineContext, Content: "another context"},
	}

	pairs := pairLines(lines)

	assert.Len(t, pairs, 4)

	// First pair: context line (appears on both sides)
	assert.NotNil(t, pairs[0].left)
	assert.NotNil(t, pairs[0].right)
	assert.Equal(t, pairs[0].left, pairs[0].right)

	// Second pair: removed + added
	assert.NotNil(t, pairs[1].left)
	assert.NotNil(t, pairs[1].right)
	assert.Equal(t, LineRemoved, pairs[1].left.Kind)
	assert.Equal(t, LineAdded, pairs[1].right.Kind)

	// Third pair: removed only
	assert.NotNil(t, pairs[2].left)
	assert.Nil(t, pairs[2].right)

	// Fourth pair: context line
	assert.NotNil(t, pairs[3].left)
	assert.NotNil(t, pairs[3].right)
	assert.Equal(t, pairs[3].left, pairs[3].right)
}

func TestGenerateDiff(t *testing.T) {
	t.Run("generates diff for simple change", func(t *testing.T) {
		before := "line 1\nline 2\nline 3"
		after := "line 1\nline 2 modified\nline 3"

		diffText, added, removed := GenerateDiff(before, after, "test.txt")

		assert.NotEmpty(t, diffText)
		assert.Equal(t, 1, added)
		assert.Equal(t, 1, removed)

		// Should contain diff markers
		assert.Contains(t, diffText, "---")
		assert.Contains(t, diffText, "+++")
		assert.Contains(t, diffText, "@@")
		assert.Contains(t, diffText, "-line 2")
		assert.Contains(t, diffText, "+line 2 modified")
	})

	t.Run("generates diff for addition", func(t *testing.T) {
		before := "line 1\nline 2"
		after := "line 1\nline 2\nline 3"

		diffText, added, removed := GenerateDiff(before, after, "test.txt")

		assert.NotEmpty(t, diffText)
		assert.Equal(t, 1, added)
		assert.Equal(t, 0, removed)
		assert.Contains(t, diffText, "+line 3")
	})

	t.Run("generates diff for removal", func(t *testing.T) {
		before := "line 1\nline 2\nline 3"
		after := "line 1\nline 3"

		diffText, added, removed := GenerateDiff(before, after, "test.txt")

		assert.NotEmpty(t, diffText)
		assert.Equal(t, 0, added)
		assert.Equal(t, 1, removed)
		assert.Contains(t, diffText, "-line 2")
	})

	t.Run("handles identical content", func(t *testing.T) {
		content := "line 1\nline 2\nline 3"

		diffText, added, removed := GenerateDiff(content, content, "test.txt")

		assert.Empty(t, diffText)
		assert.Equal(t, 0, added)
		assert.Equal(t, 0, removed)
	})
}

func TestFormatDiff(t *testing.T) {
	t.Run("formats unified diff", func(t *testing.T) {
		diffText := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+line 2 modified
 line 3`

		result, err := FormatDiff(diffText)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("handles invalid diff", func(t *testing.T) {
		invalidDiff := "not a valid diff"

		result, err := FormatDiff(invalidDiff)
		require.NoError(t, err)
		// Should return empty result for invalid diff
		assert.Empty(t, result)
	})
}

func TestSideBySideConfig(t *testing.T) {
	t.Run("creates config with defaults", func(t *testing.T) {
		config := NewSideBySideConfig()
		assert.Equal(t, 160, config.TotalWidth)
	})

	t.Run("applies custom width", func(t *testing.T) {
		config := NewSideBySideConfig(WithTotalWidth(120))
		assert.Equal(t, 120, config.TotalWidth)
	})

	t.Run("ignores invalid width", func(t *testing.T) {
		config := NewSideBySideConfig(WithTotalWidth(-10))
		assert.Equal(t, 160, config.TotalWidth) // Should keep default
	})
}

func TestParseConfig(t *testing.T) {
	t.Run("applies context size", func(t *testing.T) {
		config := &ParseConfig{}
		WithContextSize(5)(config)
		assert.Equal(t, 5, config.ContextSize)
	})

	t.Run("ignores negative context size", func(t *testing.T) {
		config := &ParseConfig{ContextSize: 3}
		WithContextSize(-1)(config)
		assert.Equal(t, 3, config.ContextSize) // Should remain unchanged
	})
}

func TestRenderSideBySideHunk(t *testing.T) {
	hunk := Hunk{
		Header: "@@ -1,3 +1,3 @@",
		Lines: []DiffLine{
			{
				Kind:      LineContext,
				Content:   "context line",
				OldLineNo: 1,
				NewLineNo: 1,
			},
			{
				Kind:      LineRemoved,
				Content:   "removed line",
				OldLineNo: 2,
				NewLineNo: 0,
			},
			{
				Kind:      LineAdded,
				Content:   "added line",
				OldLineNo: 0,
				NewLineNo: 2,
			},
		},
	}

	result := RenderSideBySideHunk("test.txt", hunk)
	assert.NotEmpty(t, result)

	// Should contain line numbers and content
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
	assert.Contains(t, result, "context line")
	assert.Contains(t, result, "removed line")
	assert.Contains(t, result, "added line")
}

func TestGetColor(t *testing.T) {
	adaptiveColor := lipgloss.AdaptiveColor{
		Light: "#000000",
		Dark:  "#ffffff",
	}

	color := getColor(adaptiveColor)
	// Should return one of the colors (light or dark)
	assert.True(t, color == "#000000" || color == "#ffffff")
}

func TestCreateStyles(t *testing.T) {
	// Mock theme for testing
	mockTheme := newMockTheme()

	removedStyle, addedStyle, contextStyle, lineNumberStyle := createStyles(mockTheme)

	// Styles should be created (non-nil)
	assert.NotNil(t, removedStyle)
	assert.NotNil(t, addedStyle)
	assert.NotNil(t, contextStyle)
	assert.NotNil(t, lineNumberStyle)
}

// Mock theme for testing - simplified implementation using embedded BaseTheme
type mockTheme struct {
	theme.BaseTheme
}

func newMockTheme() *mockTheme {
	return &mockTheme{
		BaseTheme: theme.BaseTheme{
			PrimaryColor:   lipgloss.AdaptiveColor{Light: "#0000ff", Dark: "#0000ff"},
			SecondaryColor: lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"},
			AccentColor:    lipgloss.AdaptiveColor{Light: "#ff00ff", Dark: "#ff00ff"},

			ErrorColor:   lipgloss.AdaptiveColor{Light: "#ff0000", Dark: "#ff0000"},
			WarningColor: lipgloss.AdaptiveColor{Light: "#ffaa00", Dark: "#ffaa00"},
			SuccessColor: lipgloss.AdaptiveColor{Light: "#00ff00", Dark: "#00ff00"},
			InfoColor:    lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},

			TextColor:           lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			TextMutedColor:      lipgloss.AdaptiveColor{Light: "#666666", Dark: "#aaaaaa"},
			TextEmphasizedColor: lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},

			BackgroundColor:          lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"},
			BackgroundSecondaryColor: lipgloss.AdaptiveColor{Light: "#f0f0f0", Dark: "#111111"},
			BackgroundDarkerColor:    lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: "#222222"},

			BorderNormalColor:  lipgloss.AdaptiveColor{Light: "#cccccc", Dark: "#444444"},
			BorderFocusedColor: lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},
			BorderDimColor:     lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#333333"},

			DiffAddedColor:               lipgloss.AdaptiveColor{Light: "#00ff00", Dark: "#00ff00"},
			DiffRemovedColor:             lipgloss.AdaptiveColor{Light: "#ff0000", Dark: "#ff0000"},
			DiffContextColor:             lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			DiffHunkHeaderColor:          lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},
			DiffHighlightAddedColor:      lipgloss.AdaptiveColor{Light: "#88ff88", Dark: "#88ff88"},
			DiffHighlightRemovedColor:    lipgloss.AdaptiveColor{Light: "#ff8888", Dark: "#ff8888"},
			DiffAddedBgColor:             lipgloss.AdaptiveColor{Light: "#eeffee", Dark: "#001100"},
			DiffRemovedBgColor:           lipgloss.AdaptiveColor{Light: "#ffeeee", Dark: "#110000"},
			DiffContextBgColor:           lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"},
			DiffLineNumberColor:          lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"},
			DiffAddedLineNumberBgColor:   lipgloss.AdaptiveColor{Light: "#ccffcc", Dark: "#002200"},
			DiffRemovedLineNumberBgColor: lipgloss.AdaptiveColor{Light: "#ffcccc", Dark: "#220000"},

			// Markdown colors
			MarkdownTextColor:            lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			MarkdownHeadingColor:         lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},
			MarkdownLinkColor:            lipgloss.AdaptiveColor{Light: "#0000ff", Dark: "#0000ff"},
			MarkdownLinkTextColor:        lipgloss.AdaptiveColor{Light: "#0000ff", Dark: "#0000ff"},
			MarkdownCodeColor:            lipgloss.AdaptiveColor{Light: "#ff6600", Dark: "#ff6600"},
			MarkdownBlockQuoteColor:      lipgloss.AdaptiveColor{Light: "#666666", Dark: "#aaaaaa"},
			MarkdownEmphColor:            lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			MarkdownStrongColor:          lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			MarkdownHorizontalRuleColor:  lipgloss.AdaptiveColor{Light: "#cccccc", Dark: "#444444"},
			MarkdownListItemColor:        lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			MarkdownListEnumerationColor: lipgloss.AdaptiveColor{Light: "#666666", Dark: "#aaaaaa"},
			MarkdownImageColor:           lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},
			MarkdownImageTextColor:       lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},
			MarkdownCodeBlockColor:       lipgloss.AdaptiveColor{Light: "#ff6600", Dark: "#ff6600"},

			// Syntax highlighting colors
			SyntaxCommentColor:     lipgloss.AdaptiveColor{Light: "#666666", Dark: "#aaaaaa"},
			SyntaxKeywordColor:     lipgloss.AdaptiveColor{Light: "#0000ff", Dark: "#0000ff"},
			SyntaxFunctionColor:    lipgloss.AdaptiveColor{Light: "#0088ff", Dark: "#0088ff"},
			SyntaxVariableColor:    lipgloss.AdaptiveColor{Light: "#ff6600", Dark: "#ff6600"},
			SyntaxStringColor:      lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#00aa00"},
			SyntaxNumberColor:      lipgloss.AdaptiveColor{Light: "#aa0000", Dark: "#aa0000"},
			SyntaxTypeColor:        lipgloss.AdaptiveColor{Light: "#aa00aa", Dark: "#aa00aa"},
			SyntaxOperatorColor:    lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
			SyntaxPunctuationColor: lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
		},
	}
}

func TestLipglossToHex(t *testing.T) {
	testCases := []struct {
		name     string
		input    lipgloss.Color
		expected string
	}{
		{
			name:     "hex color",
			input:    lipgloss.Color("#ff0000"),
			expected: "#ff0000",
		},
		{
			name:     "named color",
			input:    lipgloss.Color("red"),
			expected: "red",
		},
		{
			name:     "ansi color",
			input:    lipgloss.Color("1"),
			expected: "1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lipglossToHex(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty diff", func(t *testing.T) {
		result, err := ParseUnifiedDiff("")
		require.NoError(t, err)
		assert.Empty(t, result.Hunks)
	})

	t.Run("malformed hunk header", func(t *testing.T) {
		diffText := `--- a/test.txt
+++ b/test.txt
invalid hunk header
 some content`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)
		// Should handle gracefully
		assert.Empty(t, result.Hunks)
	})

	t.Run("diff without file headers", func(t *testing.T) {
		diffText := `@@ -1,2 +1,2 @@
-old line
+new line`

		result, err := ParseUnifiedDiff(diffText)
		require.NoError(t, err)
		assert.Len(t, result.Hunks, 1)
		assert.Empty(t, result.OldFile)
		assert.Empty(t, result.NewFile)
	})
}
