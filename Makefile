VER := 0.0.1
REV := $(shell git rev-parse HEAD | cut -c 1-7)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null || echo "master")
REPOSITORY := github.com/nerocrux/sample-rp

TESTPKGS = $(shell go list ./... | grep -v -e cmd -e test)
LINTPKGS = $(shell go list ./...)
FMTPKGS = $(foreach pkg,$(LINTPKGS),$(shell go env GOPATH)/src/$(pkg))

LDFLAGS_OPT = -X=main.version=${VER}-${REV}
LDFLAGS = -ldflags="$(LDFLAGS_OPT)"

PORT = 9001

SRVBIN = rp

IMAGE_TAG ?= latest

GOTEST ?= go test

all: rp


.PHONY: init
init:  ## install developer tools
	go get -u \
		github.com/kisielk/errcheck \
		golang.org/x/lint/golint \
		golang.org/x/tools/cmd/goimports \
		honnef.co/go/tools/cmd/staticcheck

.PHONY: rp 
rp:  ## build server binary
	go build $(LDFLAGS) -o $@ $(REPOSITORY)/cmd/sample-rp


.PHONY: run
run:  ## run rp command
	./$(SRVBIN)

