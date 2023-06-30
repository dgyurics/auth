CREATE SCHEMA "auth";

-- event table stores events that occur in the system.
-- uuid column is tied to a unique object/row in the system.
-- type column is used to identify the type of event that occurred
-- body column is used to store the data associated with the event
CREATE TABLE "auth"."event" (
  "id"         serial PRIMARY KEY not NULL,
  "uuid"       uuid NOT NULL,
  "type"       text NOT NULL,
  "body"       jsonb,
  "created_at" timestamp without time zone DEFAULT (now() at time zone 'utc')
);

-- user table stores user data
CREATE TABLE "auth"."user" (
  "id"       uuid	PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "password" char(60) NOT NULL -- bcrypt hash
);

-- session table stores user session data
CREATE TABLE "auth"."session" (
  "id" char(44) NOT NULL,
  "user_id"    uuid NOT NULL REFERENCES "auth"."user" ("id"),
  "created_at" timestamp without time zone DEFAULT (now() at time zone 'utc')
);
