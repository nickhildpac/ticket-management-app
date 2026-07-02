from __future__ import annotations

import argparse
import logging
from pathlib import Path

from sqlalchemy import create_engine

from app.ai.embeddings import HashingEmbedder
from app.ai.vectorstore import VectorStore
from app.core.config import get_settings

logger = logging.getLogger(__name__)


def _chunk(text: str, max_chars: int = 1200) -> list[str]:
    """Naive paragraph-based chunking. Swap for a smarter splitter as needed."""
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
        buf.append(para)
        size += len(para)
    if buf:
        chunks.append("\n\n".join(buf))
    return chunks


def ingest_path(store: VectorStore, root: Path) -> int:
    """Ingest every markdown/text file under root into the vector store."""
    count = 0
    for path in sorted(root.rglob("*")):
        if path.suffix.lower() not in {".md", ".txt"} or not path.is_file():
            continue
        text = path.read_text(encoding="utf-8", errors="ignore")
        for chunk in _chunk(text):
            store.add(source=str(path), content=chunk)
            count += 1
    return count


def main() -> None:
    logging.basicConfig(level=logging.INFO)
    parser = argparse.ArgumentParser(description="Ingest a knowledge base into the vector store.")
    parser.add_argument(
        "path",
        nargs="?",
        default="knowledge",
        help="Directory of .md/.txt files to ingest (default: ./knowledge).",
    )
    args = parser.parse_args()

    settings = get_settings()
    engine = create_engine(settings.database_url, pool_pre_ping=True)
    store = VectorStore(engine, HashingEmbedder(dim=settings.embedding_dim))

    root = Path(args.path)
    if not root.exists():
        raise SystemExit(f"knowledge path not found: {root}")
    n = ingest_path(store, root)
    logger.info("ingested %d chunk(s) from %s", n, root)


if __name__ == "__main__":
    main()
