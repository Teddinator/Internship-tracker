#!/usr/bin/env bash

set -euo pipefail

docker compose run --rm api \
    sh -c 'goose -dir migrations postgres "$DATABASE_URL" down'