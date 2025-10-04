// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hcgp

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

type Logger = *logrus.Logger

type logAdapter struct {
	log Logger
}

func NewLogAdapter(log Logger) hclog.Logger {
	return &logAdapter{log}
}

func (l *logAdapter) Debug(msg string, args ...interface{}) {
	l.log.WithFields(adaptFields(args...)).Debug(msg)
}

func (l *logAdapter) Error(msg string, args ...interface{}) {
	l.log.WithFields(adaptFields(args...)).Error(msg)
}

func (l *logAdapter) Info(msg string, args ...interface{}) {
	l.log.WithFields(adaptFields(args...)).Info(msg)
}

func (l *logAdapter) Trace(msg string, args ...interface{}) {
	l.log.WithFields(adaptFields(args...)).Trace(msg)
}

func (l *logAdapter) Warn(msg string, args ...interface{}) {
	l.log.WithFields(adaptFields(args...)).Warn(msg)
}

func (l *logAdapter) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l *logAdapter) IsDebug() bool {
	switch l.log.Level {
	case logrus.TraceLevel:
		fallthrough
	case logrus.DebugLevel:
		return true
	default:
		return false
	}
}

func (l *logAdapter) IsError() bool {
	switch l.log.Level {
	case logrus.ErrorLevel:
		fallthrough
	default:
		return false
	}
}

func (l *logAdapter) IsInfo() bool {
	switch l.log.Level {
	case logrus.InfoLevel:
		fallthrough
	case logrus.WarnLevel:
		fallthrough
	case logrus.ErrorLevel:
		return true
	default:
		return false
	}
}

func (l *logAdapter) IsTrace() bool {
	switch l.log.Level {
	case logrus.TraceLevel:
		return true
	default:
		return false
	}
}

func (l *logAdapter) IsWarn() bool {
	switch l.log.Level {
	case logrus.WarnLevel:
		fallthrough
	case logrus.InfoLevel:
		fallthrough
	case logrus.TraceLevel:
		fallthrough
	case logrus.DebugLevel:
		return true
	default:
		return false
	}
}

func adaptLevelToLogurs(l hclog.Level) logrus.Level {
	switch l {
	case hclog.Error:
		return logrus.ErrorLevel
	case hclog.Warn:
		return logrus.WarnLevel
	case hclog.Info:
		return logrus.InfoLevel
	case hclog.Debug:
		return logrus.DebugLevel
	case hclog.Trace:
		fallthrough
	default:
		return logrus.TraceLevel
	}
}

func (l *logAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	logArgs := []interface{}{"msg", msg}
	logArgs = append(logArgs, args...)

	l.log.Log(adaptLevelToLogurs(level), logArgs...)
}

func (l *logAdapter) Name() string {
	return ""
}

func (l *logAdapter) Named(_ string) hclog.Logger {
	return l
}

func (l *logAdapter) ResetNamed(_ string) hclog.Logger {
	return l
}

func (l *logAdapter) SetLevel(level hclog.Level) {
	l.log.SetLevel(adaptLevelToLogurs(level))
}

func (l *logAdapter) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(l.log.Writer(), "", 0)
}

func (l *logAdapter) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return l.log.Writer()
}

func (l *logAdapter) With(_ ...interface{}) hclog.Logger {
	return l
}

func (l *logAdapter) GetLevel() hclog.Level {
	switch l.log.Level {
	case logrus.TraceLevel:
		return hclog.Trace
	case logrus.DebugLevel:
		return hclog.Debug
	case logrus.InfoLevel:
		return hclog.Info
	case logrus.WarnLevel:
		return hclog.Warn
	case logrus.ErrorLevel:
		return hclog.Error
	case logrus.FatalLevel:
		return hclog.Error // Fatal is treated as Error in hclog
	case logrus.PanicLevel:
		return hclog.Error // Panic is treated as Error in hclog
	default:
		return hclog.NoLevel
	}
}

func adaptFields(args ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		argsKey := args[i]
		argsVal := args[i+1]

		var key string
		switch k := argsKey.(type) {
		case string:
			key = k
		default:
			key = fmt.Sprintf("%s", k)
		}

		var val string
		switch v := argsVal.(type) {
		case error:
			switch v.(type) { //nolint: errorlint
			case json.Marshaler, encoding.TextMarshaler:
			default:
				val = v.Error()
			}
		case []interface{}:
			val = fmt.Sprintf(v[0].(string), v[1:]...)
		}

		fields[key] = val
	}
	return fields
}
