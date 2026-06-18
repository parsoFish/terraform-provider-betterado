TEST?=$$(go list ./azuredevops/internal/acceptancetests |grep -v 'vendor')
UNITTEST?=$$(go list ./... |grep -v 'vendor')
PKG_NAME=azuredevops
TESTTIMEOUT=180m
TESTTAGS=all

ifeq ($(GOPATH),)
	GOPATH:=$(shell go env GOPATH)
endif

.EXPORT_ALL_VARIABLES:
  TF_SCHEMA_PANIC_ON_ERROR=1
  GO111MODULE=on

default: build

tools:
	@echo "==> installing required tooling..."
	go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
	go install github.com/katbyte/terrafmt@latest
	go install mvdan.cc/gofumpt@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(GOPATH)/bin" v1.64.8

build: fmtcheck depscheck
	go install

fmt:
	@echo "==> Fixing source code with gofmt..."
	@echo "# This logic should match the search logic in scripts/gofmtcheck.sh"
	find . -name '*.go' | grep -v vendor | grep -v third_party | xargs gofmt -s -w

fumpt:
	@echo "==> Fixing source code with Gofumpt..."
	# This logic should match the search logic in scripts/gofmtcheck.sh
	find . -name '*.go' | grep -v vendor | grep -v third_party | xargs gofumpt -s -w

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

terrafmt:
	@echo "==> Fixing acceptance test terraform blocks code with terrafmt..."
	@if command -v terrafmt; \
		then (find azuredevops | egrep "_test.go" | sort | while read f; do terrafmt fmt -f $$f; done) \
		else (find azuredevops | egrep "_test.go" | sort | while read f; do $(GOPATH)/bin/terrafmt fmt -f $$f; done); \
	  fi
	@echo "==> Fixing website terraform blocks code with terrafmt..."
	@if command -v terrafmt; \
		then (find . | egrep html.markdown | sort | while read f; do terrafmt fmt $$f; done); \
		else (find . | egrep html.markdown | sort | while read f; do $(GOPATH)/bin/terrafmt fmt $$f; done); \
	  fi

terrafmt-check:
	./scripts/terrafmt.sh

lint:
	@echo "==> Checking source code against linters..."
	golangci-lint run ./...

test: fmtcheck
	go test -v ./...

testacc: fmtcheck
	@echo "==> Sourcing .env file if available"
	if [ -f .env ]; then set -o allexport; . ./.env; set +o allexport; fi; \
	TF_ACC=1 go test -tags "$(TESTTAGS)" $(TEST) -v $(TESTARGS) -timeout 120m

# sweep reaps ADO projects left behind by acceptance runs that were killed,
# timed out, or panicked before their per-test cleanup could fire (test-acc-* /
# AccTest* names; real projects are allowlisted). Per-test t.Cleanup handles the
# normal path; this is the backstop. NOTE: the package path MUST precede -sweep.
sweep:
	@echo "==> Reaping leaked acceptance-test projects"
	@if [ -f secrets.env ]; then set -o allexport; . ./secrets.env; set +o allexport; fi; \
	go test -tags "$(TESTTAGS)" -count=1 -v ./azuredevops/internal/acceptancetests/ -timeout 40m -sweep=ado -sweep-run=betterado_project

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

install:
	./scripts/build.sh --SkipTests --Install

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)
	@echo "==> Checking source code with go mod vendor..."
	@go mod vendor
	@git diff --compact-summary --exit-code -- vendor || \
		(echo; echo "Unexpected difference in vendor/ directory. Run 'go mod vendor' command or revert any go.mod/go.sum/vendor changes and commit."; exit 1)

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

ci: depscheck lint test

clean-cache:
	@echo "==> Cleaning Go build and test caches..."
	go clean -cache -testcache
	@echo "==> Done. Disk space reclaimed."

# Generate Terraform Registry documentation (docs/) from the provider schema,
# the examples/ tree, and templates/. Pinned for reproducibility. See
# docs/RELEASING.md for the one-time legacy website/ -> docs/ migration.
TFPLUGINDOCS_VERSION ?= v0.20.0
docs:
	@echo "==> Generating registry docs with tfplugindocs $(TFPLUGINDOCS_VERSION)..."
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$(TFPLUGINDOCS_VERSION)
	tfplugindocs generate --provider-name betterado

.PHONY: build test testacc vet fmt fmtcheck lint tools test-compile clean-cache docs
