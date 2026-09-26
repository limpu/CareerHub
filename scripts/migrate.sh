#!/usr/bin/env bash
set -e

CMD=${1:-validate}
DIR=${2:-../../migrations}

echo "Running database migration tool ($CMD)..."
go run -C services/core ./cmd/migrate -dir "$DIR" -cmd "$CMD"