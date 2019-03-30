.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/hls-multirate-transcoder cmd/hls-multirate-transcoder/main.go
	zip -r hls-multirate-transcoder.zip bin
clean:
	rm -rf ./bin/*
	rm -rf hls-multirate-transcoder.zip
