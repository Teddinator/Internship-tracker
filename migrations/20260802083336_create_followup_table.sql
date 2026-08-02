-- +goose Up

CREATE TABLE IF NOT EXISTS followups (
    id BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL
        REFERENCES applications(id)
        ON DELETE CASCADE,
    due_date DATE NOT NULL,
    message TEXT NOT NULL,
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_followups_application_id
ON followups (application_id);

CREATE INDEX idx_followups_incomplete_due_date
ON followups (due_date)
WHERE completed_at is NULL;

-- +goose Down

SELECT 'down SQL query';
