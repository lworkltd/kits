package logutil

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/lvhuat/kits/service/profile"
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

	// TODO: add hooks
	return nil
}

func IsMultiLineFormat(fmt string) bool {
	return fmt == "text"
}
