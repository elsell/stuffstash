#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -eq 0 ]; then
  exit 0
fi

go_files=()
for file in "$@"; do
  case "$file" in
    *.go)
      if [ -f "$file" ]; then
        go_files+=("$file")
      fi
      ;;
  esac
done

if [ "${#go_files[@]}" -eq 0 ]; then
  exit 0
fi

print_pattern='fmt\.Print(f|ln)?\(|(^|[^[:alnum:]_])println?\('
sql_pattern='"[^"]*(SELECT|INSERT[[:space:]]+INTO|UPDATE[[:space:]]+[^"]+[[:space:]]+SET|DELETE[[:space:]]+FROM|CREATE[[:space:]]+TABLE|ALTER[[:space:]]+TABLE|DROP[[:space:]]+TABLE)[[:space:]]+'

if rg --line-number --regexp "$print_pattern" "${go_files[@]}"; then
  echo "ad hoc print statements are not allowed; use injected observability ports" >&2
  exit 1
fi

if rg --ignore-case --line-number --regexp "$sql_pattern" "${go_files[@]}"; then
  echo "raw SQL in Go application code is not allowed; use GORM behind repositories/adapters" >&2
  exit 1
fi
