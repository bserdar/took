# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

.PHONY: build clean fmt lint test vet
.DEFAULT: all
all:  fmt vet lint build test

# Package list
PKGS=$(shell go list  ./...| grep -v /vendor/)

# Resolving binary dependencies for specific targets
GOLINT=$(shell which golint || echo '')

vet:
	@echo "+ $@"
	@go vet  $(PKGS)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l cfg 2>&1 | tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false)
	@test -z "$$(gofmt -s -l cmd 2>&1 |  tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false)
	@test -z "$$(gofmt -s -l crypta 2>&1 | tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false)
	@test -z "$$(gofmt -s -l proto 2>&1 |  tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false)


lint:
	@echo "+ $@"
	-$(GOLINT) cfg/...
	-$(GOLINT) cmd/...
	-$(GOLINT) crypta/...
	-$(GOLINT) proto/...

build: fmt vet lint
	@echo "+ $@"
	@go build

build-prod: fmt vet lint
	@echo "+ $@"
	@go build -ldflags "-s -w"

dist: build-prod test
	{ \
	tag=`git tag -l --points-at HEAD`; \
	if [ ! -z $$tag ] ; then tag="-$$tag";	fi; \
	mkdir lin ; GOOS=linux GOARCH=amd64 go build -o lin/took -ldflags "-s -w"; \
	tar cfz took$$tag-linux-x86_64.tar.gz  readme.md LICENSE -C lin took ;\
	mkdir win ; GOOS=windows GOARCH=amd64 go build -o win/took -ldflags "-s -w" ;\
	tar cfz took$$tag-windows-x86_64.tar.gz readme.md LICENSE -C win took ;\
	mkdir darwin ; GOOS=darwin GOARCH=amd64 go build -o darwin/took -ldflags "-s -w" ;\
	tar cfz took$$tag-darwin-x86_64.tar.gz readme.md LICENSE -C darwin took ;\
	}

test:
	@echo "+ $@"
	go test  $(PKGS) -coverprofile=cover

clean:
	@echo "+ $@"

