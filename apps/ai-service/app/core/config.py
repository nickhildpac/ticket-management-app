from __future__ import annotations

from functools import lru_cache
from typing import Literal

from pydantic import Field, model_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    app_name: str = Field(default="Ticket Management API (FastAPI)", alias="APP_NAME")
    app_env: str = Field(default="development", alias="APP_ENV")
    api_v1_prefix: str = Field(default="/api/v1", alias="API_V1_PREFIX")

    database_url: str = Field(
        default="postgresql+psycopg://postgres:postgres@localhost:5432/ticket_management",
        alias="DATABASE_URL",
    )

    jwt_secret: str = Field(default="", alias="JWT_SECRET")
    jwt_algorithm: str = Field(default="HS256", alias="JWT_ALGORITHM")
    access_token_expire_minutes: int = Field(default=15, alias="ACCESS_TOKEN_EXPIRE_MINUTES")
    refresh_token_expire_hours: int = Field(default=24, alias="REFRESH_TOKEN_EXPIRE_HOURS")

    refresh_cookie_name: str = Field(default="tapp-refresh_token", alias="REFRESH_COOKIE_NAME")
    refresh_cookie_path: str = Field(default="/", alias="REFRESH_COOKIE_PATH")
    refresh_cookie_secure: bool = Field(default=False, alias="REFRESH_COOKIE_SECURE")
    refresh_cookie_samesite: Literal["lax", "strict", "none"] = Field(
        default="strict", alias="REFRESH_COOKIE_SAMESITE"
    )
    refresh_cookie_domain: str | None = Field(default=None, alias="REFRESH_COOKIE_DOMAIN")

    # Stored as plain text so .env can use comma-separated origins; pydantic-settings
    # JSON-decodes list fields from env before validators run, which breaks that format.
    cors_origins_csv: str = Field(
        default="http://localhost:5173",
        alias="CORS_ORIGINS",
    )

    # ---- AI / RAG / triage settings ------------------------------------
    redis_url: str = Field(default="redis://localhost:6379/0", alias="REDIS_URL")
    ticket_service_url: str = Field(default="http://localhost:8080", alias="TICKET_SERVICE_URL")
    anthropic_api_key: str = Field(default="", alias="ANTHROPIC_API_KEY")
    triage_model: str = Field(default="claude-opus-4-8", alias="TRIAGE_MODEL")
    # Auto-answer only when the model is confident and raised no safety flags.
    auto_answer_confidence_threshold: float = Field(
        default=0.75, alias="AUTO_ANSWER_CONFIDENCE_THRESHOLD"
    )
    embedding_dim: int = Field(default=384, alias="EMBEDDING_DIM")
    # Semantic embeddings via OpenAI when set; otherwise falls back to the
    # offline HashingEmbedder (see app/ai/embeddings.py:build_embedder).
    openai_api_key: str = Field(default="", alias="OPENAI_API_KEY")
    embedding_model: str = Field(default="text-embedding-3-small", alias="EMBEDDING_MODEL")
    # Final passages after hybrid RRF + cross-encoder re-rank.
    rag_top_k: int = Field(default=3, alias="RAG_TOP_K")
    # Candidates fetched from each retrieval lane (semantic / keyword) before RRF.
    rag_candidate_k: int = Field(default=20, alias="RAG_CANDIDATE_K")
    # Fused candidates scored by the cross-encoder before cutting to rag_top_k.
    rag_rerank_pool: int = Field(default=10, alias="RAG_RERANK_POOL")
    # RRF constant (standard value is 60).
    rrf_k: int = Field(default=60, alias="RRF_K")
    # OpenRouter key for Cohere re-rank (POST /api/v1/rerank).
    openrouter_api_key: str = Field(default="", alias="OPENROUTER_API_KEY")
    cross_encoder_model: str = Field(
        default="cohere/rerank-v3.5",
        alias="CROSS_ENCODER_MODEL",
    )
    # Redis Streams consumer group + this replica's consumer name (set per replica
    # when scaling horizontally so pending-entry recovery can tell them apart).
    consumer_group: str = Field(default="ai-triage", alias="CONSUMER_GROUP")
    consumer_name: str = Field(default="worker-1", alias="CONSUMER_NAME")

    # Service account the worker uses to call back into the Go ticket API. The
    # minted JWT must satisfy the ticket-service auth middleware: an existing
    # agent/admin user id as `sub`, matching issuer/audience, signed with the
    # shared jwt_secret. See docs/adr/0002-service-topology.md.
    # Defaults match the seeded AI service account (migration 000013). It is an
    # admin because CanCommentOnTicket only lets admins comment on tickets they
    # aren't assigned to, and the worker comments on arbitrary tickets.
    service_account_id: str = Field(
        default="00000000-0000-4000-8000-0000000000a1", alias="AI_SERVICE_ACCOUNT_ID"
    )
    service_account_role: str = Field(default="admin", alias="AI_SERVICE_ACCOUNT_ROLE")
    jwt_issuer: str = Field(default="example.com", alias="JWT_ISSUER")
    jwt_audience: str = Field(default="example.com", alias="JWT_AUDIENCE")

    @property
    def cors_origins(self) -> list[str]:
        return [item.strip() for item in self.cors_origins_csv.split(",") if item.strip()]

    @model_validator(mode="after")
    def validate_jwt_secret(self) -> Settings:
        env = self.app_env.strip().lower()
        if env in {"", "local", "development", "test"}:
            if not self.jwt_secret:
                self.jwt_secret = "local-dev-only-secret"
            return self

        if self.jwt_secret.strip() in {"", "secret", "changeme", "change-me", "local-dev-only-secret"}:
            raise ValueError("JWT_SECRET is required and must not use a known weak value outside local/development/test")
        return self


@lru_cache
def get_settings() -> Settings:
    return Settings()
