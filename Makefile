# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

.PHONY: build clean fmt lint test vet
.DEFAULT: all
all:  fmt vet lint build test

# Package list
PKGS=$(shell go list  ./...| grep -v /vendor/)

# Either define VER env variable, or use version detection using head hash
VER?=${version}

# Resolving binary dependencies for specific targets
GOLINT=$(shell which golint || echo '')
ver:
	{ \
	ver=${VER}; \
	if [ -z $$ver ]; then ver=`sh getver.sh`; fi; \
	echo "package cmd" >cmd/version.go; \
	echo "// Version constant" >>cmd/version.go; \
	echo "const Version = \"$$ver\"" >>cmd/version.go;\
	gofmt -s -w cmd/version.go;\
	}

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

build: ver fmt vet lint
	@echo "+ $@"
	@go build

build-prod: ver fmt vet lint
	@echo "+ $@"
	@go build -ldflags "-s -w"

dist: build-prod test
	{ \
	ver=${VER}; \
	if [ -z $$ver ]; then ver=`sh getver.sh`; fi; \
	mkdir lin ; GOOS=linux GOARCH=amd64 go build -o lin/took -ldflags "-s -w"; \
	tar cfz took-$$ver-linux-x86_64.tar.gz  readme.md LICENSE -C lin took ;\
	mkdir win ; GOOS=windows GOARCH=amd64 go build -o win/took -ldflags "-s -w" ;\
	tar cfz took-$$ver-windows-x86_64.tar.gz readme.md LICENSE -C win took ;\
	mkdir darwin ; GOOS=darwin GOARCH=amd64 go build -o darwin/took -ldflags "-s -w" ;\
	tar cfz took-$$ver-darwin-x86_64.tar.gz readme.md LICENSE -C darwin took ;\
	}

test:
	@echo "+ $@"
	go test  $(PKGS) -coverprofile=cover

clean:
	@echo "+ $@"

