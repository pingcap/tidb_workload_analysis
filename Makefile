GO              := GO111MODULE=on go
GOBUILD         := $(GO) build $(BUILD_FLAG) -tags codes

LDFLAGS += -X "github.com/pingcap/index_advisor/version.GitHash=$(shell git rev-parse HEAD)"

default: build

build:
	CGO_ENABLED=1 $(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o bin/index-advisor ./
