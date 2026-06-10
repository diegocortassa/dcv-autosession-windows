//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// Config holds all application configuration, combining autosession and log settings.
type Config struct {
	Autosession AutosessionConfig
	Log         LogConfig
}

// AutosessionConfig contains settings for DCV session creation and management.
type AutosessionConfig struct {
	DCVPath      string // Path to the DCV command
	DCVServerLog string // Path to the DCV server log file
	ReapInterval int    // Interval in seconds between idle session reaper checks (0 disables)
	SessionOwner string // Owner of the auto-created session
	SessionID    string // ID of the auto-created session
	TriggerRegex string // Regex to match in the DCV server log for triggering session creation
}

// LogConfig contains settings for logging output and rotation.
type LogConfig struct {
	Level     string // Log level (e.g. "info", "debug", "error")
	Directory string // Directory for log file output
	Rotation  int    // Maximum number of rotated log files to retain
}

func getDefaultDCVServerLog() string {
	amazonLog := `C:\ProgramData\Amazon\dcv\log\server.log`
	if _, err := os.Stat(amazonLog); err == nil {
		return amazonLog
	}
	return `C:\ProgramData\NICE\dcv\log\server.log`
}

func getDefaultDCVPath() string {
	amazonPath := `C:\Program Files\Amazon\DCV\Server\bin\dcv.exe`
	if _, err := os.Stat(amazonPath); err == nil {
		return amazonPath
	}
	return `C:\Program Files\NICE\DCV\Server\bin\dcv.exe`
}

func getExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %w", err)
	}
	return filepath.Dir(ex), nil
}

func loadDefaults() (*Config, error) {
	defaultFile, err := ini.Load(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded default config: %w", err)
	}

	dAutosession := defaultFile.Section("dcv-autosession")
	dLog := defaultFile.Section("log")

	return &Config{
		Autosession: AutosessionConfig{
			DCVPath:      getDefaultDCVPath(),
			DCVServerLog: getDefaultDCVServerLog(),
			ReapInterval: dAutosession.Key("reap_interval").MustInt(60),
			SessionOwner: dAutosession.Key("session_owner").MustString("Administrator"),
			SessionID:    dAutosession.Key("session_id").MustString("autosession"),
			TriggerRegex: dAutosession.Key("trigger_regex").MustString("Received authentication request from client"),
		},
		Log: LogConfig{
			Level:     dLog.Key("level").String(),
			Directory: dLog.Key("directory").String(),
			Rotation:  dLog.Key("rotation").MustInt(2),
		},
	}, nil
}

// LoadConfig reads and parses the configuration file at configPath. If configPath is empty,
// it defaults to "dcv-autosession.conf" alongside the executable. If the file does not exist,
// a default configuration is written and returned.
func LoadConfig(configPath string) (*Config, error) {
	cfg, err := loadDefaults()
	if err != nil {
		return nil, err
	}

	if configPath == "" {
		execPath, err := getExecutablePath()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		configPath = filepath.Join(execPath, "dcv-autosession.conf")
	}

	file, err := ini.Load(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := os.WriteFile(configPath, defaultConfig, 0640); err != nil {
			return nil, fmt.Errorf("failed to write default config file: %w", err)
		}
		log.Infof("Default configuration file %s created successfully.", configPath)
		return cfg, nil
	}

	// Load Autosession section
	autosessionSection := file.Section("dcv-autosession")
	if autosessionSection.HasKey("dcv_path") {
		cfg.Autosession.DCVPath = autosessionSection.Key("dcv_path").String()
	}
	if autosessionSection.HasKey("dcvserver_log") {
		cfg.Autosession.DCVServerLog = autosessionSection.Key("dcvserver_log").String()
	}
	cfg.Autosession.ReapInterval = autosessionSection.Key("reap_interval").MustInt(cfg.Autosession.ReapInterval)
	if autosessionSection.HasKey("session_owner") {
		cfg.Autosession.SessionOwner = autosessionSection.Key("session_owner").String()
	}
	if autosessionSection.HasKey("session_id") {
		cfg.Autosession.SessionID = autosessionSection.Key("session_id").String()
	}
	if autosessionSection.HasKey("trigger_regex") {
		cfg.Autosession.TriggerRegex = autosessionSection.Key("trigger_regex").String()
	}

	// Load Log section
	logSection := file.Section("log")
	if logSection.HasKey("level") {
		cfg.Log.Level = logSection.Key("level").String()
	}
	if logSection.HasKey("directory") {
		cfg.Log.Directory = logSection.Key("directory").String()
	}
	cfg.Log.Rotation = logSection.Key("rotation").MustInt(cfg.Log.Rotation)

	log.Info("Loaded configuration file ", configPath)

	return cfg, nil
}
