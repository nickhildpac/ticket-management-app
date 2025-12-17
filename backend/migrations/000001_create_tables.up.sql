CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "first_name" varchar NOT NULL,
  "last_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "role" varchar DEFAULT 'user',
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tickets" (
  "id" bigserial PRIMARY KEY,
  "created_by" varchar NOT NULL,
  "assigned_to" varchar,
  "title" varchar NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz
);

CREATE TABLE "comments" (
  "id" bigserial PRIMARY KEY,
  "ticket_id" bigint NOT NULL,
  "created_by" varchar NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz
);

ALTER TABLE "tickets" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("username") ON DELETE SET NULL;

ALTER TABLE "tickets" ADD FOREIGN KEY ("assigned_to") REFERENCES "users" ("username") ON DELETE SET NULL;

ALTER TABLE "comments" ADD FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON DELETE CASCADE;

ALTER TABLE "comments" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("username") ON DELETE SET NULL;
