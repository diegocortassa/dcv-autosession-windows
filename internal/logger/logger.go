//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/diegocortassa/dcv-autosession-windows/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SetupLogger configures logrus with the given configuration, including log level,
// rotation via lumberjack, and output to both stdout (when available) and a log file.
func SetupLogger(cfg config.LogConfig) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Parse log level
	logLevel, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Set up log rotation
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Directory, "dcv-autosession.log"),
		MaxSize:    100,          // megabytes
		MaxBackups: cfg.Rotation, // keep at most N rotated files
	}

	// Configure logrus
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if term.IsTerminal(int(os.Stdout.Fd())) {
		logrus.SetOutput(io.MultiWriter(os.Stdout, logFile))
	} else {
		logrus.SetOutput(logFile)
	}

	return nil
}
