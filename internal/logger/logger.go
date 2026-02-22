package logger

import (
	"os"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger instance.
// Accessible anywhere in the app via logger.Log.
var Log *zap.Logger

// Init builds and sets the global logger based on the environment.
// Call this once at the top of main() before anything else starts.
//
// Development → human readable, colorized, debug level
// Production  → JSON, info level, no colors

func Init(env string) {
	var err error

	if env == "production" {
		Log, err = buildProductionLogger()
	} else {
		Log, err = buildDevelopmentLogger()
	}

	if err != nil {
		// Can't use the logger to log this — fall back to stderr
		os.Stderr.WriteString("failed to initialise logger: " + err.Error())
		os.Exit(1)
	}

	// Replace the global zap logger too so any third-party
	// packages that use zap.L() also use our configured logger
	zap.ReplaceGlobals(Log)
}

// buildProductionLogger returns a JSON logger at Info level.
// Timestamps are ISO8601, caller info included for traceability.
func buildProductionLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder // human readable time in JSON
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg), // JSON format
		zapcore.AddSync(os.Stdout),  // write to stdout (Docker captures this)
		zapcore.InfoLevel,           // Info and above only
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// buildDevelopmentLogger returns a colorized console logger at Debug level.
func buildDevelopmentLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	cfg.ConsoleSeparator = " | "

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(colorWriter()), // ← force colors even in Docker
		zapcore.DebugLevel,
	)

	return zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(),
	), nil
}

// Sync flushes any buffered log entries.
// Always call this on app shutdown — defer logger.Sync() in main().
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// ---- Shorthand helpers -------------------------------------------------------
// These let you log from anywhere without importing zap directly.
// Usage: logger.Info("user signed up", zap.String("email", email))

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// colorWriter forces ANSI color support —
// needed for Docker terminals that strip colors by default
func colorWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(colorable.NewColorableStdout())
}
