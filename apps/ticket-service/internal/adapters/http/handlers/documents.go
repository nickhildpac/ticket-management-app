package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

// maxUploadBytes bounds the request body the proxy will accept, protecting both
// this service and ai-service from oversized uploads.
const maxUploadBytes = 25 << 20 // 25 MiB

// ingestProxyClient carries a generous timeout: ingestion embeds every chunk, so
// a multi-file upload can take a while.
var ingestProxyClient = &http.Client{Timeout: 60 * time.Second}

// IngestDocuments transparently proxies an admin's multipart document upload to
// the ai-service ingest endpoint, which chunks and embeds the files into the
// knowledge base. It is mounted behind AdminRequired: ai-service's own auth does
// not check role, so admin enforcement lives here. The caller's access token is
// forwarded verbatim — it already satisfies ai-service's shared-secret JWT check.
//
// @Summary		Ingest knowledge-base documents
// @Description	Admin-only. Uploads documents (multipart field "files") to the AI service for KB ingestion.
// @Tags			Admin
// @Accept			multipart/form-data
// @Produce		json
// @Success		200	{object}	object{files=[]object,total_chunks=int}
// @Failure		401	{object}	util.ErrorBody
// @Failure		403	{object}	util.ErrorBody
// @Failure		502	{object}	util.ErrorBody
// @Router			/admin/documents [post]
func (h *Handler) IngestDocuments(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, maxUploadBytes)

	upstream, err := http.NewRequestWithContext(
		r.Context(), http.MethodPost, h.config.AIServiceURL+"/api/v1/ingest", body,
	)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	// Preserve the multipart boundary and forward the admin's bearer token.
	upstream.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	upstream.Header.Set("Authorization", r.Header.Get("Authorization"))

	resp, err := ingestProxyClient.Do(upstream)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadGateway, errors.New("ai-service unreachable"))
		return
	}
	defer resp.Body.Close()

	// Stream ai-service's JSON response (success or error) back to the caller unchanged.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
