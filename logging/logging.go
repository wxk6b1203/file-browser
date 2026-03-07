package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DeRuina/timberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogOptions struct {
	Level string   `json:"level,omitempty" yaml:"level,omitempty"`
	Path  []string `json:"path,omitempty" yaml:"path,omitempty"`
}

var (
	loggerMu      sync.Mutex
	activeLogger  *zap.Logger
	activeClosers []io.Closer
)

func InitLogging(opt *LogOptions) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	level := parseLevel(opt)
	paths := normalizePaths(opt)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding

	cores := make([]zapcore.Core, 0, len(paths))
	closers := make([]io.Closer, 0)
	seen := make(map[string]struct{}, len(paths))

	for _, target := range paths {
		normalized := normalizeTarget(target)
		if _, ok := seen[normalized]; ok {
			continue
		}

		syncer, closer := buildSyncer(target)
		if syncer == nil {
			continue
		}
		seen[normalized] = struct{}{}
		if closer != nil {
			closers = append(closers, closer)
		}

		core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), syncer, level)
		cores = append(cores, core)
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stderr), level))
	}

	logger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	if activeLogger != nil {
		_ = activeLogger.Sync()
	}
	for _, c := range activeClosers {
		_ = c.Close()
	}

	activeLogger = logger
	activeClosers = closers
	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger.Named("stdlog"))
}

func parseLevel(opt *LogOptions) zapcore.Level {
	if opt == nil || strings.TrimSpace(opt.Level) == "" {
		return zapcore.InfoLevel
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(opt.Level)))); err != nil {
		return zapcore.InfoLevel
	}
	return level
}

func normalizePaths(opt *LogOptions) []string {
	if opt == nil || len(opt.Path) == 0 {
		return []string{"stdout"}
	}

	result := make([]string, 0, len(opt.Path))
	for _, p := range opt.Path {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}

	if len(result) == 0 {
		return []string{"stdout"}
	}
	return result
}

func buildSyncer(target string) (syncer zapcore.WriteSyncer, closer io.Closer) {
	switch normalizeTarget(target) {
	case "stdout":
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	}

	file := filepath.Clean(strings.TrimSpace(target))
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil
	}

	rotatingWriter := &timberjack.Logger{
		Filename:         file,
		MaxSize:          100,
		MaxAge:           7,
		MaxBackups:       30,
		LocalTime:        true,
		Compression:      "none",
		RotationInterval: 24 * time.Hour,
	}
	return zapcore.AddSync(rotatingWriter), rotatingWriter
}

func normalizeTarget(target string) string {
	trimmed := strings.TrimSpace(target)
	lower := strings.ToLower(trimmed)
	if lower == "stdout" || lower == "stderr" {
		return lower
	}
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(trimmed)
}
