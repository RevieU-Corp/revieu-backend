#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <commit-message-file>" >&2
  exit 2
fi

msg_file="$1"

if [ ! -f "$msg_file" ]; then
  echo "❌ COMMIT MESSAGE INVALID!" >&2
  echo "" >&2
  echo "Commit message file not found: $msg_file" >&2
  exit 1
fi

subject="$(sed -n '1p' "$msg_file")"
conventional_regex='^(feat|fix|docs|style|refactor|perf|test|chore|ci|build|revert)(\(.+\))?: .{1,}$'

if ! echo "$subject" | grep -qE "$conventional_regex"; then
  echo "❌ COMMIT MESSAGE INVALID!" >&2
  echo "" >&2
  echo "Subject must follow Conventional Commits format:" >&2
  echo "  <type>(<scope>): <description>" >&2
  echo "" >&2
  echo "Types: feat, fix, docs, style, refactor, perf, test, chore, ci, build, revert" >&2
  echo "Example: feat(core): add user authentication" >&2
  echo "" >&2
  echo "Current subject: $subject" >&2
  exit 1
fi

second_line="$(sed -n '2p' "$msg_file")"
if [ -n "$second_line" ]; then
  echo "❌ COMMIT MESSAGE INVALID!" >&2
  echo "" >&2
  echo "Second line must be blank to separate subject and body." >&2
  echo "Example:" >&2
  echo "  feat(core): add user authentication" >&2
  echo "  " >&2
  echo "  - implement auth service" >&2
  exit 1
fi

read -r body_found closes_found coauth_found <<EOF
$(awk '
NR <= 2 { next }
/^[[:space:]]*#/ { next }
/^[[:space:]]*$/ { next }
{
  if ($0 ~ /^(Close|Closes|Fix|Fixes|Resolve|Resolves)[[:space:]]+#[0-9]+([[:space:]].*)?$/) {
    closes = 1
    next
  }
  if ($0 ~ /^Co-Authored-By:[[:space:]]+.+<[^<>[:space:]]+@[^<>[:space:]]+>$/) {
    coauth = 1
    next
  }
  if ($0 ~ /^[A-Za-z-]+:[[:space:]].+$/) {
    next
  }
  body = 1
}
END { printf "%d %d %d\n", body, closes, coauth }
' "$msg_file")
EOF

missing=()

if [ "$body_found" -ne 1 ]; then
  missing+=("body")
fi

if [ "$closes_found" -ne 1 ]; then
  missing+=("issue reference trailer (e.g. Closes #106)")
fi

if [ "$coauth_found" -ne 1 ]; then
  missing+=("Co-Authored-By trailer")
fi

if [ "${#missing[@]}" -gt 0 ]; then
  echo "❌ COMMIT MESSAGE INVALID!" >&2
  echo "" >&2
  echo "Missing required section(s):" >&2
  for item in "${missing[@]}"; do
    echo "  - $item" >&2
  done
  echo "" >&2
  echo "Required format:" >&2
  echo "  <type>(<scope>): <description>" >&2
  echo "  " >&2
  echo "  <body>" >&2
  echo "  " >&2
  echo "  Closes #<issue-id>" >&2
  echo "  " >&2
  echo "  Co-Authored-By: Name <email@example.com>" >&2
  exit 1
fi

exit 0
