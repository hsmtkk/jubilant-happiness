package handle

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hsmtkk/jubilant-happiness/s3"
	"github.com/hsmtkk/jubilant-happiness/zip"
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
	//printer := env.New(h.logger)
	//printer.PrintAll()

	h.logger.Infof("version", "version", 1)

	dstBucket := os.Getenv("UNZIPPED_ARTIFACT_BUCKET")
	if dstBucket == "" {
		e := fmt.Errorf("UNZIPPED_ARTIFACT_BUCKET is not set")
		h.logger.Error(e)
		return e
	}

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

	sess := session.Must(session.NewSession())

	downloader := s3.NewDownloader(h.logger, sess)
	downloadedZipPath, err := downloader.Download(bucket, key, zipContentPath+tempZip)
	if err != nil {
		h.logger.Errorw("download failed", "error", err)
		return err
	}

	unzipper := zip.New()
	if err := unzipper.Unzip(downloadedZipPath, unzipContentPath); err != nil {
		h.logger.Errorw("unzip failed", "error", err)
		return err
	}

	uploader := s3.NewUploader(sess, h.logger)
	if err := uploader.Upload(tempUnzipPath, dstBucket); err != nil {
		h.logger.Errorw("upload failed", "error", err)
	}

	h.logger.Info("handler finish")

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
