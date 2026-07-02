from __future__ import annotations

import json
from pathlib import Path

from app.main import app

if __name__ == "__main__":
    openapi = app.openapi()
    out_path = Path(__file__).resolve().parents[1] / "openapi.json"
    out_path.write_text(json.dumps(openapi, indent=2), encoding="utf-8")
    print(f"wrote {out_path}")
