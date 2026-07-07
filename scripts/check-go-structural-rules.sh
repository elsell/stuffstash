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

max_go_file_lines=800
oversized_files=()
for file in "${go_files[@]}"; do
  if head -n 1 "$file" | rg --quiet --regexp '^// Code generated .* DO NOT EDIT\.$'; then
    continue
  fi
  line_count="$(wc -l < "$file" | tr -d '[:space:]')"
  if [ "$line_count" -gt "$max_go_file_lines" ]; then
    oversized_files+=("$file:$line_count lines")
  fi
done

if [ "${#oversized_files[@]}" -gt 0 ]; then
  printf '%s\n' "${oversized_files[@]}" >&2
  echo "Go files over ${max_go_file_lines} lines are a structural smell; do serious organization refactoring before adding more code" >&2
  exit 1
fi

print_pattern='fmt\.Print(f|ln)?\(|(^|[^[:alnum:]_])println?\('
sql_pattern='"[^"]*(SELECT|INSERT[[:space:]]+INTO|UPDATE[[:space:]]+[^"]+[[:space:]]+SET|DELETE[[:space:]]+FROM|CREATE[[:space:]]+TABLE|ALTER[[:space:]]+TABLE|DROP[[:space:]]+TABLE)[[:space:]]+'
gorm_sql_fragment_pattern='\.(Joins|Where|Order)\("[^"]*(JOIN|[[:space:]]AND[[:space:]]|[[:space:]]OR[[:space:]]|[[:space:]]IN[[:space:]]|[[:space:]]IS[[:space:]]|[[:space:]]ASC|[[:space:]]DESC|[<>=?])'

if rg --line-number --regexp "$print_pattern" "${go_files[@]}"; then
  echo "ad hoc print statements are not allowed; use injected observability ports" >&2
  exit 1
fi

if rg --ignore-case --line-number --regexp "$sql_pattern" "${go_files[@]}"; then
	echo "raw SQL in Go application code is not allowed; use GORM behind repositories/adapters" >&2
	exit 1
fi

if rg --ignore-case --line-number --regexp "$gorm_sql_fragment_pattern" "${go_files[@]}"; then
	echo "raw SQL fragments in GORM calls are not allowed; use structured GORM clauses or typed repository helpers" >&2
	exit 1
fi

httpserver_touched=false
gormstore_touched=false
for file in "${go_files[@]}"; do
  case "$file" in
    apps/api/internal/adapters/httpserver/*)
      httpserver_touched=true
      ;;
    apps/api/internal/adapters/gormstore/*)
      gormstore_touched=true
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

if [ "$gormstore_touched" = true ]; then
  if [ -f apps/api/internal/adapters/gormstore/store.go ] && rg --line-number --regexp '^func \(s Store\) (Save|Create|Update|List|Claim|Mark|Asset|Tenant|Inventory|Custom|Attachment)' apps/api/internal/adapters/gormstore/store.go; then
    echo "gormstore/store.go must stay limited to Store construction; repository behavior belongs in focused *_repository.go files" >&2
    exit 1
  fi

  if [ -f apps/api/internal/adapters/gormstore/store_test.go ]; then
    store_test_names="$(rg --only-matching --replace '$1' --regexp '^func (Test[A-Za-z0-9_]+)' apps/api/internal/adapters/gormstore/store_test.go || true)"
    if [ -n "$store_test_names" ]; then
      printf '%s\n' "$store_test_names" >&2
      echo "gormstore/store_test.go must stay limited to shared helpers; repository tests belong in focused *_repository_test.go files" >&2
      exit 1
    fi
  fi
fi
