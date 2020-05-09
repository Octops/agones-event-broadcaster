package log

import "github.com/sirupsen/logrus"

var logger *logrus.Entry

func init() {
	log := logrus.New()
	logger = logrus.NewEntry(log)
}

func NewLoggerWithField(key, value string) *logrus.Entry {
	return logger.WithField(key, value)
}
