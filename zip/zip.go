package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Unzipper interface {
	Unzip(src, dst string) error
}

func New() Unzipper {
	return &unzipperImpl{}
}

type unzipperImpl struct {
}

func (z *unzipperImpl) Unzip(src, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open file; %s; %w", src, err)
	}
	defer reader.Close()

	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file; %s; %w", f.Name, err)
		}
		defer rc.Close()

		path := filepath.Join(dst, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to open file; %s; %w", path, err)
			}
			defer f.Close()

			if _, err := io.Copy(f, rc); err != nil {
				return fmt.Errorf("failed to copy contents; %w", err)
			}
		}
	}

	return nil
}
