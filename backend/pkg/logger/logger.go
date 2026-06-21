package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.Logger

func Init(mode string) error {
	var cfg zap.Config
	if mode == "debug" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	l, err := cfg.Build()
	if err != nil {
		return err
	}
	global = l
	return nil
}

func Get() *zap.Logger {
	if global == nil {
		global, _ = zap.NewProduction()
	}
	return global
}

func Info(msg string, fields ...zap.Field)  { Get().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { Get().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { Get().Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { Get().Fatal(msg, fields...) }
func With(fields ...zap.Field) *zap.Logger  { return Get().With(fields...) }
func Sync()                                 { _ = Get().Sync() }
