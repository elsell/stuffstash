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

  if [ -f apps/api/internal/adapters/httpserver/server_test.go ]; then
    server_test_names="$(rg --only-matching --replace '$1' --regexp '^func (Test[A-Za-z0-9_]+)' apps/api/internal/adapters/httpserver/server_test.go || true)"
    allowed_server_tests='^(TestHealthEndpointReturnsHealthyStatus|TestIndexEndpointReturnsHelpfulLinks|TestUnknownGetPathStillReturnsNotFound|TestOpenAPIIsGenerated)$'
    unexpected_server_tests="$(printf '%s\n' "$server_test_names" | rg --invert-match --regexp "$allowed_server_tests" || true)"
    if [ -n "$unexpected_server_tests" ]; then
      printf '%s\n' "$unexpected_server_tests" >&2
      echo "httpserver/server_test.go must stay limited to platform-level tests; domain endpoint tests belong in focused *_test.go files" >&2
      exit 1
    fi
  fi

  if [ -f apps/api/internal/adapters/httpserver/helpers_test.go ] && rg --line-number --regexp '^(type|func) (asset|auditRecord|customField|inventoryAccess|decodeAsset|decodeAuditRecord|decodeCustomField|decodeInventoryAccess|assertCustomField|assertInventoryAccess)' apps/api/internal/adapters/httpserver/helpers_test.go; then
    echo "httpserver/helpers_test.go must stay limited to cross-cutting test helpers; domain wire helpers belong in focused *_helpers_test.go files" >&2
    exit 1
  fi
fi
