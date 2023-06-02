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

-- user table stores user information
CREATE TABLE "auth"."user" (
  "id"       uuid	PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "password" char(60) NOT NULL -- bcrypt hash
);

-- session table stores user sessions
-- session data is stored in redis, however, by keeping track of sessions in the database
-- we can easily see the number of active sessions per user, as well as invalidate
-- all sessions for a specific user when necessary
CREATE TABLE "auth"."user_session" (
  "user_id"    uuid NOT NULL REFERENCES "auth"."user" ("id"),
  "session_id" char(36) NOT NULL, -- underlying type is uuid
  "created_at" timestamp without time zone DEFAULT (now() at time zone 'utc')
);
