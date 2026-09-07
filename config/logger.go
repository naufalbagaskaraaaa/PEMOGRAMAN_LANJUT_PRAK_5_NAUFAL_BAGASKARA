package config

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var appLog = &lumberjack.Logger{
	Filename:   "logs/app.log",
	MaxSize:    10,
	MaxBackups: 5,
	MaxAge:     30,
	Compress:   true,
}

func LoggerWriter() io.Writer {
	return io.MultiWriter(os.Stdout, appLog)
}
