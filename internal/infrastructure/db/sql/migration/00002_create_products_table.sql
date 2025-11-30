-- +goose Up
CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  created_by UUID,
  updated_at TIMESTAMP WITH TIME ZONE,
  updated_by UUID,
  deleted_at TIMESTAMP WITH TIME ZONE,
  deleted_by UUID
);

-- +goose Down
DROP TABLE IF EXISTS products;
