from __future__ import annotations

from functools import lru_cache

from fastapi import APIRouter, Depends, File, UploadFile
from pydantic import BaseModel
from sqlalchemy import create_engine

from app.ai.embeddings import build_embedder
from app.ai.ingest import ingest_document
from app.ai.vectorstore import VectorStore
from app.api.triage import require_auth
from app.core.config import get_settings


@lru_cache
def _store() -> VectorStore:
    """Process-singleton store. Constructing it opens a SQLAlchemy engine (which
    owns a connection pool), so it must not be rebuilt per request. Uses the same
    embedder as the triage retriever so ingested vectors are searchable."""
    settings = get_settings()
    engine = create_engine(settings.database_url, pool_pre_ping=True)
    return VectorStore(engine, build_embedder(settings))


class IngestedFile(BaseModel):
    source: str
    chunks: int
    skipped: bool = False
    reason: str | None = None


class IngestResponse(BaseModel):
    files: list[IngestedFile]
    total_chunks: int


router = APIRouter(prefix="/ingest", tags=["ingest"], dependencies=[Depends(require_auth)])


@router.post("", response_model=IngestResponse)
async def ingest(files: list[UploadFile] = File(...)) -> IngestResponse:
    """Ingest one or more uploaded documents into the vector store.

    Multipart upload; each file is chunked and embedded. Binary and empty files
    are reported as skipped rather than failing the whole request. Auth is the
    same shared-secret JWT as the triage endpoint, so only callers holding a
    ticket-service token can add to the knowledge base."""
    store = _store()
    results: list[IngestedFile] = []
    total = 0
    for upload in files:
        raw = await upload.read()
        source = upload.filename or "upload"
        added, reason = ingest_document(store, source, raw)
        results.append(
            IngestedFile(source=source, chunks=added, skipped=added == 0, reason=reason)
        )
        total += added
    return IngestResponse(files=results, total_chunks=total)
