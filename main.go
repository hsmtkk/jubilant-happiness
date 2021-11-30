package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hsmtkk/jubilant-happiness/handle"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	handler := handle.New(sugar)
	lambda.Start(handler.Handle)
}
