CREATE TABLE "users" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "hashed_password" varchar NOT NULL,
  "first_name" varchar NOT NULL,
  "last_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "role" varchar DEFAULT 'user',
  "updated_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "tickets" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "created_by" UUID NOT NULL,
  "assigned_to" UUID[],
  "title" varchar NOT NULL,
  "description" varchar NOT NULL,
  "state" INT NOT NULL DEFAULT 1,
  "priority" INT NOT NULL DEFAULT 4,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL
);

CREATE TABLE "comments" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "ticket_id" UUID NOT NULL,
  "created_by" UUID NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL
);

ALTER TABLE "tickets" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL;

ALTER TABLE "comments" ADD FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON DELETE CASCADE;

ALTER TABLE "comments" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL;
