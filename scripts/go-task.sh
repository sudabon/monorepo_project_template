#!/usr/bin/env bash
set -euo pipefail

task="${1:?usage: go-task.sh <setup|gen|fmt|fmt-check|lint|test|test-unit|build|run-api|migrate-up|migrate-down>}"
expected=$(awk '$1 == "go" { print $2 }' go.work)
toolchain=$expected
case $expected in
  [0-9]*.[0-9]*.[0-9]*) ;;
  [0-9]*.[0-9]*) toolchain=$expected.0 ;;
  *)
    echo "go.work: go ${expected:-<missing>} must be X.Y or X.Y.Z" >&2
    exit 1
    ;;
esac
export GOTOOLCHAIN="go$toolchain"

case "$task" in
  run-api) exec go -C apps/api run ./cmd/api ;;
  migrate-up) exec go -C apps/api run ./cmd/migrate up ;;
  migrate-down) exec go -C apps/api run ./cmd/migrate down ;;
  test-unit) exec go -C apps/api test ./internal/domain ./internal/usecase ;;
  test)
    if [ -z "${TEST_DATABASE_URL:-}" ]; then
      docker compose up -d --wait postgres
      export TEST_DATABASE_URL="postgres://template:template@localhost:${POSTGRES_PORT:-55432}/template?sslmode=disable"
    fi
    ;;
esac

# Generation runs before the module loop: only apps/api holds the contract tool,
# and routing it here keeps GOTOOLCHAIN derived in exactly one place.
if [ "$task" = gen ]; then
  exec go -C apps/api tool oapi-codegen \
    -config ../../api/oapi-codegen.yaml ../../api/openapi.yaml
fi

gofmt_bin="$(go env GOROOT)/bin/gofmt"

modules=$(go list -m -f '{{if .Main}}{{.Dir}}{{end}}')
while IFS= read -r module; do
  [ -n "$module" ] || continue
  version=$(awk '$1 == "go" { print $2 }' "$module/go.mod")
  if [ "$version" != "$expected" ]; then
    echo "$module/go.mod: go $version must match go.work ($expected)" >&2
    exit 1
  fi
  echo "Go $task: $module"
  case "$task" in
    setup) (cd "$module" && go mod download) ;;
    fmt) "$gofmt_bin" -w "$module" ;;
    fmt-check)
      unformatted=$("$gofmt_bin" -l "$module")
      if [ -n "$unformatted" ]; then
        echo "$unformatted" >&2
        echo 'Run make fmt-go' >&2
        exit 1
      fi
      ;;
    lint|test|build)
      packages=$(cd "$module" && go list ./...)
      if [ -z "$packages" ]; then
        echo 'No Go packages yet; skip'
        continue
      fi
      case "$task" in
        lint)
          (cd "$module" && go vet ./...)
          if [ -f "$module/.go-arch-lint.yml" ]; then
            (cd "$module" && go tool go-arch-lint check)
          fi
          ;;
        test) (cd "$module" && go test -count=1 ./...) ;;
        build) (cd "$module" && go build ./...) ;;
      esac
      ;;
    *) echo "Unknown Go task: $task" >&2; exit 2 ;;
  esac
done <<< "$modules"

if [ "$task" = setup ]; then
  go work sync
fi
