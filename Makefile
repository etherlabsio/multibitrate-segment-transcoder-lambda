.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/hls-multirate-transcoder cmd/hls-multirate-transcoder/main.go
	wget https://s3.amazonaws.com/io.etherlabs.test/packages/amazonlinux.tgz && tar -xzf amazonlinux.tgz && rm amazonlinux.tgz

clean:
	rm -rf ./bin/hls-multirate-transcoder

deploy: clean build
	sls deploy --verbose
