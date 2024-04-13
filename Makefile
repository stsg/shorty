###
### Makefile
###

VERSION=0.0.1dev

B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)
DATE=$(shell date +%Y%m%d-%H:%M:%S)
COMMIT=$(shell git log -n 1 --pretty=format:"%H")

info:
	- @echo "revision $(REV)"

build: info
	@ echo
	@ echo "Compiling Binary"
	@ echo
	cd cmd/shorty && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o shorty

build_macos: info
	@ echo
	@ echo "Compiling Binary for MacOS"
	@ echo
	cd cmd/shorty && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o shorty

tidy:
	@ echo
	@ echo "Tidying"
	@ echo
	go mod tidy

clean:
	@ echo
	@ echo "Cleaning"
	@ echo
	rm cmd/shorty/shorty

utest: build
	@ echo
	@ echo "Unit testing"
	@ echo
	go test ./...

test: build
	@ echo
	@ echo "Testing"
	@ echo
	shortytest -test.v -test.run=^TestIteration1\$$ -binary-path=cmd/shorty/shorty

run:
	@ echo
	@ echo "Runnig"
	@ echo
	go run cmd/shorty/main.go -d "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"

PHONY: build tidy clean utest test run
