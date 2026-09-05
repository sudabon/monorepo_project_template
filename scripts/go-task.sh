#!/usr/bin/env bash
set -euo pipefail

task="${1:?usage: go-task.sh <setup|fmt|fmt-check|lint|test|build>}"
expected=$(awk '$1 == "go" { print $2 }' go.work)
export GOTOOLCHAIN="go$expected"
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
        lint) (cd "$module" && go vet ./...) ;;
        test) (cd "$module" && go test ./...) ;;
        build) (cd "$module" && go build ./...) ;;
      esac
      ;;
    *) echo "Unknown Go task: $task" >&2; exit 2 ;;
  esac
done <<< "$modules"

if [ "$task" = setup ]; then
  go work sync
fi
