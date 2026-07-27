package httpapi

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/nickhildpac/ticket-management-ai-service/internal/rag"
)

// maxUploadBytes bounds the multipart body. It matches the cap the
// ticket-service ingest proxy already applies, so this is a defence in depth
// for direct callers.
const maxUploadBytes = 25 << 20 // 25 MiB

// ingestedFile is one file's outcome. reason is null on success, or
// "binary"/"empty" when nothing was stored.
type ingestedFile struct {
	Source  string  `json:"source"`
	Chunks  int     `json:"chunks"`
	Skipped bool    `json:"skipped"`
	Reason  *string `json:"reason"`
}

// ingestResponse is the payload the web admin UI renders (see
// apps/web/src/features/admin/queries.ts).
type ingestResponse struct {
	Files       []ingestedFile `json:"files"`
	TotalChunks int            `json:"total_chunks"`
}

// ingestHandler ingests one or more uploaded documents into the vector store.
//
// Multipart upload under the field name "files"; each file is chunked and
// embedded. Binary and empty files are reported as skipped rather than failing
// the whole request. Auth is the same shared-secret JWT as the triage endpoint,
// so only callers holding a ticket-service token can add to the knowledge base.
func ingestHandler(store *rag.VectorStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
		reader, err := r.MultipartReader()
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation_error",
				"expected a multipart/form-data upload", err.Error())
			return
		}

		resp := ingestResponse{Files: []ingestedFile{}}
		for {
			part, err := reader.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				writeError(w, http.StatusBadRequest, "validation_error",
					"failed to read upload", err.Error())
				return
			}
			if part.FormName() != "files" {
				_ = part.Close()
				continue
			}

			file, err := ingestPart(r, store, part)
			_ = part.Close()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error",
					"failed to ingest document", nil)
				return
			}
			resp.Files = append(resp.Files, file)
			resp.TotalChunks += file.Chunks
		}

		if len(resp.Files) == 0 {
			writeError(w, http.StatusBadRequest, "validation_error",
				"request validation failed",
				[]map[string]string{{
					"field": "files", "message": "field required", "type": "missing",
				}})
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// ingestPart chunks and embeds a single uploaded file.
func ingestPart(r *http.Request, store *rag.VectorStore, part *multipart.Part) (ingestedFile, error) {
	source := part.FileName()
	if source == "" {
		source = "upload"
	}
	raw, err := io.ReadAll(part)
	if err != nil {
		return ingestedFile{}, err
	}

	added, reason, err := rag.IngestDocument(r.Context(), store, source, raw)
	if err != nil {
		return ingestedFile{}, err
	}

	out := ingestedFile{Source: source, Chunks: added, Skipped: added == 0}
	if reason != rag.SkipNone {
		text := string(reason)
		out.Reason = &text
	}
	return out, nil
}
