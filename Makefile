SHELL := /bin/bash
.DEFAULT_GOAL := check
.NOTPARALLEL:

NODE ?= node
PNPM ?= pnpm
BASE ?= origin/main
GENERATED_GO_PATHS := apps/api/internal/generated
GENERATED_WEB_PATHS := packages/api-client/src/generated
GENERATED_PATHS := $(GENERATED_GO_PATHS) $(GENERATED_WEB_PATHS)
# renovate: datasource=npm depName=renovate
RENOVATE_VERSION := 44.65.2
# Renovate 44.x runs on Node.js 24.x, not the repository-pinned Node.js.
# Keep in sync with node-version in .github/workflows/renovate.yml.
RENOVATE_NODE_MAJOR := 24

.PHONY: setup setup-go setup-web gen gen-go gen-web gen-check gen-check-go gen-check-web fmt fmt-go fmt-web fmt-check \
	fmt-check-go fmt-check-web lint lint-go lint-web test test-go test-web \
		test-toolchain lint-contract build build-go build-web check check-test-plan validate-renovate

setup: setup-go setup-web

setup-go:
	bash scripts/go-task.sh setup

setup-web:
	PNPM='$(PNPM)' $(NODE) scripts/check-toolchain.mjs
	$(PNPM) install --frozen-lockfile

gen: gen-go gen-web

gen-go:
	bash scripts/go-task.sh gen

gen-web:
	$(PNPM) exec openapi-typescript api/openapi.yaml --default-non-nullable false -o packages/api-client/src/generated/schema.d.ts

gen-check: gen
gen-check-go: GENERATED_PATHS := $(GENERATED_GO_PATHS)
gen-check-go: gen-go
gen-check-web: GENERATED_PATHS := $(GENERATED_WEB_PATHS)
gen-check-web: gen-web

gen-check gen-check-go gen-check-web:
	git diff --exit-code HEAD -- $(GENERATED_PATHS)
	@test -z "$$(git ls-files --others --exclude-standard -- $(GENERATED_PATHS))" || \
		{ echo 'Untracked generated files; add the intended files to Git'; exit 1; }

fmt: fmt-go fmt-web

fmt-go:
	bash scripts/go-task.sh fmt

fmt-web:
	$(PNPM) run fmt

fmt-check: fmt-check-go fmt-check-web

fmt-check-go:
	bash scripts/go-task.sh fmt-check

fmt-check-web:
	$(PNPM) run fmt:check

lint: lint-go lint-web lint-contract

lint-contract:
	$(PNPM) exec spectral lint api/openapi.yaml --ruleset api/.spectral.yaml --fail-severity warn

lint-go:
	bash scripts/go-task.sh lint

lint-web:
	PNPM='$(PNPM)' $(NODE) scripts/check-toolchain.mjs
	$(PNPM) run lint

test: test-go test-web

test-go:
	bash scripts/go-task.sh test

test-web: test-toolchain
	$(PNPM) run test

test-toolchain:
	$(NODE) --test tests/toolchain/*.test.mjs

build: build-go build-web

build-go:
	bash scripts/go-task.sh build

build-web:
	$(PNPM) run build

# Keep language checks identical to the language workflows, even with make -j.
# check-test-plan matches the dedicated E2E plan workflow.
check:
	$(MAKE) fmt-check
	$(MAKE) lint
	$(MAKE) gen-check
	$(MAKE) test
	$(MAKE) build
	$(MAKE) check-test-plan

check-test-plan:
	bash scripts/check-test-plan.sh "$(BASE)"

# CI-only: the dedicated renovate workflow provides Node.js $(RENOVATE_NODE_MAJOR).x.
validate-renovate:
	@major=$$($(NODE) -p 'process.versions.node.split(".")[0]'); \
		test "$$major" = '$(RENOVATE_NODE_MAJOR)' || { \
			echo "validate-renovate needs Node.js $(RENOVATE_NODE_MAJOR).x for renovate@$(RENOVATE_VERSION) (running $$major.x)." >&2; \
			echo "Run it through the 'Validate renovate.json' workflow, or switch Node.js locally." >&2; \
			exit 1; }
	npx --yes -p renovate@$(RENOVATE_VERSION) renovate-config-validator
