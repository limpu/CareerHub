#!/usr/bin/env bash
set -e

echo "Running Documentation Verification Suite..."

REQUIRED_DOCS=(
    "doc/README.md"
    "doc/CONTEXT.md"
    "doc/PLAN.md"
    "doc/TASKS.md"
    "doc/PROGRESS.md"
    "doc/AI-AGENT-ARCHITECTURE.md"
)

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ ! -s "$doc" ]; then
        echo "[ERROR] Missing or empty: $doc"
        exit 1
    fi
    echo "  [OK] $doc exists and is non-empty"
done

echo "All documentation checks passed successfully!"