package log

import "github.com/sirupsen/logrus"

var logger *logrus.Entry

func init() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	logger = logrus.NewEntry(log)
}

func NewLoggerWithField(key, value string) *logrus.Entry {
	return logger.WithField(key, value)
}

func Logger() *logrus.Entry {
	return logger
}
