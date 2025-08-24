postgres:
	docker run --name ticketmgmt -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -d postgres
createdb:
	docker exec -it ticketmgmt createdb --username=postgres --owner=postgres ticket_management
dropdb:
	docker exec -it ticketmgmt dropdb ticket_management
migrateup:
	migrate -path internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose up
migratedown:
	migrate -path internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose down
migrateup1:
	migrate -path internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose up 1
migratedown1:
	migrate -path internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose down 1
sqlc:
	sqlc generate
CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "first_name" varchar NOT NULL,
  "last_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "role" varchar,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tickets" (
  "id" bigserial PRIMARY KEY,
  "user" varchar NOT NULL,
  "balance" bigint NOT NULL,
  "currency" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz
);

CREATE TABLE "comments" (
  "id" bigserial PRIMARY KEY,
  "ticket_id" bigint NOT NULL,
  "user" varchar NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timeshtamptz
);

ALTER TABLE "tickets" ADD FOREIGN KEY ("user") REFERENCES "users" ("username");

ALTER TABLE "comments" ADD FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("user") REFERENCES "users" ("username");
test:
	go test -v -cover ./...
server:
	go run main.go
mock:
	~/go/bin/mockgen -package mockdb -destination db/mock/store.go github.com/nickhildpac/simplebank/db/sqlc Store
.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock
