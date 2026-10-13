#!/usr/bin/env bash
# Fail when the UI in internal/tui changes without a matching demo/ change.
# Import specs, comments and blank lines are not UI changes and are ignored.
# Usage: scripts/check-demo-sync.sh [base-ref]   (default: origin/main)
set -euo pipefail
base="${1:-origin/main}"

if git diff --quiet "$base"...HEAD -- demo/; then
  ui_changes=$(git diff -U0 "$base"...HEAD -- 'internal/tui/*.go' \
    | grep -E '^[+-]' \
    | grep -vE '^(\+\+\+|---)' \
    | grep -vE '^[+-][[:space:]]*(//.*)?$' \
    | grep -vE '^[+-][[:space:]]*import[[:space:]]*\(?[[:space:]]*$' \
    | grep -vE '^[+-][[:space:]]*([A-Za-z_.]+[[:space:]]+)?"[^"]+"[[:space:]]*$' || true)
  if [ -n "$ui_changes" ]; then
    echo "internal/tui changed but demo/ did not; update demo/ (cast, gif, record.exp, simulate.sh) to match." >&2
    echo "$ui_changes" >&2
    exit 1
  fi
fi
