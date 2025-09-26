include .project/gomod-project.mk

BIN=frame
SRC=frame

BINDIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X frame/version.Version=$(VERSION) -X frame/version.GitCommit=$(COMMIT) -X frame/version.GitBranch=$(BRANCH) -X frame/version.BuildTime=$(BUILD_TIME)"

all: tools build

tools:
	go install golang.org/x/tools/cmd/stringer@latest
	go install golang.org/x/tools/cmd/godoc@latest
	go install golang.org/x/tools/cmd/guru@latest

build: config
	@$(GOBUILD) $(LDFLAGS) -o $(BINDIR)/$(BIN)

clean:
	-rm bin/*

config:
	mkdir -p $(BINDIR)

# Verify that there are no symlinks in our /vendor paths. Vendor'd directories sometime
# bring in symlinks that are recursive. This command will error out if 'find' encounters a recursive symlink
verify-symlinks:
	@find -L . -printf ""

skaffold-build:
	env GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(BIN)-linux-amd64 main.go

docker-build: skaffold-build
	docker build -t framework:dev .

migrate-up:
	atlas schema apply -u postgres://postgres:postgres@127.0.0.1:15432/framework?sslmode=require --to file://schema.hcl

migrate-diff:
	atlas migrate diff --env local

migrate-apply:
	atlas migrate apply --env local

pg_dump:
	pg_dump -d framework -h 127.0.0.1 -p 15432 -U postgres -W >> backup.sql

.PHONY: clean test config
