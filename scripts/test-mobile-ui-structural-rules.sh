#!/usr/bin/env bash

set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
checker="$repository_root/scripts/check-mobile-ui-structural-rules.sh"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

mkdir -p "$workdir/apps/mobile/src/ui/components" "$workdir/apps/mobile/src/ui/screens"

cat > "$workdir/apps/mobile/src/ui/components/AppTextInput.tsx" <<'EOF'
import { TextInput } from 'react-native';
export function AppTextInput() { return <TextInput />; }
EOF

cat > "$workdir/apps/mobile/src/ui/screens/Allowed.tsx" <<'EOF'
import type { TextInput } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
export const inputRef = null as TextInput | null;
export function Allowed() { return <AppTextInput />; }
EOF

"$checker" "$workdir/apps/mobile/src"

cat > "$workdir/apps/mobile/src/ui/screens/Bypassed.tsx" <<'EOF'
import { TextInput } from 'react-native';
export function Bypassed() { return <TextInput />; }
EOF

if "$checker" "$workdir/apps/mobile/src" >"$workdir/output" 2>&1; then
  echo "expected direct TextInput JSX to fail the mobile UI structural check" >&2
  exit 1
fi

grep -F "Bypassed.tsx" "$workdir/output" >/dev/null
grep -F "AppTextInput" "$workdir/output" >/dev/null

rm "$workdir/apps/mobile/src/ui/screens/Bypassed.tsx"
cat > "$workdir/apps/mobile/src/ui/screens/Aliased.tsx" <<'EOF'
import { TextInput as NativeInput } from 'react-native';
export function Aliased() { return <NativeInput />; }
EOF

if "$checker" "$workdir/apps/mobile/src" >"$workdir/output" 2>&1; then
  echo "expected an aliased native TextInput import to fail the mobile UI structural check" >&2
  exit 1
fi

grep -F "Aliased.tsx" "$workdir/output" >/dev/null

rm "$workdir/apps/mobile/src/ui/screens/Aliased.tsx"
cat > "$workdir/apps/mobile/src/ui/screens/Namespaced.tsx" <<'EOF'
import * as ReactNative from 'react-native';
export function Namespaced() { return <ReactNative.TextInput />; }
EOF

if "$checker" "$workdir/apps/mobile/src" >"$workdir/output" 2>&1; then
  echo "expected a namespaced native TextInput use to fail the mobile UI structural check" >&2
  exit 1
fi

grep -F "Namespaced.tsx" "$workdir/output" >/dev/null

rm "$workdir/apps/mobile/src/ui/screens/Namespaced.tsx"
cat > "$workdir/apps/mobile/src/ui/screens/RequiredNamespace.tsx" <<'EOF'
const ReactNative = require('react-native');
export function RequiredNamespace() { return <ReactNative.TextInput />; }
EOF

if "$checker" "$workdir/apps/mobile/src" >"$workdir/output" 2>&1; then
  echo "expected a required React Native namespace TextInput use to fail the mobile UI structural check" >&2
  exit 1
fi

grep -F "RequiredNamespace.tsx" "$workdir/output" >/dev/null

echo "mobile UI structural rule tests passed"
