package s3

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

type Uploader interface {
	Upload(src, dst string) error
}

func NewUploader(sess *session.Session, logger *zap.SugaredLogger) Uploader {
	mgr := s3manager.NewUploader(sess)
	return &uploaderImpl{mgr, logger}
}

type uploaderImpl struct {
	mgr    *s3manager.Uploader
	logger *zap.SugaredLogger
}

func (u *uploaderImpl) Upload(src, dst string) error {
	u.logger.Infow("upload start", "src", src, "dst", dst)
	var wg sync.WaitGroup
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		u.logger.Infow("handling file", "path", path)
		wg.Add(1)
		go func() {
			defer wg.Done()
			file, err := os.Open(path)
			if err != nil {
				u.logger.Errorw("failed to open file", "path", path, "error", err)
				return
			}
			defer file.Close()

			key := strings.Replace(file.Name(), src, "", 1)
			upInput := &s3manager.UploadInput{
				Bucket: aws.String(dst),
				Key:    aws.String(key),
				Body:   file,
			}
			if _, err := u.mgr.Upload(upInput); err != nil {
				u.logger.Errorw("failed to upload file", "error", err)
				return
			}
		}()
		return nil
	})
	wg.Wait()
	u.logger.Infow("upload finish")
	return nil
}
