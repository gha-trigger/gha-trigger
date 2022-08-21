package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	lmb "github.com/suzuki-shunsuke/gha-dispatcher/pkg/handler/lambda"
	"go.uber.org/zap"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	logCfg := zap.NewProductionConfig()
	logger, _ := logCfg.Build()
	defer logger.Sync()
	logger = logger.With(
		zap.String("program", "gha-dispatcher-lambda"),
		zap.String("program_version", version),
		zap.String("program_sha", commit),
		zap.String("program_built_date", date),
	)
	logger.Info("start the program")
	if err := core(logger); err != nil {
		logger.Fatal("gha-dispatcher failed", zap.Error(err))
	}
}

func core(logger *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler, err := lmb.New()
	if err != nil {
		return err
	}
	logger.Debug("start handler")
	lambda.StartWithContext(ctx, handler.Do)
	return nil
}
