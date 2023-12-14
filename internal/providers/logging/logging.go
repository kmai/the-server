package logging

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const loggerContextKey = contextKey(iota)

func Init() *logrus.Logger {
	log := logrus.New()

	if strings.ToLower(viper.GetString("log.format")) == "json" {
		log.Formatter = &logrus.JSONFormatter{PrettyPrint: false}
	} else {
		log.Formatter = &logrus.TextFormatter{}
	}

	logLevelString := viper.GetString("log.level")

	if level, err := logrus.ParseLevel(strings.ToLower(logLevelString)); err != nil {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Warnf("error while parsing log level \"%s\": %v", logLevelString, err)
	} else {
		logrus.SetLevel(level)
	}

	return log
}

func SetLoggerToContext(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func GetLoggerFromContext(ctx context.Context) *logrus.Logger {
	if contextValue := ctx.Value(loggerContextKey); contextValue != nil {
		if logger, valid := contextValue.(*logrus.Logger); valid {
			return logger
		}

		logrus.Errorf("logger subsystem is not present in the context")
	}

	logrus.Errorf("couldn't retrieve logger from context. Starting a new one..")

	return &logrus.Logger{}
}

type SimpleLogger interface {
	Println(v ...interface{})
}

type LogrusSimple struct {
	Logger SimpleLogger
}

func (f *LogrusSimple) Println(v ...interface{}) {
	f.Logger.Println(v...)
}
