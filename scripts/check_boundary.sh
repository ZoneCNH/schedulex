#!/usr/bin/env bash
set -euo pipefail

echo "checking forbidden dependency on x.go..."

DEPS="$(go list -deps ./...)"
FORBIDDEN_DEPS=(
  "github.com/bytechainx/x.go"
  "github.com/ZoneCNH/x.go"
)

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if grep -Fq "$dep" <<<"$DEPS"; then
    echo "ERROR: base library template must not depend on x.go dependency: $dep"
    exit 1
  fi
done

echo "checking forbidden internal dependency on public packages..."

if [[ -d ./internal && -d ./pkg ]]; then
  module_path="$(go list -m)"
  public_package_imports=()

  while IFS= read -r pkg_dir; do
    public_package_imports+=("${module_path}/pkg/$(basename "$pkg_dir")")
  done < <(find ./pkg -mindepth 1 -maxdepth 1 -type d | sort)

  if [[ ${#public_package_imports[@]} -gt 0 ]]; then
    internal_deps="$(go list -deps ./internal/...)"

    for dep in "${public_package_imports[@]}"; do
      if grep -Fxq "$dep" <<<"$internal_deps"; then
        echo "ERROR: internal runtime code must not depend on public package: $dep"
        exit 1
      fi
    done
  fi
fi

echo "checking forbidden rendered runtime dependency on template package..."

module_path="$(go list -m)"
template_module_path="github.com/ZoneCNH/xlib-standard"

if [[ "$module_path" != "$template_module_path" ]]; then
  search_roots=()
  for dir in ./pkg ./internal ./examples; do
    if [[ -d "$dir" ]]; then
      search_roots+=("$dir")
    fi
  done

  if [[ "${#search_roots[@]}" -eq 0 ]]; then
    echo "ERROR: no rendered source directories found for boundary check"
    exit 1
  fi

  while IFS= read -r file; do
    if grep -Fq "${template_module_path}/pkg/templatex" "$file"; then
      echo "ERROR: rendered runtime code must not depend on template package: $file"
      exit 1
    fi
  done < <(find "${search_roots[@]}" -type f -name '*.go' ! -name '*_test.go' -print)
fi

echo "checking forbidden business terms..."

FORBIDDEN_TERMS=(
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "BTCUSDT"
  "ETHUSDT"
  "Kline"
  "OrderBook"
  "Position"
  "RiskGate"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R --line-number --fixed-strings "$term" ./pkg ./internal --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "boundary check passed"
