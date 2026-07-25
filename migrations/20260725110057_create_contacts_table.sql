-- +goose Up
CREATE TABLE IF NOT EXISTS contacts (
    id BIGSERIAL PRIMARY KEY,
    company_id UUID NOT NULL
        REFERENCES companies(id)
        ON DELETE CASCADE,
    name TEXT NOT NULL,
    email TEXT,
    linkedin_url TEXT,
    role TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contacts_company_id
    ON contacts(company_id);

-- +goose Down
DROP TABLE IF EXISTS contacts;