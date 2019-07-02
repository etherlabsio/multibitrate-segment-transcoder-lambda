package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/etherlabsio/hls/pkg/hls"
	"github.com/etherlabsio/pkg/commander"
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

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}

func downlaodUntarFile(src, dst string) error {

	resp, err := http.Get(src)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return Untar(dst, resp.Body)
}

func downloadMediaExecs(tmppath string) (string, error) {

	mediaExecURL := os.Getenv("MEDIA_EXEC_URL")
	if mediaExecURL != "" {
		err := downlaodUntarFile(mediaExecURL, tmppath)
		if err != nil {
			return "", err
		}
	} else {
		return "", errors.New("missing env variable MEDIA_EXEC_URL")
	}

	mediaDepsURL := os.Getenv("MEDIA_DEPS_URL")
	if mediaDepsURL != "" {
		err := downlaodUntarFile(mediaDepsURL, tmppath)
		if err != nil {
			return "", err
		}
		os.Setenv("LD_LIBRARY_PATH", tmppath+"/lib")
	}

	script := tmppath + "/ffmpeg.sh"

	out, err := os.Create(script)
	if err != nil {
		return "", err
	}
	defer out.Close()

	out.Write([]byte("#!/bin/sh\nchmod 755 " + tmppath + "/ffmpeg \nexport LD_LIBRARY_PATH=" + tmppath + "/lib \n" + tmppath + "/ffmpeg \"$@\""))
	if err != nil {
		return "", err
	}

	err = commander.Exec("chmod", "755", script)
	if err != nil {
		return "", err
	}
	return script, nil
}

func main() {

	logger := logutil.NewServerLogger(false)
	tmppath, err := ioutil.TempDir("", "")
	if err != nil {
		logutil.WithError(logger, err).Log("failed to create temporary directory")
		return
	}
	defer os.RemoveAll(tmppath)

	script, err := downloadMediaExecs(tmppath)
	if err != nil {
		logutil.WithError(logger, err).Log("failed to download media transcoder app")
		return
	}

	lambda.Start(func(ctx context.Context, event hls.TranscodeEvent) error {

		key := event.Key
		// the file should be under two levels of directories
		if 4 != len(strings.Split(key, "/")) {
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

		m, err := hls.NewMultirateTranscoder(bucket, event, script)
		defer m.Close()
		if err != nil {
			logutil.WithError(logger, err).Log("failure in creating new multirate transcoder call")
			return nil
		}
		args, err := m.GenerateCommand()
		if err != nil {
			logutil.WithError(logger, err).Log("failed to get command arguments")
		}
		err = commander.Exec(args...)
		if err != nil {
			logutil.WithError(logger, err).Log("failed to transcode")
			return nil
		}

		err = m.Upload()
		if err != nil {
			logutil.WithError(logger, err).Log("failed to upload")
		}

		return nil
	})
}
