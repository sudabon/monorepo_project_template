SHELL := /bin/bash
.DEFAULT_GOAL := check
.NOTPARALLEL:

NODE ?= node
PNPM ?= pnpm
BASE ?= origin/main
# renovate: datasource=npm depName=renovate
RENOVATE_VERSION := 44.65.2

.PHONY: setup setup-go setup-web gen gen-check fmt fmt-go fmt-web fmt-check \
	fmt-check-go fmt-check-web lint lint-go lint-web test test-go test-web \
	test-toolchain build build-go build-web check check-test-plan validate-renovate

setup: setup-go setup-web

setup-go:
	bash scripts/go-task.sh setup

setup-web:
	$(NODE) scripts/check-toolchain.mjs
	@test "$$($(PNPM) --version)" = "$$($(NODE) -p "require('./package.json').packageManager.split('@')[1]")"
	$(PNPM) install --frozen-lockfile

# Phase 1 adds the OpenAPI generators here.
gen:
	@echo 'No code generators in Phase 0; skip'

gen-check:
	$(MAKE) gen
	git diff --exit-code -- api/ apps/ packages/
	@test -z "$$(git ls-files --others --exclude-standard -- api/ apps/ packages/)" || \
		{ echo 'Untracked files in generation directories; add the intended files to Git'; exit 1; }

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

lint: lint-go lint-web

lint-go:
	bash scripts/go-task.sh lint

lint-web:
	$(NODE) scripts/check-toolchain.mjs
	@test "$$($(PNPM) --version)" = "$$($(NODE) -p "require('./package.json').packageManager.split('@')[1]")"
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

# Keep this order identical to the language workflows, even with make -j.
check:
	$(MAKE) fmt-check
	$(MAKE) lint
	$(MAKE) gen-check
	$(MAKE) test
	$(MAKE) build

check-test-plan:
	bash scripts/check-test-plan.sh "$(BASE)"

validate-renovate:
	$(PNPM) --package=renovate@$(RENOVATE_VERSION) dlx renovate-config-validator --strict renovate.json
