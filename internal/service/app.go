//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package service handles the service lifecycle and configuration.

package service

import (
	"fmt"

	"github.com/diegocortassa/dcv-autosession-windows/internal/config"
	"github.com/diegocortassa/dcv-autosession-windows/internal/dcv"
	"github.com/diegocortassa/dcv-autosession-windows/internal/logger"
	"github.com/diegocortassa/dcv-autosession-windows/internal/reaper"
	"github.com/diegocortassa/dcv-autosession-windows/internal/sessioncreator"

	log "github.com/sirupsen/logrus"
)

// InitApplication loads config, sets up logging, and creates the reaper and session creator.
func InitApplication(configPath string) (*config.Config, *reaper.Reaper, *sessioncreator.SessionCreator, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := logger.SetupLogger(cfg.Log); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	log.Infof("Loaded configuration file %s", configPath)

	dcvManager := dcv.NewDCVManager(cfg.Autosession.DCVPath)

	r, err := reaper.NewReaper(cfg, dcvManager)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create reaper: %w", err)
	}

	sc, err := sessioncreator.NewSessionCreator(dcvManager, cfg.Autosession.DCVServerLog, cfg.Autosession.SessionOwner, cfg.Autosession.SessionID, cfg.Autosession.TriggerRegex)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create session creator: %w", err)
	}

	return cfg, r, sc, nil
}
