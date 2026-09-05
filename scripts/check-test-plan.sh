#!/usr/bin/env bash
set -euo pipefail
# テンプレート自身では無効。フラグなし・false・設定なしの案件では有効。
template_repo=false
if [ -f .openspec-e2e-kit.json ]; then
  template_repo=$(node --input-type=module -e '
    import { readFileSync } from "node:fs";
    const config = JSON.parse(readFileSync(".openspec-e2e-kit.json", "utf8"));
    console.log(config.templateRepo === true);
  ')
fi
if [ "$template_repo" = true ]; then
  echo 'templateRepo=true: E2E 計画ガードを skip'
  exit 0
fi

# test-plan.md を持つ change にタグ付きテストを要求する。
base="${1:-origin/main}"
ids=$(git diff --name-only "$base"...HEAD -- 'openspec/changes/**' \
  | awk -F/ '$3 != "archive" && NF >= 4 { print $3 }' | sort -u)
[ -z "$ids" ] && { echo "openspec change の差分なし。skip"; exit 0; }

fail=0
for id in $ids; do
  plan="openspec/changes/$id/test-plan.md"
  if [ ! -f "$plan" ]; then
    echo "$id: test-plan.md なし（E2E 不要）。skip"; continue
  fi
  escaped=$(printf '%s' "$id" | sed 's/[^A-Za-z0-9]/\\&/g')
  if [ ! -d tests/e2e/ ] || ! grep -rqE -- "@${escaped}([^A-Za-z0-9_-]|$)" tests/e2e/; then
    echo "::error::@$id タグ付きの E2E テストが tests/e2e/ にありません"; fail=1
  fi
done
exit $fail
