#!/usr/bin/env bash
set -euo pipefail

module_path="github.com/stuffstash/stuff-stash"
domain_root="apps/api/internal/domain"

if [ "$#" -eq 0 ]; then
  exit 0
fi

if [ ! -d "$domain_root" ]; then
  exit 0
fi

failed=0

for file in "$@"; do
  case "$file" in
    "$domain_root"/*/*.go|"$domain_root"/*/*/*.go|"$domain_root"/*/*/*/*.go)
      [ -f "$file" ] || continue

      relative="${file#"$domain_root"/}"
      source_domain="${relative%%/*}"
      [ -n "$source_domain" ] || continue

      while IFS= read -r line; do
        imported_domain="${line#*internal/domain/}"
        imported_domain="${imported_domain%%/*}"

        if [ -n "$imported_domain" ] && [ "$imported_domain" != "$source_domain" ]; then
          echo "$file imports domain '$imported_domain' from domain '$source_domain': $line" >&2
          failed=1
        fi
      done < <(rg --no-heading --line-number --only-matching "\"$module_path/internal/domain/[^\"]+\"" "$file" || true)
      ;;
  esac
done

if [ "$failed" -ne 0 ]; then
  echo "direct cross-domain imports are not allowed; use application services, ports, or domain events" >&2
  exit 1
fi
