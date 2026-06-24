package config

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)


var Log *zap.Logger
func InitLogger() {
	logFolder := "logs"
	logFile := filepath.Join(logFolder, "auth-service.log")
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		panic("Failed to initialize system log folder structure: " + err.Error())
	}

	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	logFileStream, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open streaming log output target: " + err.Error())
	}
	fileWriteSyncer := zapcore.AddSync(logFileStream)


	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, fileWriteSyncer, zap.DebugLevel),
		zapcore.NewCore(consoleEncoder, consoleWriteSyncer, zap.DebugLevel),
	)

	Log = zap.New(core, zap.AddCaller())
	defer Log.Sync()
}
