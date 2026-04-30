from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

pytestmark = pytest.mark.integration


def test_register_login_refresh_logout_flow(client: TestClient):
    register_payload = {
        "first_name": "Jane",
        "last_name": "Doe",
        "email": "jane@example.com",
        "password": "password123",
        "skills": ["incident-management"],
    }

    register_res = client.post("/api/v1/user", json=register_payload)
    assert register_res.status_code == 201

    login_res = client.post(
        "/api/v1/login",
        json={"email": "jane@example.com", "password": "password123"},
    )
    assert login_res.status_code == 200
    assert "access_token" in login_res.json()
    assert "user" in login_res.json()

    refresh_res = client.get("/api/v1/refresh")
    assert refresh_res.status_code == 200
    assert "access_token" in refresh_res.json()

    logout_res = client.get("/api/v1/logout")
    assert logout_res.status_code == 202
