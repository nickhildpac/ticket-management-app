-- Add new bigint columns for rollback
ALTER TABLE "tickets" ADD COLUMN "bigint_id" bigserial PRIMARY KEY;
ALTER TABLE "comments" ADD COLUMN "bigint_id" bigserial PRIMARY KEY;
ALTER TABLE "comments" ADD COLUMN "ticket_bigint_id" bigint;

-- Populate bigint columns with sequential IDs
UPDATE "tickets" SET "bigint_id" = nextval('tickets_bigint_id_seq'::regclass);
UPDATE "comments" SET "bigint_id" = nextval('comments_bigint_id_seq'::regclass);

-- Update ticket references in comments using row_number
UPDATE "comments" c1 
SET "ticket_bigint_id" = t.bigint_id
FROM (
    SELECT c.ticket_uuid_id, c.bigint_id as comment_id, t.bigint_id 
    FROM comments c
    JOIN tickets t ON c.ticket_uuid_id = t.id
    ORDER BY c.ticket_uuid_id, c.bigint_id
) t
WHERE c1.bigint_id = t.comment_id;

-- Drop current constraints
ALTER TABLE "tickets" DROP CONSTRAINT "tickets_pkey";
ALTER TABLE "comments" DROP CONSTRAINT "comments_pkey";
ALTER TABLE "comments" DROP CONSTRAINT "comments_ticket_id_fkey_new";

-- Drop current UUID columns
ALTER TABLE "tickets" DROP COLUMN "id" CASCADE;
ALTER TABLE "comments" DROP COLUMN "id" CASCADE;
ALTER TABLE "comments" DROP COLUMN "ticket_id" CASCADE;

-- Rename bigint columns to original names
ALTER TABLE "tickets" RENAME COLUMN "bigint_id" TO "id";
ALTER TABLE "comments" RENAME COLUMN "bigint_id" TO "id";
ALTER TABLE "comments" RENAME COLUMN "ticket_bigint_id" TO "ticket_id";

-- Add original constraints
ALTER TABLE "tickets" ADD PRIMARY KEY ("id");
ALTER TABLE "comments" ADD PRIMARY KEY ("id");
ALTER TABLE "comments" ADD FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON DELETE CASCADE;
ALTER TABLE "comments" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL;
ALTER TABLE "tickets" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL;
ALTER TABLE "tickets" ADD FOREIGN KEY ("assigned_to") REFERENCES "users" ("id") ON DELETE SET NULL;