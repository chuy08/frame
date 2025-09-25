-- Create "exercise_names" table
CREATE TABLE "exercise_names" (
 "id" uuid NOT NULL,
 "name" text NOT NULL,
 "created_at" timestamptz NOT NULL,
 "updated_at" timestamptz NOT NULL,
 PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "users" (
 "id" uuid NOT NULL,
 "first_name" character varying(100) NULL,
 "last_name" character varying(100) NULL,
 "email" character varying(100) NULL,
 "created_at" timestamptz NOT NULL,
 "updated_at" timestamptz NOT NULL,
 PRIMARY KEY ("id")
);
-- Create "address" table
CREATE TABLE "address" (
 "id" uuid NOT NULL,
 "user_id" uuid NOT NULL,
 "name" text NULL,
 "street" text NULL,
 "suite" text NULL,
 "city" text NULL,
 "state" text NULL,
 "zip" text NULL,
 "created_at" timestamptz NOT NULL,
 "updated_at" timestamptz NOT NULL,
 PRIMARY KEY ("id"),
 CONSTRAINT "user_fk" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "phone" table
CREATE TABLE "phone" (
 "id" uuid NOT NULL,
 "user_id" uuid NOT NULL,
 "name" text NOT NULL,
 "number" text NOT NULL,
 "created_at" timestamptz NOT NULL,
 "updated_at" timestamptz NOT NULL,
 PRIMARY KEY ("id"),
 CONSTRAINT "user_fk" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
