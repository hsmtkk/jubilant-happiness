package handle

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hsmtkk/jubilant-happiness/s3"
	"go.uber.org/zap"
)

const (
	tempArtifactPath = "/tmp/artifact/"
	tempZipPath      = tempArtifactPath + "zipped/"
	tempUnzipPath    = tempArtifactPath + "unzipped/"
	tempZip          = "temp.zip"
	dirPerm          = 0777
)

type S3EventHandler interface {
	Handle(context.Context, events.S3Event) error
}

func New(logger *zap.SugaredLogger) S3EventHandler {
	return &handlerImpl{logger}
}

type handlerImpl struct {
	logger *zap.SugaredLogger
}

func (h *handlerImpl) Handle(ctx context.Context, evt events.S3Event) error {
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		h.logger.Infow("handle", "AWS Request ID", lc.AwsRequestID)
	}

	bucket := evt.Records[0].S3.Bucket.Name
	key := evt.Records[0].S3.Object.Key

	h.logger.Infof("bucket: %s", bucket)
	h.logger.Infof("key: %s", key)

	zipContentPath, unzipContentPath, err := h.prepareDirectory()
	if err != nil {
		h.logger.Errorw("failed to prepare directory", "error", err)
		return err
	}

	sess, err := session.NewSession()
	if err != nil {
		h.logger.Errorw("failed to create a new session", "error", err)
		return err
	}

	downloader := s3.NewDownloader(h.logger, sess)
	name, err := downloader.Download(bucket, key, zipContentPath+tempZip)
	if err != nil {
		h.logger.Errorw("download failed", "error", err)
		return err
	}

	log.Print(unzipContentPath)
	log.Print(name)

	return nil
}

func (h *handlerImpl) prepareDirectory() (string, string, error) {
	now := strconv.Itoa(int(time.Now().UnixNano()))
	zipContentPath := tempZipPath + now + "/"
	unzipContentPath := tempUnzipPath + now + "/"
	if _, err := os.Stat(tempArtifactPath); err != nil {
		if err := os.RemoveAll(tempArtifactPath); err != nil {
			return "", "", fmt.Errorf("failed to remove directory; %s; %w", tempArtifactPath, err)
		}
	}
	if err := h.createDirectory(zipContentPath); err != nil {
		return "", "", err
	}
	if err := h.createDirectory(unzipContentPath); err != nil {
		return "", "", err
	}
	return zipContentPath, unzipContentPath, nil
}

func (h *handlerImpl) createDirectory(path string) error {
	if err := os.MkdirAll(path, dirPerm); err != nil {
		return fmt.Errorf("failed to create directory; %s; %w", path, err)
	}
	return nil
}
