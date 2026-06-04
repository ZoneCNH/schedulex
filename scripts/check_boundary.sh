#!/usr/bin/env bash
set -euo pipefail

export GOMODCACHE="${GOMODCACHE:-/tmp/schedulex-gomodcache}"

fail() {
  echo "boundary check failed: $*" >&2
  exit 1
}

module="$(awk '/^module /{print $2}' go.mod)"
[[ "$module" == "github.com/ZoneCNH/schedulex" ]] || fail "unexpected module: $module"
[[ ! -d pkg/templatex ]] || fail "pkg/templatex must not exist"

if rg -n 'github.com/ZoneCNH/(x\.go|redisx|postgresx|taosx|kafkax|ossx|clickhousex|natsx)' --glob '*.go' --glob 'go.mod' .; then
  fail "forbidden ZoneCNH L2/x.go dependency"
fi

if rg -n 'github.com/(redis|jackc/pgx|Shopify/sarama|segmentio/kafka-go|nats-io|ClickHouse)|go\.mongodb\.org|gorm\.io|aliyun-oss-go-sdk' --glob '*.go' --glob 'go.mod' .; then
  fail "forbidden runtime datastore/message dependency"
fi

if go list -deps ./... | rg -n '(^|/)github.com/ZoneCNH/(x\.go|redisx|postgresx|taosx|kafkax|ossx|clickhousex|natsx)($|/)'; then
  fail "forbidden dependency in go list graph"
fi

if rg -n '\b(redis|postgres|taos|kafka|oss|clickhouse|nats|macro|market|order|trade)\b' pkg/schedulex examples contracts --glob '!*.md'; then
  fail "business/L2 vocabulary leaked into L1 scheduler surface"
fi

if rg -n 'time\.(Now|Sleep|After|NewTimer|Tick|NewTicker)' pkg/schedulex --glob '*.go' --glob '!clock.go' --glob '!*_test.go'; then
  fail "core scheduling code must use Clock instead of direct time sources"
fi

echo "boundary check passed"
