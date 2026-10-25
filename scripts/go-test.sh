#!/usr/bin/env bash
set -euo pipefail

EXCLUDE="github.com/fool-pigeon/pigeon/cmd/pg-gui/build"

pkgs=$(go list ./... | grep -vF "$EXCLUDE")
[ -z "$pkgs" ] && { echo "no packages to test"; exit 0; }

go test "$@" $pkgs