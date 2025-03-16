package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "02-01-2006 15:04:05.000",
		ForceColors:     true,
		FullTimestamp:   true},
	)
}

func Debug(format string, v ...interface{}) {
	if v != nil {
		log.Debugf(format, v)
	} else {
		log.Debug(format)
	}
}

func Info(format string, v ...interface{}) {
	if v != nil {
		log.Infof(format, v)
	} else {
		log.Info(format)
	}
}

func Error(format string, v ...interface{}) {
	if v != nil {
		log.Errorf(format, v)
	} else {
		log.Error(format)
	}
}

func Fatal(format string, v ...interface{}) {
	log.Fatalf(format, v)
}
