#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
python3 "$script_dir/check-mobile-ui-structural-rules.py" "${1:-apps/mobile/src}"
