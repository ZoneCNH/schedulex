#!/usr/bin/env bash
set -euo pipefail

mode="write"
if [[ "${1:-}" == "--check" ]]; then
  mode="check"
  shift
fi

manifest="${1:-release/manifest/latest.json}"
checksum="${2:-${manifest}.sha256}"

if [[ ! -s "${manifest}" ]]; then
  echo "release evidence manifest is missing or empty: ${manifest}" >&2
  exit 1
fi

if [[ "${mode}" == "check" ]]; then
  if [[ ! -s "${checksum}" ]]; then
    echo "release evidence checksum is missing or empty: ${checksum}" >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "${checksum}"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "${checksum}"
  else
    echo "sha256sum or shasum is required to verify ${checksum}" >&2
    exit 1
  fi
  exit 0
fi

mkdir -p "$(dirname "${checksum}")"

if command -v sha256sum >/dev/null 2>&1; then
  digest="$(sha256sum "${manifest}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  digest="$(shasum -a 256 "${manifest}" | awk '{print $1}')"
else
  echo "sha256sum or shasum is required to hash ${manifest}" >&2
  exit 1
fi

if [[ ! "${digest}" =~ ^[0-9a-f]{64}$ ]]; then
  echo "invalid sha256 digest for ${manifest}: ${digest}" >&2
  exit 1
fi

printf '%s  %s\n' "${digest}" "${manifest}" > "${checksum}"
printf '%s\n' "${digest}"
