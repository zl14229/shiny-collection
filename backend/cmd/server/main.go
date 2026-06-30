package main

import (
	"fmt"
	"os"

	"shiny-collection/internal/config"
	"shiny-collection/internal/router"
	"shiny-collection/pkg/database"
	"shiny-collection/seed"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// load config
	cfg := config.Load()

	// init logger
	logger := initLogger(&cfg.Log)
	defer logger.Sync()

	// init database
	if err := database.Init(&cfg.Database, logger); err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	// seed data
	if err := seed.All(database.GetDB(), logger); err != nil {
		logger.Fatal("failed to seed database", zap.Error(err))
	}

	// setup router
	r := router.Setup(database.GetDB(), logger, cfg.CORS.AllowOrigins)

	// start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("server starting", zap.String("addr", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}

func initLogger(cfg *config.LogConfig) *zap.Logger {
	// lumberjack for log rotation
	writer := &lumberjack.Logger{
		Filename:   cfg.OutputPath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		LocalTime:  true,
	}

	// also write to stdout
	multiWriter := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(writer),
	)

	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		multiWriter,
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
