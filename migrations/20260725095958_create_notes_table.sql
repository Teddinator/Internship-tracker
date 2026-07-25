-- +goose Up
CREATE TABLE IF NOT EXISTS notes (
    id BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL
        REFERENCES applications(id)
        ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notes_application_id
    on notes(application_id);

-- +goose Down
DROP TABLE IF EXISTS notes;
