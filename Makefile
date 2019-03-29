.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/hls-multirate-transcoder cmd/hls-multirate-transcoder/main.go
	cd bin && wget https://s3.amazonaws.com/io.etherlabs.artifacts/shared/ffmpeg-4.0.2-amazonlinux.tgz -O ffmpeg.tgz && tar -xzf ffmpeg.tgz && rm ffmpeg.tgz && cd -
	wget https://s3.amazonaws.com/io.etherlabs.artifacts/shared/ffmpeg-libs-amazonlinux.tgz -O libs.tgz && tar -xzf libs.tgz && rm libs.tgz
	cp scripts/ffmpegexec.sh bin/
clean:
	rm -rf ./bin/*
	rm -rf ./lib 
	
deploy: clean build
	sls deploy --verbose
