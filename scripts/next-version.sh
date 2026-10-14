#!/usr/bin/env bash
# Print the next release version given the bump and the released tags on stdin.
# Only plain vX.Y.Z tags count; with none, the base is v0.0.0.
# Usage: gh release list --exclude-drafts --limit 1000 --json tagName --jq '.[].tagName' \
#          | scripts/next-version.sh <patch|minor|major>
set -euo pipefail
bump="${1:-}"

case "$bump" in
  patch | minor | major) ;;
  *)
    echo "usage: $0 <patch|minor|major>  (released tags on stdin)" >&2
    exit 2
    ;;
esac

latest=$({ grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' || true; } | sed 's/^v//' | sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1)
IFS=. read -r major minor patch <<<"${latest:-0.0.0}"

case "$bump" in
  patch) patch=$((patch + 1)) ;;
  minor) minor=$((minor + 1)) patch=0 ;;
  major) major=$((major + 1)) minor=0 patch=0 ;;
esac

echo "v${major}.${minor}.${patch}"
