package log

import "github.com/sirupsen/logrus"

func NewLoggerWithField(key, value string) *logrus.Entry {
	return logrus.WithField(key, value)
}
