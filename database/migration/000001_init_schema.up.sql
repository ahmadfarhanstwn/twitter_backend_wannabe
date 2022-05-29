CREATE TABLE "tweets" (
  "id" bigserial PRIMARY KEY,
  "tweet" varchar NOT NULL,
  "username" varchar NOT NULL,
  "likes" int DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT 'now()'
);

CREATE TABLE "relations" (
  "id" bigserial PRIMARY KEY,
  "follower_username" varchar NOT NULL,
  "followed_username" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT 'now()'
);

CREATE TABLE "like_relations" (
  "id" bigserial PRIMARY KEY,
  "username" varchar NOT NULL,
  "tweet_id" bigserial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "name" varchar NOT NULL,
  "followers_count" int DEFAULT 0,
  "following_count" int DEFAULT 0,
  "changed_password_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "tweets" ("username", "created_at");

CREATE INDEX ON "relations" ("follower_username", "followed_username");

CREATE INDEX ON "like_relations" ("username");

ALTER TABLE "tweets" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");

ALTER TABLE "relations" ADD FOREIGN KEY ("follower_username") REFERENCES "users" ("username");

ALTER TABLE "relations" ADD FOREIGN KEY ("followed_username") REFERENCES "users" ("username");

ALTER TABLE "like_relations" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");

ALTER TABLE "like_relations" ADD FOREIGN KEY ("tweet_id") REFERENCES "tweets" ("id");