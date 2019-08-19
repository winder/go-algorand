// Copyright (C) 2019 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/logging/telemetryspec"
)

const telemetryPrefix = "/"
const telemetrySeparator = "/"

// ReadTelemetryConfigOrDefault reads telemetry config from file or defaults if no config file found.
func ReadTelemetryConfigOrDefault(dataDir *string, genesisID string) (cfg TelemetryConfig, err error) {
	err = nil
	if dataDir != nil && *dataDir != "" {
		configPath := filepath.Join(*dataDir, TelemetryConfigFilename)
		cfg, err = LoadTelemetryConfig(configPath)
	}
	if err != nil && os.IsNotExist(err) {
		var configPath string
		configPath, err = config.GetConfigFilePath(TelemetryConfigFilename)
		if err != nil {
			cfg = createTelemetryConfig()
			return
		}
		cfg, err = LoadTelemetryConfig(configPath)
	}
	if err != nil {
		cfg = createTelemetryConfig()
		if os.IsNotExist(err) {
			err = nil
		} else {
			return
		}
	}
	ch := config.GetCurrentVersion().Channel
	// Should not happen, but default to "dev" if channel is unspecified.
	if ch == "" {
		ch = "dev"
	}
	cfg.ChainID = fmt.Sprintf("%s-%s", ch, genesisID)
	return cfg, err
}

// EnsureTelemetryConfig creates a new TelemetryConfig structure with a generated GUID and the appropriate Telemetry endpoint
// Err will be non-nil if the file doesn't exist, or if error loading.
// Cfg will always be valid.
func EnsureTelemetryConfig(dataDir *string, genesisID string) (TelemetryConfig, error) {
	cfg, _, err := EnsureTelemetryConfigCreated(dataDir, genesisID)
	return cfg, err
}

// EnsureTelemetryConfigCreated is the same as EnsureTelemetryConfig but it also returns a bool indicating
// whether EnsureTelemetryConfig had to create the config.
func EnsureTelemetryConfigCreated(dataDir *string, genesisID string) (TelemetryConfig, bool, error) {
	configPath := ""
	var cfg TelemetryConfig
	var err error
	if dataDir != nil && *dataDir != "" {
		configPath = filepath.Join(*dataDir, TelemetryConfigFilename)
		cfg, err = LoadTelemetryConfig(configPath)
		if err != nil && os.IsNotExist(err) {
			// if it just didn't exist, try again at the other path
			configPath = ""
		}
	}
	if configPath == "" {
		configPath, err = config.GetConfigFilePath(TelemetryConfigFilename)
		if err != nil {
			cfg := createTelemetryConfig()
			initializeConfig(cfg)
			return cfg, true, err
		}
		cfg, err = LoadTelemetryConfig(configPath)
	}
	created := false
	if err != nil {
		err = nil
		created = true
		cfg = createTelemetryConfig()
		cfg.FilePath = configPath // Initialize our desired cfg.FilePath

		// There was no config file, create it.
		err = cfg.Save(configPath)
	}

	ch := config.GetCurrentVersion().Channel
	// Should not happen, but default to "dev" if channel is unspecified.
	if ch == "" {
		ch = "dev"
	}
	cfg.ChainID = fmt.Sprintf("%s-%s", ch, genesisID)

	initializeConfig(cfg)
	return cfg, created, err
}

func logMetrics(l logger, category telemetryspec.Category, metrics telemetryspec.MetricDetails, details interface{}) {
	if metrics == nil {
		return
	}
	l = l.WithFields(logrus.Fields{
		"metrics": metrics,
	}).(logger)

	logTelemetry(l, buildMessage(string(category), string(metrics.Identifier())), details)
}

func logEvent(l logger, category telemetryspec.Category, identifier telemetryspec.Event, details interface{}) {
	logTelemetry(l, buildMessage(string(category), string(identifier)), details)
}

func buildMessage(args ...string) string {
	message := telemetryPrefix + strings.Join(args, telemetrySeparator)
	return message
}

func logTelemetry(l logger, message string, details interface{}) {
	if details != nil {
		l = l.WithFields(logrus.Fields{
			"details": details,
		}).(logger)
	}

	entry := l.entry.WithFields(Fields{
		"telemetry":    l.GetTelemetryEnabled(),
		"session":      l.GetTelemetrySession(),
		"instanceName": l.GetInstanceName(),
		"chainID":      l.GetChainId(),
	})

	entry.Info(message)
}
