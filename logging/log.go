// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Changes from original
// - No more use of kingpin
// - No more Error Log writer
// - Extracted output setting from NewLogger
// - Added support for function name as an addition
// - Added support for WithFields
// - General refactoring
// - Added Testing
// - No general log which is not created by NewLogger
// - Added some base

/*
Example --
To log to the base logger
Base().Info("New wallet was created")

To log to a new logger
logger = NewLogger()
logger.Info("New wallet was created")
*/

package logging

import (
	"io"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/algorand/go-algorand/logging/telemetryspec"
)

// Level refers to the log logging level
type Level uint32

// Create a general Base logger
var (
	baseLogger      Logger
	telemetryConfig TelemetryConfig
)

const (
	// Panic Level level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	Panic Level = iota
	// Fatal Level level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	Fatal
	// Error Level level. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	Error
	// Warn Level level. Non-critical entries that deserve eyes.
	Warn
	// Info Level level. General operational entries about what's going on inside the
	// application.
	Info
	// Debug Level level. Usually only enabled when debugging. Very verbose logging.
	Debug
)

const stackPrefix = "[Stack]"

var once sync.Once

// Init needs to be called to ensure our logging has been initialized
func Init() {
	once.Do(func() {
		// By default, log to stderr (logrus's default), only warnings and above.
		baseLogger = NewLogger()
		baseLogger.SetLevel(Warn)
	})
}

func init() {
	Init()
}

func initializeConfig(cfg TelemetryConfig) {
	telemetryConfig = cfg
}

// Fields maps logrus fields
type Fields = logrus.Fields

// Logger is the interface for loggers.
type Logger interface {
	// Debug logs a message at level Debug.
	Debug(...interface{})
	Debugln(...interface{})
	Debugf(string, ...interface{})

	// Info logs a message at level Info.
	Info(...interface{})
	Infoln(...interface{})
	Infof(string, ...interface{})

	// Warn logs a message at level Warn.
	Warn(...interface{})
	Warnln(...interface{})
	Warnf(string, ...interface{})

	// Error logs a message at level Error.
	Error(...interface{})
	Errorln(...interface{})
	Errorf(string, ...interface{})

	// Fatal logs a message at level Fatal.
	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	// Panic logs a message at level Panic.
	Panic(...interface{})
	Panicln(...interface{})
	Panicf(string, ...interface{})

	// Add one key-value to log
	With(key string, value interface{}) Logger

	// WithFields logs a message with specific fields
	WithFields(Fields) Logger

	// Set the logging level (Info by default)
	SetLevel(Level)

	IsLevelEnabled(level Level) bool

	// Set the telemetry logging level (Info by default)
	SetTelemetryLevel(Level)

	// Sets the output target
	SetOutput(io.Writer)

	// Sets the logger to JSON Format
	SetJSONFormatter()

	// source adds file, line and function fields to the event
	source() *logrus.Entry

	EnableTelemetry(enabled bool)
	GetTelemetryEnabled() bool
	Metrics(category telemetryspec.Category, metrics telemetryspec.MetricDetails, details interface{})
	Event(category telemetryspec.Category, identifier telemetryspec.Event)
	EventWithDetails(category telemetryspec.Category, identifier telemetryspec.Event, details interface{})
	GetTelemetrySession() string
	GetTelemetryHostName() string
	GetInstanceName() string
	GetChainId() string
}

type loggerState struct {
	telemetryEnabled bool
	loggingLevel     Level
	telemetryLevel   Level
}

type logger struct {
	entry       *logrus.Entry
	loggerState *loggerState
}

func (l logger) With(key string, value interface{}) Logger {
	return logger{
		l.entry.WithField(key, value),
		l.loggerState,
	}
}

func (l logger) Debug(args ...interface{}) {
	if l.loggerState.loggingLevel >= Debug {
		l.source().Debug(args...)
	}
}

func (l logger) Debugln(args ...interface{}) {
	if l.loggerState.loggingLevel >= Debug {
		l.source().Debugln(args...)
	}
}

func (l logger) Debugf(format string, args ...interface{}) {
	if l.loggerState.loggingLevel >= Debug {
		l.source().Debugf(format, args...)
	}
}

func (l logger) Info(args ...interface{}) {
	if l.loggerState.loggingLevel >= Info {
		l.source().Info(args...)
	}
}

func (l logger) Infoln(args ...interface{}) {
	if l.loggerState.loggingLevel >= Info {
		l.source().Infoln(args...)
	}
}

func (l logger) Infof(format string, args ...interface{}) {
	if l.loggerState.loggingLevel >= Info {
		l.source().Infof(format, args...)
	}
}

func (l logger) Warn(args ...interface{}) {
	if l.loggerState.loggingLevel >= Warn {
		l.source().Warn(args...)
	}
}

func (l logger) Warnln(args ...interface{}) {
	if l.loggerState.loggingLevel >= Warn {
		l.source().Warnln(args...)
	}
}

func (l logger) Warnf(format string, args ...interface{}) {
	if l.loggerState.loggingLevel >= Warn {
		l.source().Warnf(format, args...)
	}
}

func (l logger) Error(args ...interface{}) {
	if l.loggerState.loggingLevel >= Error {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Error(args...)
	}
}

func (l logger) Errorln(args ...interface{}) {
	if l.loggerState.loggingLevel >= Error {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Errorln(args...)
	}
}

func (l logger) Errorf(format string, args ...interface{}) {
	if l.loggerState.loggingLevel >= Error {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Errorf(format, args...)
	}
}

func (l logger) Fatal(args ...interface{}) {
	if l.loggerState.loggingLevel >= Fatal {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Fatal(args...)
	}
}

func (l logger) Fatalln(args ...interface{}) {
	if l.loggerState.loggingLevel >= Fatal {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Fatalln(args...)
	}
}

func (l logger) Fatalf(format string, args ...interface{}) {
	if l.loggerState.loggingLevel >= Fatal {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Fatalf(format, args...)
	}
}

func (l logger) Panic(args ...interface{}) {
	if l.loggerState.loggingLevel >= Panic {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Panic(args...)
	}
}

func (l logger) Panicln(args ...interface{}) {
	if l.loggerState.loggingLevel >= Panic {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Panicln(args...)
	}
}

func (l logger) Panicf(format string, args ...interface{}) {
	if l.loggerState.loggingLevel >= Panic {
		l.source().Errorln(stackPrefix, string(debug.Stack()))
		l.source().Panicf(format, args...)
	}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{
		l.source().WithFields(fields),
		l.loggerState,
	}
}

func (l logger) SetLevel(lvl Level) {
	l.loggerState.loggingLevel = lvl
}

func (l logger) IsLevelEnabled(level Level) bool {
	return l.loggerState.loggingLevel >= level
}

func (l logger) SetTelemetryLevel(lvl Level) {
	l.loggerState.telemetryLevel = lvl
}

func (l logger) SetOutput(w io.Writer) {
	l.entry.Logger.Out = w
}

func (l logger) SetJSONFormatter() {
	l.entry.Logger.Formatter = &logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000000Z07:00"}
}

func (l logger) source() *logrus.Entry {
	event := l.entry

	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		// Add file name and number
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
		event = event.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		})

		// Add function name if possible
		if function := runtime.FuncForPC(pc); function != nil {
			event = event.WithField("function", function.Name())
		}
	}
	return event
}

// Base returns the default Logger logging to
func Base() Logger {
	return baseLogger
}

// NewLogger returns a new Logger logging to out.
func NewLogger() Logger {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	out := logger{
		logrus.NewEntry(l),
		&loggerState{
			telemetryEnabled: false,
			loggingLevel:     Info,
			telemetryLevel:   Info,
		},
	}
	formatter := out.entry.Logger.Formatter
	tf, ok := formatter.(*logrus.TextFormatter)
	if ok {
		tf.TimestampFormat = "2006-01-02T15:04:05.000000 -0700"
	}
	return out
}

func (l logger) EnableTelemetry(enabled bool) {
	l.loggerState.telemetryEnabled = enabled
}

func (l logger) GetTelemetryEnabled() bool {
	return l.loggerState.telemetryEnabled
}

func (l logger) GetTelemetrySession() string {
	return telemetryConfig.SessionGUID
}

func (l logger) GetTelemetryHostName() string {
	return telemetryConfig.getHostName()
}

func (l logger) GetInstanceName() string {
	return telemetryConfig.getInstanceName()
}

func (l logger) GetChainId() string {
	return telemetryConfig.ChainID
}

func (l logger) Metrics(category telemetryspec.Category, metrics telemetryspec.MetricDetails, details interface{}) {
	if l.loggerState.telemetryEnabled {
		logMetrics(l, category, metrics, details)
	}
}

func (l logger) Event(category telemetryspec.Category, identifier telemetryspec.Event) {
	l.EventWithDetails(category, identifier, nil)
}

func (l logger) EventWithDetails(category telemetryspec.Category, identifier telemetryspec.Event, details interface{}) {
	if l.loggerState.telemetryEnabled {
		logEvent(l, category, identifier, details)
	}
}
