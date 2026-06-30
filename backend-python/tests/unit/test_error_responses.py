from __future__ import annotations

from fastapi import FastAPI
from fastapi.testclient import TestClient
from pydantic import BaseModel

from app.api.errors import register_exception_handlers
from app.core.exceptions import BadRequestError


class Payload(BaseModel):
    name: str


def test_app_error_uses_structured_envelope():
    app = FastAPI()
    register_exception_handlers(app)

    @app.get("/bad")
    def bad():
        raise BadRequestError("invalid request")

    res = TestClient(app).get("/bad")

    assert res.status_code == 400
    assert res.json() == {
        "code": "bad_request",
        "message": "invalid request",
        "details": None,
    }


def test_validation_error_includes_field_details():
    app = FastAPI()
    register_exception_handlers(app)

    @app.post("/payload")
    def payload(_: Payload):
        return {"ok": True}

    res = TestClient(app).post("/payload", json={})

    assert res.status_code == 400
    body = res.json()
    assert body["code"] == "validation_error"
    assert body["message"] == "request validation failed"
    assert body["details"][0]["field"] == "body.name"
    assert body["details"][0]["message"]


def test_generic_error_hides_internal_message():
    app = FastAPI()
    register_exception_handlers(app)

    @app.get("/boom")
    def boom():
        raise RuntimeError("database password leaked")

    res = TestClient(app, raise_server_exceptions=False).get("/boom", headers={"x-request-id": "req-1"})

    assert res.status_code == 500
    assert res.json() == {
        "code": "internal_server_error",
        "message": "internal server error",
        "details": None,
    }
