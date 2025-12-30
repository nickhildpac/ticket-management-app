ALTER TABLE "tickets" DROP CONSTRAINT IF EXISTS "tickets_assigned_to_fkey";

ALTER TABLE "tickets"
  ALTER COLUMN "assigned_to" TYPE UUID[]
  USING CASE
    WHEN "assigned_to" IS NULL THEN NULL
    ELSE ARRAY["assigned_to"]
  END;
