GITVERSION := $(shell git describe --tags --always)
GITCOMMIT := $(shell git log -1 --pretty=format:"%H")
GITTREESTATE := $(shell if git diff --quiet; then echo 'clean'; else echo 'dirty'; fi)
BUILDDATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS += -s -w
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.gitVersion=$(GITVERSION)
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.gitCommit=$(GITCOMMIT)
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.gitTreeState=$(GITTREESTATE)
LDFLAGS += -X github.com/toolctl/toolctl/internal/cmd.buildDate=$(BUILDDATE)
FLAGS = -ldflags "$(LDFLAGS)"

build:
	go build $(FLAGS)
