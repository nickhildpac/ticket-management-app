ALTER TABLE "tickets"
  ALTER COLUMN "assigned_to" TYPE UUID
  USING CASE
    WHEN "assigned_to" IS NULL OR array_length("assigned_to", 1) IS NULL OR array_length("assigned_to", 1) = 0 THEN NULL
    ELSE "assigned_to"[1]
  END;

ALTER TABLE "tickets" ADD FOREIGN KEY ("assigned_to") REFERENCES "users" ("id") ON DELETE SET NULL;
