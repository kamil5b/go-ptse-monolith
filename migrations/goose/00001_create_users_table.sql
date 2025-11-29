-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT false,
  activation_token TEXT,
  reset_token TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS users;
