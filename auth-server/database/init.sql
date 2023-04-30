CREATE SCHEMA "auth";

CREATE TABLE "auth"."event" (
  "id"         serial PRIMARY KEY not NULL,
  "uuid"       uuid NOT NULL,
  "type"       text NOT NULL,
  "body"       jsonb,
  "created_at" timestamp without time zone DEFAULT (now() at time zone 'utc')
);

CREATE TABLE "auth"."user" (
  "id"       uuid	PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "password" char(60) NOT NULL -- bcrypt hash
);
