package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/etherlabsio/hls/pkg/hls"
	"github.com/etherlabsio/pkg/logutil"
	"github.com/go-kit/kit/log"
	"github.com/google/go-cloud/blob/s3blob"
)

func makeAWSSession(l log.Logger) *session.Session {
	logger := aws.LoggerFunc(func(args ...interface{}) {
		l.Log(args...)
	})
	c := &aws.Config{
		Logger: logger,
	}
	s := session.Must(session.NewSession(c))
	return s
}

func main() {

	lambda.Start(func(ctx context.Context, event hls.TranscodeEvent) error {
		logger := logutil.NewServerLogger(false)

		key := event.Key
		// the file should be under two levels of directories
		if 3 != len(strings.Split(key, "/")) {
			logutil.WithError(logger, errors.New("invalid key")).Log("Invalid key", key, "length is", fmt.Sprintf("%d", len(strings.Split(key, "/"))))
			return nil
		}
		// supports processing of ts files only
		if path.Ext(key) != ".ts" {
			logutil.WithError(logger, errors.New("invalid key")).Log("Invalid key", key, "This supports only ts files")
			return nil
		}
		// should not process previous process's output
		if strings.Contains(key, "/720p/") || strings.Contains(key, "/480p/") || strings.Contains(key, "/360p/") {
			logutil.WithError(logger, errors.New("invalid key")).Log("keyword 720p/480p/360p is present in the path")
			return nil
		}
		bucketName := event.Bucket
		sess := makeAWSSession(logger)
		bucket, err := s3blob.OpenBucket(ctx, bucketName, sess, nil)
		if err != nil {
			logutil.WithError(logger, err).Log("failed to open the bucket " + bucketName)
			return nil
		}

		logger.Log("lambda-function", "ether-hls-multirate", "bucket", bucketName, "segment", key)

		m, err := hls.NewMultirateTranscoder(bucket, event, "./bin/ffmpegexec.sh")
		defer m.Close()
		if err != nil {
			logutil.WithError(logger, err).Log("failure in creating new multirate transcoder call")
			return err
		}
		err = m.Transcode()
		if err != nil {
			logutil.WithError(logger, err).Log("transcode call is failed")
		}
		return err
	})
}
