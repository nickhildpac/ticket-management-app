package rag

import (
	"bytes"
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// DefaultMaxChunkChars bounds a single knowledge-base chunk.
const DefaultMaxChunkChars = 1200

var headingRE = regexp.MustCompile(`^#{1,6}\s+\S`)

// splitSections splits markdown into sections at heading lines (any level).
//
// Length-based packing alone glued section bodies onto the tail of unrelated
// content (e.g. an "Updating certificates" answer buried after a wall of
// endpoint URLs), which sank them in both embedding and re-rank scoring.
// Content before the first heading (frontmatter, preamble) is its own section.
func splitSections(text string) []string {
	var sections []string
	var buf []string
	for line := range strings.SplitSeq(text, "\n") {
		if headingRE.MatchString(line) && len(buf) > 0 {
			sections = append(sections, strings.Join(buf, "\n"))
			buf = nil
		}
		buf = append(buf, line)
	}
	if len(buf) > 0 {
		sections = append(sections, strings.Join(buf, "\n"))
	}
	return sections
}

// Chunk performs heading-aware chunking: split into markdown sections, then
// pack each section's paragraphs up to maxChars. Chunks never span a heading
// boundary.
//
// A paragraph longer than maxChars on its own (e.g. an unbroken block or a run
// of markdown image-link URLs) is hard-sliced rather than kept whole —
// otherwise it can exceed the embedding model's input token limit.
func Chunk(text string, maxChars int) []string {
	if maxChars <= 0 {
		maxChars = DefaultMaxChunkChars
	}
	var chunks []string
	for _, section := range splitSections(text) {
		var buf []string
		size := 0
		for para := range strings.SplitSeq(section, "\n\n") {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}
			if size+len(para) > maxChars && len(buf) > 0 {
				chunks = append(chunks, strings.Join(buf, "\n\n"))
				buf, size = nil, 0
			}
			if len(para) > maxChars {
				chunks = append(chunks, hardSlice(para, maxChars)...)
				continue
			}
			buf = append(buf, para)
			size += len(para)
		}
		if len(buf) > 0 {
			chunks = append(chunks, strings.Join(buf, "\n\n"))
		}
	}
	return chunks
}

// hardSlice cuts s into maxChars-sized pieces, nudging each cut forward to the
// next rune boundary so a slice never splits a multi-byte character.
func hardSlice(s string, maxChars int) []string {
	var out []string
	for start := 0; start < len(s); {
		end := min(start+maxChars, len(s))
		for end < len(s) && !utf8.RuneStart(s[end]) {
			end++
		}
		out = append(out, s[start:end])
		start = end
	}
	return out
}

// SkipReason explains why a document contributed no chunks.
type SkipReason string

const (
	// SkipNone means the document was ingested.
	SkipNone SkipReason = ""
	// SkipBinary means the document looked like a binary file.
	SkipBinary SkipReason = "binary"
	// SkipEmpty means the document had no non-whitespace content.
	SkipEmpty SkipReason = "empty"
)

// looksBinary treats a file as binary if it contains a NUL byte in its opening
// bytes — a cheap, reliable heuristic that keeps images/archives/etc. out of
// the store.
func looksBinary(data []byte) bool {
	return bytes.IndexByte(data[:min(len(data), 4096)], 0) >= 0
}

// IngestDocument chunks and stores one document's bytes. It returns the number
// of chunks added and a skip reason, which is SkipNone on success. Shared by
// the CLI directory walk and the HTTP upload endpoint.
func IngestDocument(ctx context.Context, store *VectorStore, source string, raw []byte) (int, SkipReason, error) {
	if looksBinary(raw) {
		return 0, SkipBinary, nil
	}
	// Drop invalid UTF-8 rather than failing, matching the Python
	// decode(errors="ignore") behaviour.
	text := strings.ToValidUTF8(string(raw), "")
	if strings.TrimSpace(text) == "" {
		return 0, SkipEmpty, nil
	}
	count := 0
	for _, chunk := range Chunk(text, DefaultMaxChunkChars) {
		if err := store.Add(ctx, source, chunk); err != nil {
			return count, SkipNone, err
		}
		count++
	}
	return count, SkipNone, nil
}

// IngestPath ingests every readable text file under root — both files sitting
// directly in root and files in any nested subfolder — into the vector store.
// Binary files (images, archives, ...) are skipped so they don't pollute
// retrieval.
func IngestPath(ctx context.Context, store *VectorStore, root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		added, reason, err := IngestDocument(ctx, store, path, raw)
		if err != nil {
			return err
		}
		if reason != SkipNone {
			slog.Debug("skipping file", "reason", reason, "path", path)
		}
		count += added
		return nil
	})
	return count, err
}
