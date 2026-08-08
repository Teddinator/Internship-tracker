#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

BACKUP_DIR="${PROJECT_ROOT}/backups"
TIMESTAMP="$(date '+%Y-%m-%d_%H-%M-%S')"
BACKUP_FILE="${BACKUP_DIR}/internship_tracker_${TIMESTAMP}.sql.gz"

mkdir -p "${BACKUP_DIR}"

echo "Starting database..."

docker compose up --detach --wait db

echo "Creating backup..."

if docker compose exec --no-TTY db sh -c \
    'pg_dump \
        --username="$POSTGRES_USER" \
        --dbname="$POSTGRES_DB" \
        --clean \
        --if-exists \
        --no-owner \
        --no-privileges' \
    | gzip > "${BACKUP_FILE}"; then

    echo "Backup created:"
    echo "${BACKUP_FILE}"
else
    rm -f "${BACKUP_FILE}"
    echo "Backup failed." >&2
    exit 1
fi