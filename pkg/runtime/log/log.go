package log

import "github.com/sirupsen/logrus"

var logger *logrus.Entry

func init() {
	logger = logrus.NewEntry(logrus.New())
}

func NewLoggerWithField(key, value string) *logrus.Entry {
	return logger.WithField(key, value)
}

func Logger() *logrus.Entry {
	return logger
}
