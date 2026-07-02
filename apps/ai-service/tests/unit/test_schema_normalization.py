from app.enums.ticket_state import TicketState
from app.schemas.ticket import TicketListQuery, UpdateTicketRequest
from app.schemas.user import UserCreateRequest


def test_ticket_state_normalization_accepts_spaces():
    payload = UpdateTicketRequest(state="in progress")
    assert payload.state == TicketState.IN_PROGRESS


def test_ticket_state_normalization_accepts_underscores():
    payload = UpdateTicketRequest(state="in_progress")
    assert payload.state == TicketState.IN_PROGRESS


def test_user_email_and_skill_normalization():
    payload = UserCreateRequest(
        first_name="Jane",
        last_name="Doe",
        email="  JANE@EXAMPLE.COM  ",
        password="password123",
        skills=[" Incident-Management ", "incident-management", "log-analysis"],
    )
    assert payload.email == "jane@example.com"
    assert payload.skills == ["incident-management", "log-analysis"]


def test_ticket_list_query_validates_sort_and_order():
    query = TicketListQuery(sort="ticket_number", order="asc")
    assert query.sort == "ticket_number"
    assert query.order == "asc"
