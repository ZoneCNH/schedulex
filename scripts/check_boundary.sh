#!/usr/bin/env bash
set -euo pipefail

export GOMODCACHE="${GOMODCACHE:-/tmp/schedulex-gomodcache}"

fail() {
  echo "boundary check failed: $*" >&2
  exit 1
}

search_all_go_mod() {
  local rg_pattern="$1"
  local grep_pattern="${2:-$rg_pattern}"

  if command -v rg >/dev/null 2>&1; then
    rg -n "$rg_pattern" --glob '*.go' --glob 'go.mod' .
    return $?
  fi

  find . -path ./.git -prune -o -type f \( -name '*.go' -o -name go.mod \) -print0 |
    xargs -0 -r grep -nE -- "$grep_pattern"
}

search_paths_without_markdown() {
  local rg_pattern="$1"
  local grep_pattern="$2"
  shift 2

  if command -v rg >/dev/null 2>&1; then
    rg -n "$rg_pattern" "$@" --glob '!*.md'
    return $?
  fi

  find "$@" -type f ! -name '*.md' -print0 |
    xargs -0 -r grep -nE -- "$grep_pattern"
}

search_core_time_sources() {
  local pattern='time\.(Now|Sleep|After|NewTimer|Tick|NewTicker)'

  if command -v rg >/dev/null 2>&1; then
    rg -n "$pattern" pkg/schedulex --glob '*.go' --glob '!clock.go' --glob '!*_test.go'
    return $?
  fi

  find pkg/schedulex -type f -name '*.go' ! -name 'clock.go' ! -name '*_test.go' -print0 |
    xargs -0 -r grep -nE -- "$pattern"
}

module="$(awk '/^module /{print $2}' go.mod)"
[[ "$module" == "github.com/ZoneCNH/schedulex" ]] || fail "unexpected module: $module"
[[ ! -d pkg/templatex ]] || fail "pkg/templatex must not exist"

if search_all_go_mod 'github.com/ZoneCNH/(x\.go|redisx|postgresx|taosx|kafkax|ossx|clickhousex|natsx)'; then
  fail "forbidden ZoneCNH L2/x.go dependency"
fi

if search_all_go_mod 'github.com/(redis|jackc/pgx|Shopify/sarama|segmentio/kafka-go|nats-io|ClickHouse)|go\.mongodb\.org|gorm\.io|aliyun-oss-go-sdk'; then
  fail "forbidden runtime datastore/message dependency"
fi

if command -v rg >/dev/null 2>&1; then
  graph_search=(rg -n '(^|/)github.com/ZoneCNH/(x\.go|redisx|postgresx|taosx|kafkax|ossx|clickhousex|natsx)($|/)')
else
  graph_search=(grep -nE '(^|/)github.com/ZoneCNH/(x\.go|redisx|postgresx|taosx|kafkax|ossx|clickhousex|natsx)($|/)')
fi

if go list -deps ./... | "${graph_search[@]}"; then
  fail "forbidden dependency in go list graph"
fi

if search_paths_without_markdown '\b(redis|postgres|taos|kafka|oss|clickhouse|nats|macro|market|order|trade)\b' '(^|[^[:alnum:]_])(redis|postgres|taos|kafka|oss|clickhouse|nats|macro|market|order|trade)([^[:alnum:]_]|$)' pkg/schedulex examples contracts; then
  fail "business/L2 vocabulary leaked into L1 scheduler surface"
fi

if search_core_time_sources; then
  fail "core scheduling code must use Clock instead of direct time sources"
fi

echo "boundary check passed"
