//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

//go:build windows

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "DCVAutosession"
const serviceDisplayName = "DCV Autosession"
const serviceDescription = "DCV Autosession service for managing DCV sessions"

type autosessionService struct {
	logger *eventlog.Log
}

func (s *autosessionService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {

	exePath, err := os.Executable()
	if err != nil {
		s.logger.Error(1, fmt.Sprintf("Could not get executable path: %v", err))
	} else {
		exeDir := filepath.Dir(exePath)
		err = os.Chdir(exeDir)
		if err != nil {
			s.logger.Error(1, fmt.Sprintf("Could not change working directory: %v", err))
		} else {
			s.logger.Info(1, fmt.Sprintf("Changed working directory to: %s", exeDir))
		}
	}

	changes <- svc.Status{State: svc.StartPending}

	cfg, rp, sc, err := InitApplication("")
	if err != nil {
		s.logger.Error(1, fmt.Sprintf("Failed to initialize application: %v", err))
		return true, 1
	}

	s.logger.Info(1, "Starting...")

	go rp.StartReaper()
	s.logger.Info(1, fmt.Sprintf("Session reaper started, checking every %d seconds", cfg.Autosession.ReapInterval))

	go sc.Start()
	s.logger.Info(1, fmt.Sprintf("Session creator started, monitoring %s", cfg.Autosession.DCVServerLog))

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			s.logger.Info(1, "Stopping service")
			changes <- svc.Status{State: svc.StopPending}
			return false, 0
		default:
			s.logger.Error(1, fmt.Sprintf("Unexpected control request: %v", c))
		}
	}

	return false, 0
}

func InstallService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	s, err = m.CreateService(serviceName,
		exePath,
		mgr.Config{
			DisplayName: serviceDisplayName,
			Description: serviceDescription,
			StartType:   mgr.StartAutomatic,
		})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	// Set recovery actions
	recoveryActions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 120 * time.Second},
	}
	resetPeriod := uint32(24 * time.Hour / time.Second) // Reset after 24 hours

	if err := s.SetRecoveryActions(recoveryActions, resetPeriod); err != nil {
		return fmt.Errorf("failed to set recovery actions: %w", err)
	}

	// Create event log
	if err := eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		s.Delete()
		return fmt.Errorf("failed to install event logger: %w", err)
	}

	// Start the service after installation
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func UninstallService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found", serviceName)
	}
	defer s.Close()

	// Stop the service if it is running
	status, err := s.Query()
	if err == nil && status.State != svc.Stopped {
		_, err = s.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
		// Wait for the service to actually stop
		for i := 0; i < 30; i++ { // wait up to ~15 seconds
			status, err = s.Query()
			if err != nil {
				break
			}
			if status.State == svc.Stopped {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Remove event log
	if err := eventlog.Remove(serviceName); err != nil {
		log.Warnf("Failed to remove event logger: %v", err)
	}

	// Delete service
	if err := s.Delete(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

func RunService() error {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return fmt.Errorf("failed to open event log: %w", err)
	}
	defer elog.Close()

	return svc.Run(serviceName, &autosessionService{logger: elog})
}

func IsWindowsService() (bool, error) {
	return svc.IsWindowsService()
}
