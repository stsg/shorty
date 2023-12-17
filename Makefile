PHONY: all build test clean

build:
	@ echo
	@ echo "Compiling Binary"
	@ echo
	go build -o cmd/shortener/shortener cmd/shortener/*.go

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
