#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
SEED_FILE="${PROJECT_ROOT}/seeds/dev.sql"

cd "${PROJECT_ROOT}"

if [[ ! -f .env ]]; then
    echo "Error: .env missing" >&2
    echo "Create it with:"
    echo "  cp  .env.example .env"
    exit 1
fi

if [[ ! -f "${SEED_FILE}" ]]; then
    echo "Error: seed-file missing: ${SEED_FILE}" >&2
    exit 1
fi

echo "Initilazing database..."
docker compose up --detach --wait db

echo "Running seed-data..."
docker compose exec --no-TTY db sh -c \
    'psql --username="$POSTGRES_USER" --dbname="$POSTGRES_DB"' \
    < "${SEED_FILE}"

echo "Seeding is completed."