package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var errInvalidLogFormatFlag = errors.New("--log-format must be either 'json' or 'logfmt'")

type LoggerConfig struct {
	LogLevel  string `default:"info" help:"Used to set the logging level of the application. Valid options are 'debug', 'info', 'warn' or 'error'"`
	LogFormat string `default:"logfmt" help:"Used to set the logging format. Valid options are 'logfmt' or 'json'"`
}

// Validate is a method called automatically by the kong cli framework so this deals with validating our LoggerConfig struct.
func (lc *LoggerConfig) Validate() error {
	if _, err := level.Parse(lc.LogLevel); err != nil {
		return fmt.Errorf("%w: must be 'debug', 'info', 'warn' or 'error'", err)
	}

	if lc.LogFormat != "json" && lc.LogFormat != "logfmt" {
		return errInvalidLogFormatFlag
	}
	return nil
}

// configureLogger returns a go-lit logger which is customizable via the loggerConfig struct.
func configureLogger(loggerConfig LoggerConfig) log.Logger {
	var logger log.Logger
	switch loggerConfig.LogFormat {
	case "logfmt":
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	case "json":
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	}

	logger = level.NewFilter(logger, level.Allow(level.ParseDefault(loggerConfig.LogLevel, level.InfoValue())))
	logger = log.WithPrefix(logger, "caller", log.DefaultCaller)
	logger = log.WithPrefix(logger, "ts", log.DefaultTimestampUTC)
	return logger
}
