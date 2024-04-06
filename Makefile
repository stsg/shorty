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
	cd cmd/shortener && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o shortener
	# go build -o cmd/shortener/shortener cmd/shortener/*.go

build_macos: info
	@ echo
	@ echo "Compiling Binary for MacOS"
	@ echo
	cd cmd/shortener && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o shortener
	# go build -o cmd/shortener/shortener cmd/shortener/*.go

tidy:
	@ echo
	@ echo "Tidying"
	@ echo
	go mod tidy

clean:
	@ echo
	@ echo "Cleaning"
	@ echo
	rm cmd/shortener/shortener

utest: build
	@ echo
	@ echo "Unit testing"
	@ echo
	go test ./...

test: build
	@ echo
	@ echo "Testing"
	@ echo
	shortenertest -test.v -test.run=^TestIteration1\$$ -binary-path=cmd/shortener/shortener

run:
	@ echo
	@ echo "Runnig"
	@ echo
	go run cmd/shortener/main.go -d "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"

PHONY: build tidy clean utest test run
