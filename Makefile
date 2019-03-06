.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/hls_multirate cmd/hlsmultirate/main.go
clean:
	rm -rf ./bin/hls_multirate

deploy: clean build
	sls deploy --verbose
