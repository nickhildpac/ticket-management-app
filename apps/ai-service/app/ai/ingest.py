from __future__ import annotations

import argparse
import logging
from pathlib import Path

from sqlalchemy import create_engine

from app.ai.embeddings import build_embedder
from app.ai.vectorstore import VectorStore
from app.core.config import get_settings

logger = logging.getLogger(__name__)


def _chunk(text: str, max_chars: int = 1200) -> list[str]:
    """Naive paragraph-based chunking. Swap for a smarter splitter as needed.

    A paragraph longer than max_chars on its own (e.g. an unbroken block or a
    run of markdown image-link URLs) is hard-sliced rather than kept whole —
    otherwise it can exceed the embedding model's input token limit."""
    chunks: list[str] = []
    buf: list[str] = []
    size = 0
    for para in text.split("\n\n"):
        para = para.strip()
        if not para:
            continue
        if size + len(para) > max_chars and buf:
            chunks.append("\n\n".join(buf))
            buf, size = [], 0
        if len(para) > max_chars:
            for i in range(0, len(para), max_chars):
                chunks.append(para[i : i + max_chars])
            continue
        buf.append(para)
        size += len(para)
    if buf:
        chunks.append("\n\n".join(buf))
    return chunks


def _looks_binary(data: bytes) -> bool:
    """Treat a file as binary if it contains a NUL byte in its opening bytes —
    a cheap, reliable heuristic that keeps images/archives/etc. out of the store."""
    return b"\x00" in data[:4096]


def ingest_document(store: VectorStore, source: str, raw: bytes) -> tuple[int, str | None]:
    """Chunk and store one document's bytes. Returns (chunks_added, skip_reason);
    skip_reason is None on success, or "binary"/"empty" when nothing was stored.
    Shared by the CLI directory walk and the HTTP upload endpoint."""
    if _looks_binary(raw):
        return 0, "binary"
    text = raw.decode("utf-8", errors="ignore")
    if not text.strip():
        return 0, "empty"
    count = 0
    for chunk in _chunk(text):
        store.add(source=source, content=chunk)
        count += 1
    return count, None


def ingest_path(store: VectorStore, root: Path) -> int:
    """Ingest every readable text file under root — both files sitting directly in
    root and files in any nested subfolder — into the vector store. Binary files
    (images, archives, ...) are skipped so they don't pollute retrieval."""
    count = 0
    for path in sorted(root.rglob("*")):
        if not path.is_file():
            continue
        added, reason = ingest_document(store, str(path), path.read_bytes())
        if reason:
            logger.debug("skipping %s file: %s", reason, path)
        count += added
    return count


def main() -> None:
    logging.basicConfig(level=logging.INFO)
    parser = argparse.ArgumentParser(description="Ingest a knowledge base into the vector store.")
    parser.add_argument(
        "path",
        nargs="?",
        default="knowledge",
        help="Directory of text files to ingest, recursively (default: ./knowledge).",
    )
    args = parser.parse_args()

    settings = get_settings()
    engine = create_engine(settings.database_url, pool_pre_ping=True)
    store = VectorStore(engine, build_embedder(settings))

    root = Path(args.path)
    if not root.exists():
        raise SystemExit(f"knowledge path not found: {root}")
    n = ingest_path(store, root)
    logger.info("ingested %d chunk(s) from %s", n, root)


if __name__ == "__main__":
    main()
