SHELL := /bin/bash
.DEFAULT_GOAL := check
.NOTPARALLEL:

NODE ?= node
PNPM ?= pnpm
BASE ?= origin/main
# Phase 1 sets the generated output paths here.
GENERATED_PATHS :=
# renovate: datasource=npm depName=renovate
RENOVATE_VERSION := 44.65.2

.PHONY: setup setup-go setup-web gen gen-check fmt fmt-go fmt-web fmt-check \
	fmt-check-go fmt-check-web lint lint-go lint-web test test-go test-web \
	test-toolchain build build-go build-web check check-test-plan validate-renovate

setup: setup-go setup-web

setup-go:
	bash scripts/go-task.sh setup

setup-web:
	PNPM='$(PNPM)' $(NODE) scripts/check-toolchain.mjs
	$(PNPM) install --frozen-lockfile

# Phase 1 adds the OpenAPI generators here.
gen:
	@echo 'No code generators in Phase 0; skip'

gen-check:
	$(MAKE) gen
ifeq ($(strip $(GENERATED_PATHS)),)
	@echo 'No generated paths yet; skip'
else
	git diff --exit-code HEAD -- $(GENERATED_PATHS)
	@test -z "$$(git ls-files --others --exclude-standard -- $(GENERATED_PATHS))" || \
		{ echo 'Untracked generated files; add the intended files to Git'; exit 1; }
endif

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

validate-renovate:
	npx --yes -p renovate@$(RENOVATE_VERSION) renovate-config-validator
