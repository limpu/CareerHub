#!/usr/bin/env bash
set -e

echo "Bootstrapping Career / LinkedIn / Social Platform..."

if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "Created .env from .env.example"
fi

pnpm install
pnpm -r --filter=./packages/* build
(cd services/core && go build ./cmd/...)

echo "Bootstrap completed successfully!"