from __future__ import annotations

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import select
from sqlalchemy.orm import Session

from app.enums.role import UserRole
from app.models.user import User

pytestmark = pytest.mark.integration


def login(client: TestClient, email: str, password: str) -> str:
    res = client.post("/api/v1/login", json={"email": email, "password": password})
    assert res.status_code == 200
    return res.json()["access_token"]


def auth_headers(token: str) -> dict[str, str]:
    return {"Authorization": f"Bearer {token}"}


def test_role_scoped_ticket_listing(client: TestClient, db_session: Session):
    requester = {
        "first_name": "Requester",
        "last_name": "User",
        "email": "requester@example.com",
        "password": "password123",
        "skills": ["incident-management"],
    }
    agent = {
        "first_name": "Agent",
        "last_name": "User",
        "email": "agent@example.com",
        "password": "password123",
        "skills": ["incident-management"],
    }
    admin = {
        "first_name": "Admin",
        "last_name": "User",
        "email": "admin@example.com",
        "password": "password123",
        "skills": ["incident-management"],
    }

    assert client.post("/api/v1/user", json=requester).status_code == 201
    assert client.post("/api/v1/user", json=agent).status_code == 201
    assert client.post("/api/v1/user", json=admin).status_code == 201

    admin_user = db_session.scalar(select(User).where(User.email == "admin@example.com"))
    assert admin_user is not None
    admin_user.role = UserRole.ADMIN
    db_session.commit()

    requester_token = login(client, requester["email"], requester["password"])
    create_res = client.post(
        "/api/v1/ticket",
        json={"title": "Need help", "description": "Service down", "priority": "high", "skills": ["incident-management"]},
        headers=auth_headers(requester_token),
    )
    assert create_res.status_code == 202
    ticket_id = create_res.json()["id"]

    admin_token = login(client, admin["email"], admin["password"])
    agent_user = db_session.scalar(select(User).where(User.email == "agent@example.com"))
    assert agent_user is not None
    patch_res = client.patch(
        f"/api/v1/ticket/{ticket_id}",
        json={"assigned_to": [str(agent_user.id)]},
        headers=auth_headers(admin_token),
    )
    assert patch_res.status_code == 200

    agent_token = login(client, agent["email"], agent["password"])
    assigned_res = client.get("/api/v1/ticket/assigned", headers=auth_headers(agent_token))
    assert assigned_res.status_code == 200
    assert len(assigned_res.json()) == 1

    requester_all = client.get("/api/v1/ticket/all", headers=auth_headers(requester_token))
    assert requester_all.status_code == 200
    assert len(requester_all.json()) == 1
