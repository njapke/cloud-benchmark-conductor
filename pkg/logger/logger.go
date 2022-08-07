package logger

import (
	"bufio"
	"io"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return &Logger{Logger: log}
}

func (l *Logger) PrefixedReader(prefix string, reader io.Reader) {
	logLineScanner := bufio.NewScanner(reader)
	for logLineScanner.Scan() {
		l.Infof("%s %s", prefix, logLineScanner.Text())
	}
}
