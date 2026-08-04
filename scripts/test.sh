#!/usr/bin/env  bash

set -e

docker compose run --rm api go test ./...