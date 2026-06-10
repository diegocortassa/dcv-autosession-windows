//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package sessioncreator watches the DCV server log for authentication events
// and automatically creates DCV sessions for authenticated users.
package sessioncreator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/diegocortassa/dcv-autosession-windows/internal/dcv"
	log "github.com/sirupsen/logrus"
)

const pollInterval = 1 * time.Second
const reopenDelay = 5 * time.Second

// SessionCreator watches the DCV server log for authentication requests and
// automatically creates a DCV session when a matching log line is detected.
type SessionCreator struct {
	dcvManager   *dcv.DCVManager
	logPath      string
	sessionOwner string
	sessionID    string
	triggerRegex *regexp.Regexp
}

// NewSessionCreator creates a new SessionCreator that monitors the given log file
// and uses the DCVManager to create sessions on trigger matches.
func NewSessionCreator(dcvManager *dcv.DCVManager, logPath string, sessionOwner string, sessionID string, triggerRegex string) (*SessionCreator, error) {
	re, err := regexp.Compile(triggerRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid trigger_regex: %w", err)
	}
	return &SessionCreator{
		dcvManager:   dcvManager,
		logPath:      logPath,
		sessionOwner: sessionOwner,
		sessionID:    sessionID,
		triggerRegex: re,
	}, nil
}

// Start begins tailing the DCV server log and creating sessions on trigger matches.
// It runs indefinitely, restarting the log tailer on errors.
func (s *SessionCreator) Start() {
	log.Infof("Session creator started, monitoring %s for authentication requests...", s.logPath)

	for {
		s.tail()
		log.Warnf("Log tailer for %s stopped, restarting in %.0f seconds...", s.logPath, reopenDelay.Seconds())
		time.Sleep(reopenDelay)
	}
}

// tail opens the DCV server log, seeks to the end, and polls for new lines.
// It returns on log rotation, truncation, or disappearance so the caller can reopen.
func (s *SessionCreator) tail() {
	file, err := os.Open(s.logPath)
	if err != nil {
		log.Errorf("Failed to open log file %s: %v", s.logPath, err)
		return
	}
	defer file.Close()

	lastFi, err := file.Stat()
	if err != nil {
		log.Errorf("Failed to stat log file %s: %v", s.logPath, err)
		return
	}

	// Seek to end to only read new lines going forward
	_, _ = file.Seek(0, 2)

	reader := bufio.NewReader(file)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		fi, err := os.Stat(s.logPath)
		if err != nil {
			log.Warnf("Log file %s disappeared: %v", s.logPath, err)
			return
		}

		// Detect rotation: file changed or was truncated
		if !os.SameFile(fi, lastFi) || fi.Size() < lastFi.Size() {
			log.Infof("Log file %s rotated or truncated, reopening...", s.logPath)
			return
		}
		lastFi = fi

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			s.processLine(line)
		}
	}
}

// processLine checks if the line matches the trigger regex and, if so, creates
// a DCV session for the configured owner (if one does not already exist).
func (s *SessionCreator) processLine(line string) {
	if !s.triggerRegex.MatchString(line) {
		return
	}

	// Check if a session already exists for the owner
	sessions, err := s.dcvManager.ListSessions()
	if err != nil {
		log.Errorf("Failed to list sessions: %v", err)
		return
	}
	for _, session := range sessions {
		if session.Owner == s.sessionOwner {
			log.Debugf("Session %s already exists for %s, skipping", session.ID, s.sessionOwner)
			return
		}
	}

	log.Infof("Authentication request detected, creating session %s owned by %s...", s.sessionID, s.sessionOwner)
	if err := s.dcvManager.CreateSession(s.sessionOwner, "Console", s.sessionID); err != nil {
		log.Errorf("Failed to create session: %v", err)
		return
	}
	log.Infof("Session %s created for %s", s.sessionID, s.sessionOwner)
}
