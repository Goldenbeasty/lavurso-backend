create table if not exists
  "public"."assignments" (
    "id" serial primary key,
    "journal_id" INTEGER not null,
    "description" TEXT not null,
    "deadline" DATE not null,
    "type" TEXT not null,
    "created_at" TIMESTAMPTZ not null default NOW(),
    "updated_at" TIMESTAMPTZ not null default NOW()
  );

ALTER TABLE
  "public"."assignments"
ADD
  CONSTRAINT "assignments_relation_1" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;