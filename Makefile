GITVERSION := $(shell git describe --tags --always)
GITCOMMIT := $(shell git log -1 --pretty=format:"%H")
BUILDDATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS += -s -w
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.gitVersion=$(GITVERSION)
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.gitCommit=$(GITCOMMIT)
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.buildDate=$(BUILDDATE)
FLAGS = -ldflags "$(LDFLAGS)"

build:
	go build $(FLAGS)

lint:
	golangci-lint run

test:
	go test ./...
