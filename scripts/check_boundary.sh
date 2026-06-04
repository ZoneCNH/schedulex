#!/bin/sh
set -eu
[ "$(awk '/^module /{print $2}' go.mod)" = "github.com/ZoneCNH/schedulex" ] || { echo "bad module"; exit 1; }
[ ! -d pkg/templatex ] || { echo "pkg/templatex must be removed"; exit 1; }
! grep -R "github.com/ZoneCNH/xlib-standard\|pkg/templatex" -- . ':!/.git' >/dev/null 2>&1 || { echo "legacy identity remains"; exit 1; }
! grep -R "github.com/ZoneCNH/.*/x.go\|github.com/ZoneCNH/x.go\|github.com/ZoneCNH/x[a-z0-9_-]*/" --include='*.go' . >/dev/null 2>&1 || { echo "forbidden L2/x.go import"; exit 1; }
# Core package decisions use Clock; only the wall-clock adapter may call time.Now/After.
if grep -R "time\.Now\|time\.Sleep" --include='*.go' pkg/schedulex | grep -v 'pkg/schedulex/clock.go' | grep -v '_test.go'; then
  echo "core scheduling code must not call time.Now/time.Sleep directly"; exit 1
fi
