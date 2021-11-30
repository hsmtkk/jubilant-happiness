package s3

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

type Downloder interface {
	Download(bucket, key, dst string) (string, error)
}

func NewDownloader(logger *zap.SugaredLogger, sess *session.Session) Downloder {
	mgr := s3manager.NewDownloader(sess)
	return &downloaderImpl{logger, mgr}
}

type downloaderImpl struct {
	logger *zap.SugaredLogger
	mgr    *s3manager.Downloader
}

func (d *downloaderImpl) Download(bucket, key, dst string) (string, error) {
	d.logger.Infow("download start", "bucket", bucket, "key", key, "dst", dst)
	file, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("failed to create file; %s; %w", dst, err)
	}
	s3Obj := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	numBytes, err := d.mgr.Download(file, s3Obj)
	if err != nil {
		return "", fmt.Errorf("failed to get object; %s; %s; %w", bucket, key, err)
	}
	name := file.Name()
	d.logger.Infow("download finish", "name", name, "bytes", numBytes)
	return name, nil
}
