from unittest.mock import MagicMock
from uuid import uuid4

import pytest

from app.core.exceptions import BadRequestError, ForbiddenError
from app.enums.role import UserRole
from app.enums.ticket_state import TicketState
from app.models.association import TicketAssignee
from app.models.ticket import Ticket
from app.models.user import User
from app.schemas.ticket import TicketListQuery, UpdateTicketRequest
from app.services.ticket import TicketService


class RepoStub:
    def __init__(self):
        self.list_tickets_called = False
        self.deleted = False
        self.ticket = None

    def list_tickets(self, **kwargs):  # noqa: ANN003
        self.list_tickets_called = True
        return []

    def count_tickets(self, **kwargs):  # noqa: ANN003
        _ = kwargs
        return 0

    def get_by_id(self, ticket_id):  # noqa: ANN001
        _ = ticket_id
        return self.ticket

    def get_by_number(self, ticket_number):  # noqa: ANN001
        _ = ticket_number
        return self.ticket

    def delete(self, ticket):  # noqa: ANN001
        _ = ticket
        self.deleted = True

    def set_skills(self, ticket, skills):  # noqa: ANN001
        _ = (ticket, skills)

    def set_assignees(self, ticket, assignee_ids):  # noqa: ANN001
        ticket.assignee_links = [
            TicketAssignee(ticket_id=ticket.id, user_id=user_id) for user_id in assignee_ids
        ]


class UserRepoStub:
    def get_by_ids(self, user_ids):  # noqa: ANN001
        return [
            User(
                id=user_id,
                email=f"{user_id}@example.com",
                password_hash="x",
                first_name="First",
                last_name="Last",
                role=UserRole.AGENT,
                is_active=True,
            )
            for user_id in user_ids
        ]


def make_user(role: UserRole) -> User:
    return User(
        id=uuid4(),
        email=f"{role.value}@example.com",
        password_hash="x",
        first_name="F",
        last_name="L",
        role=role,
        is_active=True,
    )


def make_ticket(creator_id):
    ticket = Ticket(
        id=uuid4(),
        created_by=creator_id,
        title="title",
        description="desc",
        state=TicketState.IN_PROGRESS,
    )
    ticket.assignee_links = []
    return ticket


def test_user_cannot_widen_created_by_scope():
    session = MagicMock()
    service = TicketService(session)
    service.repo = RepoStub()

    current_user = make_user(UserRole.USER)
    query = TicketListQuery(created_by=str(uuid4()))

    with pytest.raises(ForbiddenError):
        service.list_tickets_for_actor(current_user, query)

    assert not service.repo.list_tickets_called


def test_invalid_transition_is_rejected():
    session = MagicMock()
    service = TicketService(session)
    service.repo = RepoStub()
    service.user_repo = UserRepoStub()

    creator = make_user(UserRole.ADMIN)
    ticket = make_ticket(creator.id)
    service.repo.ticket = ticket

    payload = UpdateTicketRequest(state="pending")

    with pytest.raises(BadRequestError):
        service.update_ticket(creator, ticket.id, payload)


def test_only_admin_can_delete_ticket():
    session = MagicMock()
    service = TicketService(session)
    service.repo = RepoStub()

    creator = make_user(UserRole.USER)
    ticket = make_ticket(creator.id)
    service.repo.ticket = ticket

    with pytest.raises(ForbiddenError):
        service.delete_ticket(creator, ticket.id)

    admin = make_user(UserRole.ADMIN)
    service.delete_ticket(admin, ticket.id)
    assert service.repo.deleted
