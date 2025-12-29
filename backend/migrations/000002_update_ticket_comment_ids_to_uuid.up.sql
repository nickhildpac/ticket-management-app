-- First, drop all foreign key constraints to avoid conflicts
ALTER TABLE "comments" DROP CONSTRAINT IF EXISTS "comments_ticket_id_fkey";
ALTER TABLE "comments" DROP CONSTRAINT IF EXISTS "comments_created_by_fkey";
ALTER TABLE "tickets" DROP CONSTRAINT IF EXISTS "tickets_created_by_fkey";
ALTER TABLE "tickets" DROP CONSTRAINT IF EXISTS "tickets_assigned_to_fkey";

-- Add new UUID columns with temporary names
ALTER TABLE "tickets" ADD COLUMN "uuid_id" UUID DEFAULT gen_random_uuid();
ALTER TABLE "comments" ADD COLUMN "uuid_id" UUID DEFAULT gen_random_uuid();
ALTER TABLE "comments" ADD COLUMN "ticket_uuid_id" UUID;

-- Populate new UUID columns with random values
UPDATE "tickets" SET "uuid_id" = gen_random_uuid() WHERE "uuid_id" IS NULL;
UPDATE "comments" SET "uuid_id" = gen_random_uuid() WHERE "uuid_id" IS NULL;

-- Create a mapping between old and new IDs
CREATE TEMPORARY TABLE ticket_mapping AS
SELECT "id", "uuid_id" FROM "tickets";

-- Update ticket references in comments using the mapping
UPDATE "comments" 
SET "ticket_uuid_id" = tm."uuid_id" 
FROM ticket_mapping tm
WHERE "comments"."ticket_id" = tm."id";

-- Drop primary key constraints
ALTER TABLE "tickets" DROP CONSTRAINT IF EXISTS "tickets_pkey";
ALTER TABLE "comments" DROP CONSTRAINT IF EXISTS "comments_pkey";

-- Drop old columns
ALTER TABLE "tickets" DROP COLUMN IF EXISTS "id" CASCADE;
ALTER TABLE "comments" DROP COLUMN IF EXISTS "id" CASCADE;
ALTER TABLE "comments" DROP COLUMN IF EXISTS "ticket_id" CASCADE;

-- Rename new columns to final names
ALTER TABLE "tickets" RENAME COLUMN "uuid_id" TO "id";
ALTER TABLE "comments" RENAME COLUMN "uuid_id" TO "id";
ALTER TABLE "comments" RENAME COLUMN "ticket_uuid_id" TO "ticket_id";

-- Add primary key constraints
ALTER TABLE "tickets" ADD PRIMARY KEY ("id");
ALTER TABLE "comments" ADD PRIMARY KEY ("id");

-- Recreate foreign key constraints
ALTER TABLE "tickets" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL;
ALTER TABLE "tickets" ADD FOREIGN KEY ("assigned_to") REFERENCES "users" ("id") ON DELETE SET NULL;
ALTER TABLE "comments" ADD FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON DELETE CASCADE;
ALTER TABLE "comments" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL;

-- Drop temporary table
DROP TABLE IF EXISTS ticket_mapping;