-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    website TEXT,
    industry TEXT,
    location TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS applications (
    id BIGSERIAL PRIMARY KEY,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    role TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'applied',
    applied_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_applications_company_id
    ON applications(company_id);

-- +goose Down
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS companies;
