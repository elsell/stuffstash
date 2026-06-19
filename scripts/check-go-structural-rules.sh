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

httpserver_touched=false
for file in "${go_files[@]}"; do
  case "$file" in
    apps/api/internal/adapters/httpserver/*)
      httpserver_touched=true
      ;;
  esac
done

if [ "$httpserver_touched" = true ]; then
  if rg --line-number --regexp 'huma\.(Get|Post|Patch|Put|Delete)\(' apps/api/internal/adapters/httpserver/server.go apps/api/internal/adapters/httpserver/api.go; then
    echo "httpserver server.go/api.go must compose only; domain route registration belongs under <domain>/routes/" >&2
    exit 1
  fi

  if find apps/api/internal/adapters/httpserver -path '*/routes/*.go' -print0 | xargs -0 rg --line-number --regexp '^type [A-Za-z0-9_]+ (struct|interface)' ; then
    echo "httpserver route files must not define DTOs or interfaces; use the domain dto/ or mapper/ package" >&2
    exit 1
  fi

  if find apps/api/internal/adapters/httpserver -path '*/dto/*.go' -print0 | xargs -0 rg --line-number --regexp 'internal/(app|domain|ports)' ; then
    echo "httpserver DTO files must not import app, domain, or port packages" >&2
    exit 1
  fi

  if find apps/api/internal/adapters/httpserver -path '*/mapper/*.go' -print0 | xargs -0 rg --line-number --regexp 'huma\.(Get|Post|Patch|Put|Delete)\(' ; then
    echo "httpserver mapper files must not register routes" >&2
    exit 1
  fi

  if find apps/api/internal/adapters/httpserver -path '*/mapper/*.go' -print0 | xargs -0 rg --line-number --regexp 'internal/app|httpserver/shared|shared\.SuccessEnvelope|PaginatedMeta' ; then
    echo "httpserver mapper files must only translate between domain and DTO shapes; envelopes, app services, and pagination metadata belong in routes" >&2
    exit 1
  fi
fi
