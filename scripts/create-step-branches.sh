#!/usr/bin/env bash
# Create Git branches for each Markdown spec step (one branch per file).
# Usage:
#   ./scripts/create-step-branches.sh list              # print branch names
#   ./scripts/create-step-branches.sh create-all      # create all from BASE_REF (default: main)
#   ./scripts/create-step-branches.sh create 05       # create only step 05
#
# Typical workflow: implement step N on branch step/NN-slug, open PR to main, merge, then start N+1 from updated main.

set -euo pipefail

BASE_REF="${BASE_REF:-main}"

STEPS=(
  "01:01_setup"
  "02:02_types"
  "03:03_storage"
  "04:04_httpclient"
  "05:05_tui_foundation"
  "06:06_sidebar"
  "07:07_request_editor"
  "08:08_response_display"
  "09:09_help_screen"
  "10:10_demo_collection"
  "11:11_integration"
)

branch_name() {
  local num="$1"
  local file_base="$2"
  local prefix="${num}_"
  local slug="${file_base#"$prefix"}"
  slug="${slug//_/-}"
  echo "step/${num}-${slug}"
}

ensure_base() {
  if ! git rev-parse --verify -q "$BASE_REF" >/dev/null; then
    echo "error: base ref '$BASE_REF' does not exist. Create it or set BASE_REF=..." >&2
    exit 1
  fi
}

cmd_list() {
  for entry in "${STEPS[@]}"; do
    num="${entry%%:*}"
    base="${entry#*:}"
    printf '%s  %s.md  ->  %s\n' "$num" "$base" "$(branch_name "$num" "$base")"
  done
}

cmd_create_all() {
  ensure_base
  for entry in "${STEPS[@]}"; do
    num="${entry%%:*}"
    base="${entry#*:}"
    b="$(branch_name "$num" "$base")"
    if git show-ref --verify --quiet "refs/heads/$b"; then
      echo "skip (exists): $b"
      continue
    fi
    git branch "$b" "$BASE_REF"
    echo "created: $b @ $BASE_REF"
  done
}

cmd_create_one() {
  local want="$1"
  want="${want#0}" # allow 05 or 5
  [[ "$want" =~ ^[0-9]+$ ]] || want=""
  ensure_base
  for entry in "${STEPS[@]}"; do
    num="${entry%%:*}"
    base="${entry#*:}"
    if [[ "$want" == "$num" ]] || [[ "0$want" == "$num" ]]; then
      b="$(branch_name "$num" "$base")"
      if git show-ref --verify --quiet "refs/heads/$b"; then
        echo "error: branch already exists: $b" >&2
        exit 1
      fi
      git branch "$b" "$BASE_REF"
      echo "created: $b @ $BASE_REF"
      return 0
    fi
  done
  echo "error: unknown step '$1' (use 01..11)" >&2
  exit 1
}

case "${1:-}" in
  list) cmd_list ;;
  create-all) cmd_create_all ;;
  create)
    [[ -n "${2:-}" ]] || { echo "usage: $0 create NN" >&2; exit 1; }
    cmd_create_one "$2"
    ;;
  *)
    echo "usage: $0 list | create-all | create NN" >&2
    exit 1
    ;;
esac
