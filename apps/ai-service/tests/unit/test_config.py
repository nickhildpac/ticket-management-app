from __future__ import annotations

import pytest
from pydantic import ValidationError

from app.core.config import Settings


def test_default_settings_work_without_env_file(monkeypatch):
    monkeypatch.delenv("APP_ENV", raising=False)
    monkeypatch.delenv("JWT_SECRET", raising=False)

    settings = Settings(_env_file=None)

    assert settings.app_env == "development"
    assert settings.jwt_secret == "local-dev-only-secret"


def test_local_settings_get_development_secret_fallback(monkeypatch):
    monkeypatch.delenv("JWT_SECRET", raising=False)
    settings = Settings(APP_ENV="local", JWT_SECRET="")

    assert settings.jwt_secret == "local-dev-only-secret"


def test_production_rejects_missing_jwt_secret(monkeypatch):
    monkeypatch.delenv("JWT_SECRET", raising=False)
    with pytest.raises(ValidationError):
        Settings(APP_ENV="production", JWT_SECRET="")


def test_production_rejects_known_weak_jwt_secret():
    with pytest.raises(ValidationError):
        Settings(APP_ENV="production", JWT_SECRET="secret")
