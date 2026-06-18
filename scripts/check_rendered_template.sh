#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/check_rendered_template.sh DIR MODULE_NAME MODULE_PATH PACKAGE_NAME

Checks that a rendered template has no stale template identifiers and exposes
the expected Go module and package directory.
USAGE
}

if [[ $# -ne 4 ]]; then
  usage >&2
  exit 2
fi

repo_dir="$1"
module_name="$2"
module_path="$3"
package_name="$4"

if [[ ! -d "$repo_dir" ]]; then
  echo "ERROR: rendered directory does not exist: $repo_dir" >&2
  exit 2
fi

actual_module="$(cd "$repo_dir" && GOWORK=off go list -m)"
if [[ "$actual_module" != "$module_path" ]]; then
  echo "ERROR: module path mismatch: got $actual_module, want $module_path" >&2
  exit 1
fi

if [[ ! -d "$repo_dir/pkg/$package_name" ]]; then
  echo "ERROR: rendered package directory missing: pkg/$package_name" >&2
  exit 1
fi

if [[ "$package_name" != "schedulex" && -e "$repo_dir/pkg/schedulex" ]]; then
  echo "ERROR: stale pkg/schedulex directory still exists" >&2
  exit 1
fi

rg_render_excludes=(
  --glob '!.git/**'
  --glob '!**/.git/**'
  --glob '!coverage.out'
  --glob '!**/coverage.out'
  --glob '!cover.out'
  --glob '!**/cover.out'
  --glob '!pkg.out'
  --glob '!**/pkg.out'
  --glob '!coverage.*'
  --glob '!**/coverage.*'
  --glob '!*.coverprofile'
  --glob '!**/*.coverprofile'
  --glob '!profile.cov'
  --glob '!**/profile.cov'
  --glob '!*.cov'
  --glob '!**/*.cov'
  --glob '!*.test'
  --glob '!**/*.test'
)

grep_render_excludes=(
  --exclude-dir=.git
  --exclude='coverage.out'
  --exclude='cover.out'
  --exclude='pkg.out'
  --exclude='coverage.*'
  --exclude='*.coverprofile'
  --exclude='profile.cov'
  --exclude='*.cov'
  --exclude='*.test'
)

scan_regex() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden "${rg_render_excludes[@]}" "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if grep -RInE "${grep_render_excludes[@]}" "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_fixed() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden "${rg_render_excludes[@]}" --fixed-strings "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if grep -RInF "${grep_render_excludes[@]}" "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_template_placeholders() {
  local pattern='\{\{[^}]+\}\}|TODO_TEMPLATE'

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden \
      --glob '!.git/**' \
      --glob '!**/.git/**' \
      --glob '!.github/workflows/**' \
      --glob '!**/.github/workflows/**' \
      --glob '!docs/adr/**' \
      --glob '!**/docs/adr/**' \
      --glob '!docs/archive/**' \
      --glob '!**/docs/archive/**' \
      --glob '!docs/goal.md' \
      --glob '!**/docs/goal.md' \
      --glob '!scripts/check_docs.sh' \
      --glob '!**/scripts/check_docs.sh' \
      --glob '!scripts/check_rendered_template.sh' \
      --glob '!**/scripts/check_rendered_template.sh' \
      --glob '!scripts/run_fuzz_smoke.sh' \
      --glob '!**/scripts/run_fuzz_smoke.sh' \
      --glob '!release/manifest/template.json' \
      --glob '!**/release/manifest/template.json' \
      "$pattern" "$repo_dir"; then
      echo "ERROR: found stale template placeholder" >&2
      exit 1
    fi
  else
    if find "$repo_dir" -type f \
      -not -path '*/.git/*' \
      -not -path '*/.github/workflows/*' \
      -not -path '*/docs/adr/*' \
      -not -path '*/docs/archive/*' \
      -not -path '*/docs/goal.md' \
      -not -path '*/scripts/check_docs.sh' \
      -not -path '*/scripts/check_rendered_template.sh' \
      -not -path '*/scripts/run_fuzz_smoke.sh' \
      -not -path '*/release/manifest/template.json' \
      -print0 | xargs -0 grep -InE "$pattern"; then
      echo "ERROR: found stale template placeholder" >&2
      exit 1
    fi
  fi
}

scan_template_placeholders
scan_fixed "github.com/ZoneCNH/schedulex" "module path"
scan_fixed "github.com/ZoneCNH/baselib-template" "module path"

if [[ "$module_name" != "schedulex" ]]; then
  scan_fixed "schedulex" "module name"
fi

if [[ "$module_name" != "baselib-template" ]]; then
  scan_fixed "baselib-template" "module name"
fi

if [[ "$package_name" != "schedulex" ]]; then
  scan_fixed "pkg/schedulex" "package directory reference"
  scan_fixed "schedulex_" "metrics prefix"
  scan_fixed "Templatex" "title-case package name"
  scan_fixed "TEMPLATEX" "upper-case package name"
  scan_regex '\bschedulex\b' "package name"
fi

echo "rendered template check passed: $module_name"
