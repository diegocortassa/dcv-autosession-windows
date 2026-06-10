//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/diegocortassa/dcv-autosession-windows/internal/service"
	"github.com/diegocortassa/dcv-autosession-windows/internal/version"

	log "github.com/sirupsen/logrus"
)

func main() {

	showVersion := flag.Bool("version", false, "Show version information")
	configPath := flag.String("conf", "", "Path to configuration file")
	installService := flag.Bool("install", false, "Install Windows service")
	uninstallService := flag.Bool("uninstall", false, "Uninstall Windows service")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// Handle Windows service installation/uninstallation
	if *installService {
		if err := service.InstallService(); err != nil {
			log.Fatalf("Failed to install service: %v", err)
		}
		fmt.Println("Service installed successfully")
		os.Exit(0)
	}
	if *uninstallService {
		if err := service.UninstallService(); err != nil {
			log.Fatalf("Failed to uninstall service: %v", err)
		}
		fmt.Println("Service uninstalled successfully")
		os.Exit(0)
	}

	// Check if running as a service
	isService, err := service.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as service: %v", err)
	}
	if isService {
		if err := service.RunService(); err != nil {
			log.Fatalf("Failed to run as service: %v", err)
		}
		return
	}

	// Initialize application (load config, setup logger, create reaper and session creator)
	_, r, sc, err := service.InitApplication(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Infof("Version: %s", version.String())
	log.Info("Starting...")

	go r.StartReaper()
	go sc.Start()

	select {}
}
