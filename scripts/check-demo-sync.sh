#!/usr/bin/env bash
# Fail when internal/tui changes without a matching demo/ change.
# Usage: scripts/check-demo-sync.sh [base-ref]   (default: origin/main)
set -euo pipefail
base="${1:-origin/main}"
changed=$(git diff --name-only "$base"...HEAD)
if grep -q '^internal/tui/' <<<"$changed" && ! grep -q '^demo/' <<<"$changed"; then
  echo "internal/tui changed but demo/ did not; update demo/ (cast, gif, record.exp, simulate.sh) to match." >&2
  exit 1
fi
# demo-sync test
