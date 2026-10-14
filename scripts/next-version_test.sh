#!/usr/bin/env bash
# Tests for scripts/next-version.sh. Usage: scripts/next-version_test.sh
set -uo pipefail
script="$(dirname "$0")/next-version.sh"
failures=0

expect() { # expect <want> <bump> <tags...>
  local want="$1" bump="$2"; shift 2
  local got
  got=$(printf '%s\n' "$@" | "$script" "$bump" 2>/dev/null)
  if [ "$got" != "$want" ]; then
    echo "FAIL: bump=$bump tags=[$*]: got '$got', want '$want'" >&2
    failures=$((failures + 1))
  fi
}

expect_error() { # expect_error <bump> <tags...>
  local bump="$1"; shift
  if printf '%s\n' "$@" | "$script" "$bump" >/dev/null 2>&1; then
    echo "FAIL: bump=$bump tags=[$*]: expected an error" >&2
    failures=$((failures + 1))
  fi
}

expect v0.1.1 patch v0.1.0
expect v0.2.0 minor v0.1.3
expect v1.0.0 major v0.4.2
expect v0.10.0 minor v0.9.0 v0.2.0 v0.10.0-rc1
expect v0.10.1 patch v0.2.0 v0.10.0 v0.9.5
expect v0.0.1 patch
expect v0.3.0 minor v0.2.1 not-a-version
expect_error bogus v0.1.0
expect_error ""

if [ "$failures" -gt 0 ]; then
  echo "$failures next-version test(s) failed" >&2
  exit 1
fi
echo "ok  scripts/next-version.sh"
