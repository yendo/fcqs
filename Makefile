VERSION := $(shell git describe --tags --abbrev=0 | awk -F "." '{sub("v","", $$1); printf "%s.%s.%s\n",$$1,$$2,$$3+1}')

BINARY := fcqs-cli
GO_FILES := $(shell find . -type f -name '*.go') go.* shell.bash
GOCOVERDIR := coverdir

$(BINARY): $(GO_FILES)
	go build -trimpath -ldflags "-s -w -X main.version=${VERSION}-snapshot" ./cmd/fcqs-cli/

test/$(BINARY): $(GO_FILES)
	go build -o $@ -cover -trimpath -ldflags "-s -w -X main.version=0.0.0-test" ./cmd/fcqs-cli/

.PHONY: test
test: unit-test integration-test

.PHONY: test-v
test-v: unit-test-v integration-test-v

.PHONY: unit-test
unit-test:
	go test -shuffle=on -cover -coverprofile cover-ut.out ./cmd/fcqs-cli/ .

.PHONY: unit-test-v
unit-test-v:
	go test -v -shuffle=on -cover -coverprofile cover-ut.out ./cmd/fcqs-cli/ .

.PHONY: integration-test
integration-test: test/$(BINARY)
	@mkdir -p coverdir
	GOCOVERDIR=coverdir go test -shuffle=on ./test
	@go tool covdata percent -i=$(GOCOVERDIR)
	@go tool covdata textfmt -i=$(GOCOVERDIR) -o cover-it.out

.PHONY: integration-test-v
integration-test-v: test/$(BINARY)
	@mkdir -p coverdir
	GOCOVERDIR=coverdir go test -v -shuffle=on ./test
	@go tool covdata percent -i=$(GOCOVERDIR)
	@go tool covdata textfmt -i=$(GOCOVERDIR) -o cover-it.out

.PHONY: super-linter
super-linter: clean
	docker run -e RUN_LOCAL=true -e USE_FIND_ALGORITHM=true \
		-e FILTER_REGEX_EXCLUDE=".*/testdata/.*" \
		-e BASH_EXEC_IGNORE_LIBRARIES=true \
		-e VALIDATE_GO=false \
		-e VALIDATE_JSON_PRETTIER=false \
		-e VALIDATE_MARKDOWN_PRETTIER=false \
		-e VALIDATE_YAML_PRETTIER=false \
		-v ${PWD}:/tmp/lint/ ghcr.io/super-linter/super-linter:slim-v7

.PHONY: clean
clean:
	rm -rf $(BINARY) test/$(BINARY) dist cover*
