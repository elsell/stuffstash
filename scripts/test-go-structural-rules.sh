#!/usr/bin/env bash
set -euo pipefail

checker="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-go-structural-rules.sh"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

mkdir -p "$workdir/scripts" "$workdir/apps/api/internal/app/assets" "$workdir/apps/api/internal/ports"
cp "$checker" "$workdir/scripts/check-go-structural-rules.sh"
cp "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-asset-operation-facades.go" "$workdir/scripts/check-asset-operation-facades.go"
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
