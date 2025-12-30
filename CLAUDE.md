# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a ticket management system with a React TypeScript frontend and Go backend, using PostgreSQL for persistence. The system implements a finite state machine for ticket lifecycle management with strict transition rules.

## Development Commands

### Frontend (React + Vite)
```bash
cd frontend
bun run dev          # Start development server (http://localhost:5173)
bun run build        # Build for production (runs TypeScript check + Vite build)
bun run lint         # Run ESLint
bun run preview      # Preview production build
```

### Backend (Go)
```bash
cd backend
go run cmd/api/main.go         # Run the API server
go test ./...                  # Run all tests
go test ./internal/domain/...  # Run specific package tests
```

### Database Migrations
The project uses manual SQL migrations in `backend/migrations/`. Migrations are numbered:
- `000001_create_tables.up.sql` / `.down.sql` - Initial schema (users, tickets, comments)
- `000002_update_ticket_comment_ids_to_uuid.up.sql` / `.down.sql` - Migrated from serial to UUID
- `000003_assigned_to_array.up.sql` / `.down.sql` - Changed assigned_to to UUID array

To apply migrations manually, execute the `.up.sql` files in order. To rollback, execute `.down.sql` files in reverse order.

## Architecture

### Technology Stack
- **Frontend**: React 19, TypeScript, Vite, Redux Toolkit, React Router v7, Tailwind CSS v4
- **Backend**: Go 1.24, Chi router, PostgreSQL, SQLC for type-safe DB queries
- **Authentication**: JWT with cookie-based refresh tokens

### Project Structure
```
/
├── frontend/
│   ├── src/
│   │   ├── pages/          # Route-level page components
│   │   ├── components/     # Reusable UI components
│   │   ├── store/          # Redux slices and configuration
│   │   ├── hooks/          # Custom React hooks
│   │   └── lib/            # API client (handles auto token refresh)
│   └── package.json
└── backend/
    ├── cmd/api/            # Application entry point
    ├── internal/
    │   ├── domain/         # Domain entities and business rules
    │   ├── application/    # Service layer
    │   ├── adapters/       # Database and HTTP adapters
    │   └── ports/          # Interface definitions
    └── migrations/         # Database migrations
```

### Key Architecture Patterns

**Backend (Hexagonal/Clean Architecture):**
- Domain entities in `internal/domain/` contain business logic (e.g., ticket state transitions)
- Services in `internal/application/service/` orchestrate business operations
- HTTP handlers in `internal/adapters/http/` handle requests/responses
- Database repositories in `internal/adapters/database/` abstract persistence
- SQLC generates type-safe database code from SQL queries

**Frontend (Redux + Component-based):**
- Redux Toolkit manages global state (auth, tickets, comments, users)
- Async thunks handle API calls with loading/error states
- React Router v7 handles client-side routing with auth guards
- Custom hooks encapsulate Redux logic and API interactions

## Domain Model

### Ticket State Machine

Tickets follow a finite state machine with enforced transition rules. State transitions are validated in the domain layer (`backend/internal/domain/ticket.go:91`):

```
Open → Pending, Cancelled
Pending → Open, Resolved, Cancelled
Resolved → Open, Pending, Closed, Cancelled
Closed → (final state, no transitions)
Cancelled → (final state, no transitions)
```

Key functions:
- `CanTransition(from, to TicketState) bool` - Validates state changes
- `GetValidTransitions(from TicketState) []TicketState` - Returns allowed next states
- `GetTicketState(s string) (TicketState, error)` - Parses string to enum

### Core Entities

**Ticket** (`backend/internal/domain/ticket.go`):
- `ID`: UUID (primary key)
- `CreatedBy`: UUID (foreign key to users)
- `AssignedTo`: `[]uuid.UUID` (PostgreSQL array for multiple assignees)
- `State`: Enum (Open, Pending, Resolved, Closed, Cancelled)
- `Priority`: Enum (Critical, High, Medium, Low)

**Comment**:
- `ID`: UUID
- `TicketID`: UUID
- `CreatedBy`: UUID
- `Description`: string

**User**:
- `ID`: UUID
- `Email`: unique
- `HashedPassword`: string
- `FirstName`, `LastName`, `Role`: string

### Authentication Flow

1. User logs in → receives JWT access token (15min expiry) + refresh token (cookie)
2. Frontend stores access token in Redux
3. API client (`frontend/src/lib/api.ts`) intercepts 401 responses
4. On 401, calls refresh endpoint to get new access token
5. Retries original request with new token

## API Structure

Main endpoints:
- `GET /ticket/all` - List all tickets
- `GET /ticket/assigned` - Get tickets assigned to current user
- `GET /ticket/{id}` - Get ticket details
- `POST /ticket` - Create ticket
- `PUT /ticket/{id}` - Update ticket (validates state transitions)
- `DELETE /ticket/{id}` - Delete ticket
- Auth endpoints: `/login`, `/signup`, `/refresh`

## Important Implementation Notes

### UUID Migration
The codebase recently migrated from serial IDs to UUIDs for tickets and comments. When working with database queries, always use UUID types, not integers.

### Array-based Assignments
The `assigned_to` field uses PostgreSQL arrays (`uuid[]`). SQL queries need special handling:
- Use `ANY($1)` for `WHERE assigned_to @> $1` (contains operator)
- For SQLC, ensure queries properly handle array parameters

### State Transition Validation
Always validate state transitions through the domain layer functions. Never bypass `CanTransition()` when updating ticket states.

### Type Safety
- Frontend: Full TypeScript with strict mode
- Backend: SQLC generates Go types from SQL queries - modify `.sql` files and regenerate to update types
