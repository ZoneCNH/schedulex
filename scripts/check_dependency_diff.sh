#!/usr/bin/env bash
set -euo pipefail

echo "checking dependency automation..."

required_files=(
  "renovate.json"
  ".github/dependabot.yml"
  "go.mod"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: missing dependency automation file: $file" >&2
    exit 1
  fi
done

require_text() {
  local file="$1"
  local needle="$2"

  if ! grep -Fq -- "$needle" "$file"; then
    echo "ERROR: $file must mention: $needle" >&2
    exit 1
  fi
}

require_text "renovate.json" '"gomod"'
require_text "renovate.json" '"github-actions"'
require_text "renovate.json" '"automerge": false'
require_text "renovate.json" '"major"'
require_text "renovate.json" '"dependencyDashboardApproval": true'
require_text "renovate.json" '"contracts/**"'
require_text "renovate.json" '"docs/standard/**"'
require_text "renovate.json" '"scripts/**"'
require_text ".github/dependabot.yml" 'package-ecosystem: "gomod"'
require_text ".github/dependabot.yml" 'package-ecosystem: "github-actions"'
require_text ".github/dependabot.yml" 'update-types:'
require_text ".github/dependabot.yml" '"minor"'
require_text ".github/dependabot.yml" '"patch"'

echo "Go module dependencies:"
if command -v go >/dev/null 2>&1; then
  GOWORK=off go list -m all
else
  echo "WARN: go is not installed; falling back to go.mod module declaration" >&2
  sed -n '1,20p' go.mod
fi

echo "GitHub Actions dependencies:"
if [[ -d ".github/workflows" ]]; then
  if command -v rg >/dev/null 2>&1; then
    rg -n 'uses:[[:space:]]*[^[:space:]]+' .github/workflows || true
  else
    grep -REn 'uses:[[:space:]]*[^[:space:]]+' .github/workflows || true
  fi
else
  echo "WARN: .github/workflows does not exist" >&2
fi

changed_files=()
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  base_ref="${DEPENDENCY_DIFF_BASE:-}"
  if [[ -z "$base_ref" ]]; then
    if git rev-parse --verify origin/main >/dev/null 2>&1; then
      base_ref="origin/main"
    elif git rev-parse --verify main >/dev/null 2>&1; then
      base_ref="main"
    fi
  fi

  if [[ -n "$base_ref" ]]; then
    while IFS= read -r file; do
      [[ -n "$file" ]] && changed_files+=("$file")
    done < <(git diff --name-only "$base_ref"...HEAD)
  fi

  while IFS= read -r file; do
    [[ -n "$file" ]] && changed_files+=("$file")
  done < <(git diff --name-only)

  while IFS= read -r file; do
    [[ -n "$file" ]] && changed_files+=("$file")
  done < <(git diff --cached --name-only)

  while IFS= read -r file; do
    [[ -n "$file" ]] && changed_files+=("$file")
  done < <(git ls-files --others --exclude-standard)
fi

if ((${#changed_files[@]} > 0)); then
  mapfile -t changed_files < <(printf '%s\n' "${changed_files[@]}" | sort -u)

  echo "Changed files considered for dependency governance:"
  printf '  - %s\n' "${changed_files[@]}"
else
  echo "Changed files considered for dependency governance: none"
fi

dependency_surface=false
review_surface=false

for file in "${changed_files[@]}"; do
  case "$file" in
    go.mod|go.sum|renovate.json|.github/dependabot.yml|.github/workflows/*)
      dependency_surface=true
      ;;
  esac

  case "$file" in
    contracts/*|docs/standard/*|scripts/*|cmd/goalcli/*|internal/tools/*|release/manifest/template.json)
      review_surface=true
      ;;
  esac
done

echo "dependency_surface_changed=$dependency_surface"
echo "standard_contract_generator_review_required=$review_surface"

if [[ "$review_surface" == "true" ]]; then
  echo "Review required: standard, contract, generator, harness, or evidence surface changed."
fi

echo "dependency automation check passed"
