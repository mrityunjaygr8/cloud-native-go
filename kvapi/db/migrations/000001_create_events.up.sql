CREATE TABLE IF NOT EXISTS "events" (
  "sequence" serial PRIMARY KEY,
  "key" varchar UNIQUE NOT NULL,
  "value" varchar NOT NULL,
  "eventType" integer NOT NULL
);

