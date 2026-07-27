package rag

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitSectionsBreaksAtHeadings(t *testing.T) {
	sections := splitSections("preamble\n\n# One\n\nbody one\n\n## Two\n\nbody two")

	require.Len(t, sections, 3)
	assert.True(t, strings.HasPrefix(sections[0], "preamble"))
	assert.True(t, strings.HasPrefix(sections[1], "# One"))
	assert.True(t, strings.HasPrefix(sections[2], "## Two"))
}

func TestSplitSectionsWithoutHeadingsIsWholeText(t *testing.T) {
	assert.Equal(t, []string{"just\n\nparagraphs"}, splitSections("just\n\nparagraphs"))
}

func TestChunkNeverSpansAHeadingBoundary(t *testing.T) {
	// Both sections are tiny, so pure length-packing would merge them; the
	// heading split must keep the second section's answer in its own chunk.
	text := "# Endpoints\n\n- url one\n\n- url two\n\n# Updating certificates\n\nupload a new metadata file"

	chunks := Chunk(text, 1200)

	require.Len(t, chunks, 2)
	assert.Equal(t, "# Updating certificates\n\nupload a new metadata file", chunks[1])
}

func TestChunkStillPacksAndSlicesWithinASection(t *testing.T) {
	longPara := strings.Repeat("x", 250)
	text := "# S\n\n" + strings.Repeat("a", 90) + "\n\n" + strings.Repeat("b", 90) + "\n\n" + longPara

	chunks := Chunk(text, 100)

	require.Len(t, chunks, 5)
	assert.Equal(t, "# S\n\n"+strings.Repeat("a", 90), chunks[0])
	assert.Equal(t, strings.Repeat("b", 90), chunks[1])
	// The oversized paragraph is hard-sliced rather than kept whole.
	assert.Equal(t, []int{100, 100, 50}, []int{len(chunks[2]), len(chunks[3]), len(chunks[4])})
}

// hardSlice must not cut a multi-byte rune in half, or the chunk becomes
// invalid UTF-8 and the embedding request fails.
func TestChunkSlicesOnRuneBoundaries(t *testing.T) {
	// 60 three-byte runes = 180 bytes, so a 100-byte cap lands mid-rune.
	text := strings.Repeat("→", 60)

	chunks := Chunk(text, 100)

	require.Greater(t, len(chunks), 1)
	for _, chunk := range chunks {
		assert.True(t, isValidUTF8(chunk), "chunk is not valid UTF-8: %q", chunk)
	}
	assert.Equal(t, text, strings.Join(chunks, ""))
}

func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == '�' {
			return false
		}
	}
	return true
}

func TestLooksBinary(t *testing.T) {
	assert.True(t, looksBinary([]byte("PNG\x00\x01\x02")))
	assert.False(t, looksBinary([]byte("plain markdown text")))
}

func TestTsqueryOrTermsOrsUniqueAlphanumericTokens(t *testing.T) {
	q := `SSO login failing: "Invalid SAML Response" SSO 200+`

	assert.Equal(t, "sso | login | failing | invalid | saml | response | 200", tsqueryOrTerms(q))
}

func TestTsqueryOrTermsEmptyQuery(t *testing.T) {
	assert.Equal(t, "", tsqueryOrTerms(""))
	assert.Equal(t, "", tsqueryOrTerms("!!! ???"))
}
