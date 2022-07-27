package logger

import (
	"log"
	"os"
)

var defaultLogger *log.Logger

func init() {
	defaultLogger = log.New(os.Stderr, "", log.LstdFlags)
}

func Default() *log.Logger {
	return defaultLogger
}
