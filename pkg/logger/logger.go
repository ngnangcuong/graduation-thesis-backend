package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Panic(...interface{})
	Fatal(...interface{})

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Panicf(string, ...interface{})
	Fatalf(string, ...interface{})
}

func GetLogger(level, path string) (log Logger, err error) {
	cfg := zap.Config{
		Encoding: "console",                           //encode kiểu json hoặc console
		Level:    zap.NewAtomicLevelAt(zap.InfoLevel), //chọn InfoLevel có thể log ở cả 3 level
		OutputPaths: []string{
			"stderr",
			path,
		},

		EncoderConfig: zapcore.EncoderConfig{ //Cấu hình logging, sẽ không có stacktracekey
			MessageKey:     "message",
			TimeKey:        "time",
			LevelKey:       "level",
			CallerKey:      "caller",
			EncodeCaller:   zapcore.ShortCallerEncoder, //Lấy dòng code bắt đầu log
			EncodeLevel:    CustomLevelEncoder,         //Format cách hiển thị level log
			EncodeTime:     SyslogTimeEncoder,          //Format hiển thị thời điểm log
			SkipLineEnding: false,
			LineEnding:     "\n",
		},
	}

	cfg.Level.SetLevel(getLevel(level))

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar, nil
}

func getLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "PANIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}
