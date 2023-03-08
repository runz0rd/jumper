package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Infof(format string, a ...any)
	Warnf(format string, a ...any)
	Errorf(format string, a ...any)
	Debugf(format string, a ...any)
	WriterLevel(l logrus.Level) *io.PipeWriter
}

func Log() Logger {
	return Entry("")
}

func WriterLevel(l logrus.Level) io.Writer {
	return logrus.StandardLogger().WriterLevel(l)
}

func Writer(src string) io.Writer {
	return Entry(src).Writer()
}

func Entry(src string) *logrus.Entry {
	if src != "" {
		return logrus.StandardLogger().WithField("src", src)
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func SetLevel(level logrus.Level) {
	logrus.StandardLogger().SetLevel(level)
}
