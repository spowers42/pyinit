#!/usr/bin/env bash
# Cut a release by dispatching the Release workflow on main and watching it.
# See docs/adr/0001-releases-only-via-dispatch-workflow.md.
# Usage: scripts/release.sh <patch|minor|major>   (or: task release -- <bump>)
set -euo pipefail
bump="${1:-}"

case "$bump" in
  patch | minor | major) ;;
  *)
    echo "usage: task release -- <patch|minor|major>" >&2
    exit 2
    ;;
esac

next=$(gh release list --exclude-drafts --limit 1000 --json tagName --jq '.[].tagName' \
  | "$(dirname "$0")/next-version.sh" "$bump")
echo "Next version: $next ($bump)"

confirm_major=""
if [ "$bump" = "major" ]; then
  read -r -p "This is a major release. Type $next to confirm: " confirm_major
  if [ "$confirm_major" != "$next" ]; then
    echo "Confirmation did not match $next; aborting." >&2
    exit 1
  fi
fi

dispatched_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
output=$(gh workflow run release.yml --ref main -f bump="$bump" -f confirm_major="$confirm_major" 2>&1)
echo "$output"

run_id=$(grep -oE '/actions/runs/[0-9]+' <<<"$output" | grep -oE '[0-9]+$' || true)
for _ in $(seq 1 15); do
  [ -n "$run_id" ] && break
  sleep 2
  run_id=$(gh run list --workflow release.yml --event workflow_dispatch --limit 1 \
    --created ">=$dispatched_at" --json databaseId --jq '.[0].databaseId // empty')
done
if [ -z "$run_id" ]; then
  echo "Could not find the dispatched Release run; check: gh run list --workflow release.yml" >&2
  exit 1
fi

gh run watch "$run_id" --exit-status
gh release view "$next" --json url --jq .url
