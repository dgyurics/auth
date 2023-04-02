-- execute all ddl in auth schema
CREATE TABLE IF NOT EXISTS "event" (
 "id"         serial PRIMARY KEY not NULL,
 "uuid"       uuid NOT NULL,
 "type"       text NOT NULL,
 "body"       jsonb NOT NULL,
 "created_at" timestamp without time zone default (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS "user" (
  "id"       UUID	PRIMARY KEY,
  "username" VARCHAR(50) UNIQUE NOT NULL,
  "password" CHAR(60) NOT NULL -- bcrypt hash
);