include .project/gomod-project.mk

BIN=frame
SRC=frame

BINDIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build

all: tools build

tools:
	go install golang.org/x/tools/cmd/stringer@latest
	go install golang.org/x/tools/cmd/godoc@latest
	go install golang.org/x/tools/cmd/guru@latest
	go install golang.org/x/lint/golint@latest

build: config
	$(GOBUILD) ${BUILD_FLAGS} -o $(BINDIR)/$(BIN) main.go

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
	atlas schema apply -u postgres://postgres:postgres@127.0.0.1:15432/framework?sslmode=disable --to file://schema.hcl

pg_dump:
	pg_dump -d framework -h 127.0.0.1 -p 15432 -U postgres -W >> backup.sql

.PHONY: clean test config
