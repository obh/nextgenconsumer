package config

import (
	"context"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is global logger
	Log *zap.Logger

	// timeFormat is custom Time format
	customTimeFormat string

	// onceInit guarantee initialize logger only once
	onceInit sync.Once

	// H is global logger for handler layer
	H *cfLogger
)

//LogConfig ..
type LogConfig struct {
	Debug           bool   `yaml:"debug"`
	LogLevel        int    `yaml:"loglevel"`
	LogVersion      string `yaml:"logversion"`
	Application     string `yaml:"application"`
	StackTraceLevel int    `yaml:"stacktracelevel"`
	HandlerLevel    int    `yaml:"handlerlevel"`
	ServiceLevel    int    `yaml:"servicelevel"`
	RepoLevel       int    `yaml:"repolevel"`
	LogTimeFormat   string `yaml:"logtimeformat"`
}

type RequestHeaders struct {
	XRequestID string `json:"X-Request-Id"`
	AddedOn    time.Time
}

type cfLogger struct {
	K     *zap.Logger
	level zapcore.Level
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(customTimeFormat))
}

func Init(logCfg LogConfig) error {
	var err error

	onceInit.Do(func() {
		ecfg := zapcore.EncoderConfig{
			MessageKey:     "message",
			NameKey:        "layer",
			LevelKey:       "level",
			TimeKey:        "ts",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		}

		cfg := zap.Config{
			Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
			Development: false,
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			Encoding:         "json",
			EncoderConfig:    ecfg,
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		var initialField = zap.Field{
			Key:    "application",
			Type:   zapcore.StringType,
			String: logCfg.Application,
		}
		var versionField = zap.Field{
			Key:    "version",
			Type:   zapcore.StringType,
			String: logCfg.LogVersion,
		}

		stackTraceLevel := zapcore.Level(logCfg.StackTraceLevel)
		Log, err = cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(stackTraceLevel), zap.Fields(initialField), zap.Fields(versionField))

		if err != nil {
			log.Fatalln("Cannot initialize logger")
		}

		defer Log.Sync()
		_ = zap.RedirectStdLog(Log)
		H = &cfLogger{K: Log.Named("HANDLER"), level: zapcore.Level(logCfg.HandlerLevel)}
	})
	return err
}

func addCommonFields(ctx context.Context, f []zap.Field) []zap.Field {
	return append(f)
}

// Field provides a interface to use zap field
func Field(k string, v interface{}) zap.Field {
	return zap.Any(k, v)
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *zap.Logger {
	newLogger := Log
	if ctx != nil {
		if reqHeaders, ok := ctx.Value("RequestHeaders").(RequestHeaders); ok {
			newLogger = newLogger.With(
				zap.String("x-request-id", reqHeaders.XRequestID),
			)
		}

	}
	return newLogger
}

// Info wraps info logger for handler layer
func (c *cfLogger) Info(ctx context.Context, act string, fields ...zap.Field) {
	if zapcore.InfoLevel >= c.level {
		c.K.Info(act, addCommonFields(ctx, fields)...)
	}
}

// Error wraps error logger for handler layer
func (c *cfLogger) Error(ctx context.Context, act string, err error, fields ...zap.Field) {
	if zapcore.ErrorLevel >= c.level {
		fields = append(fields, zap.String("error", err.Error()))
		c.K.Error(act, addCommonFields(ctx, fields)...)
	}
}
