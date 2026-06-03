#!/usr/bin/env bash
set -euo pipefail

echo "checking secrets..."

PATTERNS=(
  "password="
  "passwd="
  "secret="
  "token="
  "access_key="
  "secret_key="
  "AKIA[0-9A-Z]{16}"
  "ghp_[A-Za-z0-9_]{36,}"
  "github_pat_[A-Za-z0-9_]{20,}"
  "xox[baprs]-[A-Za-z0-9-]{10,}"
  "-----BEGIN [A-Z ]*PRIVATE KEY-----"
  "BEGIN RSA PRIVATE KEY"
  "BEGIN OPENSSH PRIVATE KEY"
)

for pattern in "${PATTERNS[@]}"; do
  if grep -R -E \
    --exclude-dir=.git \
    --exclude-dir=.omc \
    --exclude-dir=.omx \
    --exclude-dir=vendor \
    --exclude-dir=inbox \
    --exclude="*.sum" \
    --exclude="check_secrets.sh" \
    --exclude="goal.md" \
    -- "$pattern" .; then
    echo "ERROR: possible secret found: $pattern"
    exit 1
  fi
done

echo "secret check passed"
