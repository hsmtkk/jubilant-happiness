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
		h.logger.Infof("AWS Request ID: %s", lc.AwsRequestID)
	}

	bucket := evt.Records[0].S3.Bucket.Name
	key := evt.Records[0].S3.Object.Key

	h.logger.Infof("bucket: %s", bucket)
	h.logger.Infof("key: %s", key)

	if err := h.prepareDirectory(); err != nil {
		h.logger.Errorf("failed to prepare directory; %w", err)
		return err
	}

	sess, err := session.NewSession()
	if err != nil {
		e := fmt.Errorf("failed to create a new session")
		h.logger.Error(e)
		return e
	}

	h.logger.Infof("session: %v", sess)

	return nil
}

func (h *handlerImpl) prepareDirectory() error {
	now := strconv.Itoa(int(time.Now().UnixNano()))
	zipContentPath := tempZipPath + now + "/"
	unzipContentPath := tempUnzipPath + now + "/"
	if _, err := os.Stat(tempArtifactPath); err != nil {
		if err := os.RemoveAll(tempArtifactPath); err != nil {
			return fmt.Errorf("failed to remove directory; %s; %w", tempArtifactPath, err)
		}
	}
	if err := h.createDirectory(zipContentPath); err != nil {
		return err
	}
	if err := h.createDirectory(unzipContentPath); err != nil {
		return err
	}
	return nil
}

func (h *handlerImpl) createDirectory(path string) error {
	if err := os.MkdirAll(path, dirPerm); err != nil {
		return fmt.Errorf("failed to create directory; %s; %w", path, err)
	}
	return nil
}
