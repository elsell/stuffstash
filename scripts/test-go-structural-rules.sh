#!/usr/bin/env bash
set -euo pipefail

checker="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-go-structural-rules.sh"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

mkdir -p "$workdir/scripts" "$workdir/apps/api/internal/app/assets" "$workdir/apps/api/internal/ports"
cp "$checker" "$workdir/scripts/check-go-structural-rules.sh"
cp "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-asset-operation-facades.go" "$workdir/scripts/check-asset-operation-facades.go"
cp "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-gorm-query-fragments.go" "$workdir/scripts/check-gorm-query-fragments.go"
cp "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-go-sql-strings.go" "$workdir/scripts/check-go-sql-strings.go"
cd "$workdir"

cat > apps/api/internal/app/assets.go <<'EOF'
package app
type App struct{}
func (a App) CreateAssetWithOperation() {}
EOF
cat > apps/api/internal/app/assets/commands.go <<'EOF'
package assets
type Service struct{}
func (s Service) ArchiveAssetWithOperation() {}
EOF
cat > apps/api/internal/ports/assets.go <<'EOF'
package ports
type AssetRepository interface { CreateAsset() }
EOF

scripts/check-go-structural-rules.sh \
  apps/api/internal/app/assets.go \
  apps/api/internal/app/assets/commands.go \
  apps/api/internal/ports/assets.go

assert_rejected() {
  local file="$1"
  if scripts/check-go-structural-rules.sh "$file" >/dev/null 2>&1; then
    echo "expected legacy asset application facade to be rejected: $file" >&2
    exit 1
  fi
}

cat >> apps/api/internal/app/assets.go <<'EOF'
func (application App) CreateAsset() {}
EOF
assert_rejected apps/api/internal/app/assets.go

cat > apps/api/internal/app/assets.go <<'EOF'
package app
type App struct{}
func (a App) CreateAssetWithOperation() {}
EOF
cat >> apps/api/internal/app/assets/commands.go <<'EOF'
func (service *Service) RestoreAsset() {}
EOF
assert_rejected apps/api/internal/app/assets/commands.go

cat > apps/api/internal/app/assets.go <<'EOF'
package app
type App struct{}
func (App) RestoreAsset() {}
EOF
assert_rejected apps/api/internal/app/assets.go

cat > apps/api/internal/app/assets/commands.go <<'EOF'
package assets
type Service struct{}
func (*Service) CreateAsset() {}
EOF
assert_rejected apps/api/internal/app/assets/commands.go

cat > apps/api/internal/app/assets/commands.go <<'EOF'
package assets
type Service struct{}
func (_ Service) ArchiveAsset() {}
EOF
assert_rejected apps/api/internal/app/assets/commands.go

scripts/check-go-structural-rules.sh apps/api/internal/ports/assets.go

mkdir -p apps/api/internal/adapters/gormstore apps/api/internal/domain/search

assert_gorm_query_rejected() {
  local file="$1"
  if scripts/check-go-structural-rules.sh "$file" >/dev/null 2>&1; then
    echo "expected raw GORM query fragment to be rejected: $file" >&2
    exit 1
  fi
}

cat > apps/api/internal/adapters/gormstore/raw_order.go <<'EOF'
package gormstore
func rawOrder(db DB) { db.Order("created_at ASC") }
EOF
assert_gorm_query_rejected apps/api/internal/adapters/gormstore/raw_order.go

cat > apps/api/internal/adapters/gormstore/raw_multiline_where.go <<'EOF'
package gormstore
func rawMultilineWhere(db DB) {
  db.Where(
    "tenant_id = ? AND id = ?",
    "tenant",
    "resource",
  )
}
EOF
assert_gorm_query_rejected apps/api/internal/adapters/gormstore/raw_multiline_where.go

cat > apps/api/internal/domain/search/non_gorm_where.go <<'EOF'
package search
func filter(catalog Catalog) { catalog.Where("owner") }
EOF
scripts/check-go-structural-rules.sh apps/api/internal/domain/search/non_gorm_where.go

mkdir -p fixtures
cat > fixtures/diagnostic.go <<'EOF'
package fixture
import (
 "errors"
 f "fmt"
 cliports "github.com/stuffstash/stuff-stash/cli/internal/ports"
)
func diagnostics() {
 _ = errors.New("Select a saved context.")
 _ = f.Errorf("Select an inventory: %s", "example")
 _ = cliports.Failure("usage", "Select a saved context.")
}
EOF
scripts/check-go-structural-rules.sh fixtures/diagnostic.go
for source in 'const query = "sElEcT id FROM assets"' 'func query(db DB) { db.Query("SELECT 1") }' 'const query = `SELECT
id FROM assets`' 'const query = `UPDATE assets
SET name = ?`'  'func query(ports Other) { ports.Failure("usage", "SELECT id FROM assets") }'; do
 printf 'package fixture\n%b\n' "$source" > fixtures/sql.go
 if scripts/check-go-structural-rules.sh fixtures/sql.go >/dev/null 2>&1; then
  echo "expected raw SQL to be rejected: $source" >&2
  exit 1
 fi
done

cat > fixtures/shadow.go <<'EOF'
package fixture
import ports "github.com/stuffstash/stuff-stash/cli/internal/ports"
func query(ports Other) { ports.Failure("usage", "SELECT id FROM assets") }
EOF
if scripts/check-go-structural-rules.sh fixtures/shadow.go >/dev/null 2>&1; then
 echo "expected shadowed diagnostic function SQL to be rejected" >&2
 exit 1
fi
cat > fixtures/format_argument.go <<'EOF'
package fixture
import "fmt"
func query() { _ = fmt.Errorf("failed: %s", "SELECT id FROM assets") }
EOF
if scripts/check-go-structural-rules.sh fixtures/format_argument.go >/dev/null 2>&1; then
 echo "expected non-message SQL argument to be rejected" >&2
 exit 1
fi
