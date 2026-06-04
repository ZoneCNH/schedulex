#!/bin/sh
set -eu
[ "$(awk '/^module /{print $2}' go.mod)" = "github.com/ZoneCNH/schedulex" ] || { echo "bad module"; exit 1; }
[ ! -d pkg/templatex ] || { echo "pkg/templatex must be removed"; exit 1; }
for path in go.mod README.md docs .agent/harness.yaml Makefile contracts pkg examples scripts test; do
  [ -e "$path" ] || continue
  if grep -R "github.com/ZoneCNH/xlib-standard\|xlib-standard\|pkg/templatex\|templatex" "$path" >/dev/null 2>&1; then
    echo "legacy identity remains in $path"; exit 1
  fi
done
if grep -R "github.com/ZoneCNH/.*/x.go\|github.com/ZoneCNH/x.go\|github.com/ZoneCNH/x[a-z0-9_-]*/" --include='*.go' . >/dev/null 2>&1; then
  echo "forbidden L2/x.go import"; exit 1
fi
if grep -R "time\.Now\|time\.Sleep" --include='*.go' pkg/schedulex | grep -v 'pkg/schedulex/clock.go' | grep -v '_test.go'; then
  echo "core scheduling code must not call time.Now/time.Sleep directly"; exit 1
fi
