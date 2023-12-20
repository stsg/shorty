B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)

info:
	- @echo "revision $(REV)"

build: info
	@ echo
	@ echo "Compiling Binary"
	@ echo
	cd cmd/shortener && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.revision=$(REV) -s -w" -o shortener
	# go build -o cmd/shortener/shortener cmd/shortener/*.go

clean:
	@ echo
	@ echo "Cleaning"
	@ echo
	rm cmd/shortener/shortener

test: build
	@ echo
	@ echo "Testing"
	@ echo
	shortenertest -test.v -test.run=^TestIteration1\$$ -binary-path=cmd/shortener/shortener

PHONY: all build test clean
