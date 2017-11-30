package log

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/profile"
	"os"
)

const (
	FilelineTag     = "@fileline"
	ServiceTag      = "@service"
	ServiceIdTag    = "@service-id"
	TracingTag      = "@tracing"
	EnvTag          = "@env"
	DirectLoggerTag = "@directLoggerTag"
	ContextTag      = "@contextTempTag"
)

func InitLoggerWithProfile(cfg *profile.Logger) error {
	switch cfg.Format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: cfg.TimeFormat,
		})
		logrus.Debug("Use json format logger")
	case "text", "":
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors:   true,
			TimestampFormat: cfg.TimeFormat,
		})
		logrus.Debug("Use text format logger")
	default:
		return fmt.Errorf("unsupport logrus formatter type %s", cfg.Format)
	}
	if cfg.Level != "" {
		logLevel, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			return fmt.Errorf("cannot parse logger level %s", cfg.Level)
		}
		logrus.SetLevel(logLevel)
	}

	if "" != cfg.LogFilePath {
		file, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0660)
		if nil != err {
			return fmt.Errorf("Open log file failed, err:%v, log file path:%v", err, cfg.LogFilePath)
		}
		logrus.SetOutput(file)
	}

	// TODO: add hooks
	return nil
}

func IsMultiLineFormat(fmt string) bool {
	return fmt == "text"
}
