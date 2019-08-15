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
	"github.com/sirupsen/logrus"
)

// TelemetryConfig represents the configuration of Telemetry logging
type TelemetryConfig struct {
	Enable             bool
	URI                string
	Name               string
	GUID               string
	MinLogLevel        logrus.Level
	ReportHistoryLevel logrus.Level
	LogHistoryDepth    uint
	FilePath           string // Path to file on disk, if any
	ChainID            string `json:"-"`
	SessionGUID        string `json:"-"`
	UserName           string
	Password           string
}
