package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger holds the logger and metadata
type Logger struct {
	zl   zerolog.Logger
	id   string
	data []map[string]interface{}
	root []map[string]interface{}
}

func init() {
	zerolog.TimestampFieldName = "timestamp"
}

// New prepares and creates a new Logger instance
func New() Logger {
	host, _ := os.Hostname()
	release := os.Getenv("RELEASE")

	zl := zerolog.New(os.Stdout).With().Timestamp().Str("host", host)

	if release != "" {
		zl = zl.Str("release", release)
	}

	return Logger{
		zl:   zl.Logger(),
		data: []map[string]interface{}{},
		root: []map[string]interface{}{},
	}
}

// ID returns a new Logger with the ID set to id
func (log Logger) ID(id string) Logger {
	log.id = id
	return log
}

// Data returns a new logger with the new data appended to the old list of data
func (log Logger) Data(data map[string]interface{}) Logger {
	log.data = append(log.data, data)
	return log
}

// Root returns a new logger with the root info appended to the old list of root
// info. This root info will be displayed at the top level of the log.
func (log Logger) Root(root map[string]interface{}) Logger {
	log.root = append(log.root, root)
	return log
}

// Info outputs a info-level log with a message and any additional data provided
func (log Logger) Info(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Info(), message, fields...)
}

// Error outputs an error-level log with a message and any additional data
// provided
func (log Logger) Error(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Error(), message, fields...)
}

// Warn outputs a warn-level log with a message and any additional data provided
func (log Logger) Warn(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Warn(), message, fields...)
}

// Debug outputs a debug-level log with a message and any additional data
// provided
func (log Logger) Debug(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Debug(), message, fields...)
}

// Fatal outputs a fatal-level log with a message and any additional data
// provided. This will also call os.Exit(1)
func (log Logger) Fatal(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Fatal(), message, fields...)
}

func (log Logger) log(evt *zerolog.Event, message string, fields ...map[string]interface{}) {
	hasData := false
	data := zerolog.Dict()
	for _, field := range append(log.data, fields...) {
		if len(field) != 0 {
			hasData = true
			data = data.Fields(field)
		}
	}

	for _, field := range log.root {
		if len(field) != 0 {
			evt = evt.Fields(field)
		}
	}

	if log.id != "" {
		evt = evt.Str("id", log.id)
	}

	if hasData {
		evt = evt.Dict("data", data)
	}

	evt.Int64("nanoseconds", zerolog.TimestampFunc().UnixNano()).Msg(message)
}
