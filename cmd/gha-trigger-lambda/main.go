package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	lmb "github.com/gha-trigger/gha-trigger/pkg/handler/lambda"
	"go.uber.org/zap"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	if err := core(); err != nil {
		os.Exit(1)
	}
}

func core() error {
	logCfg := zap.NewProductionConfig()
	logger, _ := logCfg.Build()
	defer logger.Sync() //nolint:errcheck
	logger = logger.With(
		zap.String("program", "gha-trigger-lambda"),
		zap.String("program_version", version),
		zap.String("program_sha", commit),
		zap.String("program_built_date", date),
	)
	logger.Info("start the program")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler, err := lmb.New(ctx, logger)
	if err != nil {
		logger.Error("initialize a handler", zap.Error(err))
		return err
	}
	logger.Info("start handler")
	lambda.StartWithOptions(handler.Do, lambda.WithContext(ctx))
	return nil
}
