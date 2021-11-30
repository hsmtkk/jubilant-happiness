package env

import (
	"os"

	"go.uber.org/zap"
)

type Printer interface {
	PrintAll()
}

func New(logger *zap.SugaredLogger) Printer {
	return &printerImpl{logger: logger}
}

type printerImpl struct {
	logger *zap.SugaredLogger
}

func (p *printerImpl) PrintAll() {
	for _, e := range os.Environ() {
		p.logger.Info(e)
	}
}
