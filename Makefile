default: test

# -timeout 	timout in seconds
#  -v		verbose output
test:
	@ echo "+ $@"
	@ go test -timeout=5s -v

dev: build
	@ ./bin/slack-term -debug

# `CGO_ENABLED=0`
# Because of dynamically linked libraries, this will statically compile the
# app with all libraries built in. You won't be able to cross-compile if CGO
# is enabled. This is because Go binary is looking for libraries on the
# operating system it’s running in. We compiled our app, but it still is
# dynamically linked to the libraries it needs to run
# (i.e., all the C libraries it binds to). When using a minimal docker image
# the operating system doesn't have these libraries.
#
# `GOOS=linux`
# We're setting the OS to linux (in case someone builds the binary on Mac or
# Windows)
#
# `-mod=vendor`
# This ensures that the build process will use the modules in the vendor
# folder.
#
# `-a`
# Force rebuilding of package, all import will be rebuilt with cgo disabled,
# which means all the imports will be rebuilt with cgo disabled.
#
# `-installsuffix cgo`
# A suffix to use in the name of the package installation directory
#
# `-o`
# Output
#
# `./bin/slack-term`
# Placement of the binary
#
# `.`
# Location of the source files
build:
	@ echo "+ $@"
	@ CGO_ENABLED=0 go build -mod=vendor -a -installsuffix cgo -o ./bin/slack-term .

# Cross-compile
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
build-linux:
	@ echo "+ $@"
	@ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -a -installsuffix cgo -o ./bin/slack-term-linux-amd64 .

build-mac:
	@ echo "+ $@"
	@ GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -a -installsuffix cgo -o ./bin/slack-term-darwin-amd64 .

run: build
	@ echo "+ $@"
	@ ./bin/slack-term

install:
	@ echo "+ $@"
	@ go install .

modules:
	@ echo "+ $@"
	@ go mod tidy
	@ go mod vendor

build-all: build build-linux build-mac

.PHONY: default test build build-linux build-mac run install
