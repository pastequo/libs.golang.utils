## High-level targets

PROJECTS=gitutil logutil mongoutil panicrecover queryid prometrics

.PHONY: build check

build: build.tidy build.all
check: check.fmt check.imports check.lint check.test
publish: publish.projects


## Build target

.PHONY: build.tidy build.all

build.tidy:
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on go mod tidy; \
	done

build.all:
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on go build; \
	done


## Check target

LINT_FOLDER=$(PWD)/lint
METALINTER_COMMAND=gometalinter --enable-all --disable=gochecknoglobals --disable=gochecknoinits --disable=lll --min-confidence=0.0 --deadline=180s -j 4

.PHONY: check.fmt check.imports check.lint check.test check.lint.golangci check.lint.gometalinter

check.fmt:
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on gofmt -s -w ./*.go; \
	done

check.imports:
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on goimports -w ./*.go; \
	done

check.lint: check.lint.golangci check.lint.gometalinter

check.lint.golangci:
	@mkdir -p $(LINT_FOLDER)
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on golangci-lint run > $(LINT_FOLDER)/$$project.golangci 2>&1; \
	done

check.lint.gometalinter:
	@mkdir -p $(LINT_FOLDER)
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on $(METALINTER_COMMAND) ./... > $(LINT_FOLDER)/$$project.gometalinter 2>&1; \
	done

check.test:
	for project in $(PROJECTS); do \
		cd $(PWD)/$$project && GO111MODULE=on go test; \
	done


## Publish targets

VERSION=v0.0.1

.PHONY: publish.projects

publish.projects:
	for project in $(PROJECTS); do \
        git tag $$project/$(VERSION); \
    done





