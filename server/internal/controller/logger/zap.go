package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Env     string
	Service string
	Version string
	Level   string
}

func New(cfg Config) (*zap.Logger, zap.AtomicLevel, error) {
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	if cfg.Level != "" {
		parsed, err := zapcore.ParseLevel(strings.ToLower(cfg.Level))
		if err != nil {
			return nil, level, err
		}
		level.SetLevel(parsed)
	}

	zapConfig := newZapConfig(cfg, level)

	log, err := zapConfig.Build(
		zap.AddCaller(),
	)
	if err != nil {
		return nil, level, err
	}

	log = log.With(
		zap.String("service", cfg.Service),
		zap.String("env", cfg.Env),
		zap.String("version", cfg.Version),
	)

	return log, level, nil
}

func newZapConfig(cfg Config, level zap.AtomicLevel) zap.Config {
	zapConfig := zap.NewProductionConfig()
	if cfg.Env == "local" {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = level
	zapConfig.DisableStacktrace = true
	zapConfig.EncoderConfig.TimeKey = "time"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return zapConfig
}
