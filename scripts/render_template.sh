#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/render_template.sh --module-name NAME --module-path PATH --package-name NAME --out DIR

Renders xlib-standard into a concrete base library by copying the repository,
moving pkg/templatex to pkg/<package>, and replacing template identifiers.
USAGE
}

module_name=""
module_path=""
package_name=""
out_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module-name)
      module_name="${2:-}"
      shift 2
      ;;
    --module-path)
      module_path="${2:-}"
      shift 2
      ;;
    --package-name)
      package_name="${2:-}"
      shift 2
      ;;
    --out)
      out_dir="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$module_name" || -z "$module_path" || -z "$package_name" || -z "$out_dir" ]]; then
  echo "ERROR: --module-name, --module-path, --package-name and --out are required" >&2
  usage >&2
  exit 2
fi

if [[ "$package_name" =~ [^a-zA-Z0-9_] || "$package_name" =~ ^[0-9] ]]; then
  echo "ERROR: --package-name must be a valid Go package identifier" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_abs="$(realpath "$repo_root")"
out_abs="$(realpath -m "$out_dir")"

if [[ "$out_abs" == "$repo_abs" || "$out_abs" == "$repo_abs"/* ]]; then
  echo "ERROR: output directory must be outside the template repository: $out_abs" >&2
  exit 2
fi

if [[ -e "$out_abs" && ! -d "$out_abs" ]]; then
  echo "ERROR: output path exists but is not a directory: $out_abs" >&2
  exit 2
fi

if [[ -e "$out_abs/.git" || -e "$out_abs/go.mod" ]]; then
  echo "ERROR: output directory looks like an existing repository: $out_abs" >&2
  exit 2
fi

if [[ -d "$out_abs" ]] && find "$out_abs" -mindepth 1 -maxdepth 1 | read -r _; then
  echo "ERROR: output directory must be empty: $out_abs" >&2
  exit 2
fi

mkdir -p "$out_abs"
out_dir="$out_abs"

(
  cd "$repo_root"
  tar \
    --exclude='./.git' \
    --exclude='./.omx' \
    --exclude='./.worktree' \
    --exclude='./.agent/inbox' \
    --exclude='./docs/adr' \
    --exclude='./docs/goal.md' \
    --exclude='./tmp' \
    --exclude='./dist' \
    --exclude='./node_modules' \
    --exclude='./coverage.out' \
    --exclude='./coverage.*' \
    --exclude='./*.coverprofile' \
    --exclude='./profile.cov' \
    --exclude='./release/manifest/latest.json' \
    --exclude='./release/manifest/latest.json.sha256' \
    --exclude='./release/debt/latest.json' \
    --exclude='./release/debt/latest.md' \
    --exclude='./release/debt/latest.json.sha256' \
    -cf - .
) | (
  cd "$out_dir"
  tar -xf -
)

if [[ "$package_name" != "templatex" ]]; then
  mkdir -p "$out_dir/pkg"
  mv "$out_dir/pkg/templatex" "$out_dir/pkg/$package_name"
fi

replace_in_text_files() {
  local find_text="$1"
  local replace_text="$2"

  while IFS= read -r -d '' file; do
    FIND_TEXT="$find_text" REPLACE_TEXT="$replace_text" perl -0pi -e 's/\Q$ENV{FIND_TEXT}\E/$ENV{REPLACE_TEXT}/g' "$file"
  done < <(
    find "$out_dir" -type f \( \
      -name '*.go' -o \
      -name '*.md' -o \
      -name '*.json' -o \
      -name '*.sh' -o \
      -name '*.yml' -o \
      -name '*.yaml' -o \
      -name 'Makefile' -o \
      -name 'go.mod' \
    \) -print0
  )
}

replace_in_text_files '{{MODULE_NAME}}' "$module_name"
replace_in_text_files '{{MODULE_PATH}}' "$module_path"
replace_in_text_files '{{PACKAGE_NAME}}' "$package_name"
replace_in_text_files 'github.com/ZoneCNH/xlib-standard' "$module_path"
replace_in_text_files 'github.com/ZoneCNH/baselib-template' "$module_path"
replace_in_text_files 'xlib-standard' "$module_name"
replace_in_text_files 'baselib-template' "$module_name"
package_title="$(printf '%s%s' "$(printf '%s' "${package_name:0:1}" | tr '[:lower:]' '[:upper:]')" "${package_name:1}")"
package_upper="$(printf '%s' "$package_name" | tr '[:lower:]' '[:upper:]')"
replace_in_text_files 'templatex_' "${package_name}_"
replace_in_text_files 'Templatex' "$package_title"
replace_in_text_files 'TEMPLATEX' "$package_upper"
replace_in_text_files 'templatex' "$package_name"

(
  cd "$out_dir"
  gofmt -w ./pkg ./internal ./contracts ./examples ./testkit
)

echo "rendered $module_name at $out_dir"
