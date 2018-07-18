package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Data maps strings to any type
type Data map[string]interface{}

// Logger holds the logger and metadata
type Logger struct {
	zl   zerolog.Logger
	id   string
	data []Data
	root []Data
}

func init() {
	zerolog.TimestampFieldName = "timestamp"
}

// New prepares and creates a new Logger
func New() Logger {
	host, _ := os.Hostname()
	release := os.Getenv("RELEASE")

	zl := zerolog.New(os.Stdout).With().Timestamp().Str("host", host)

	if release != "" {
		zl = zl.Str("release", release)
	}

	return Logger{
		zl:   zl.Logger(),
		data: []Data{},
		root: []Data{},
	}
}

// ID sets the ID associated with the log
func (log Logger) ID(id string) Logger {
	log.id = id
	return log
}

// Data adds a map to the list of data associated with the log
func (log Logger) Data(data Data) Logger {
	log.data = append(log.data, data)
	return log
}

// Root adds a map to the list of data that will be displayed at the top level
// of the log
func (log Logger) Root(root Data) Logger {
	log.root = append(log.root, root)
	return log
}

// Info outputs a info-level log with a message and any additional data provided
func (log Logger) Info(message string, fields ...Data) {
	log.log(log.zl.Info(), message, fields...)
}

// Error outputs an error-level log with a message and any additional data
// provided
func (log Logger) Error(message string, fields ...Data) {
	log.log(log.zl.Error(), message, fields...)
}

// Warn outputs a warn-level log with a message and any additional data provided
func (log Logger) Warn(message string, fields ...Data) {
	log.log(log.zl.Warn(), message, fields...)
}

// Debug outputs a debug-level log with a message and any additional data
// provided
func (log Logger) Debug(message string, fields ...Data) {
	log.log(log.zl.Debug(), message, fields...)
}

// Fatal outputs a fatal-level log with a message and any additional data
// provided. This will also call os.Exit(1)
func (log Logger) Fatal(message string, fields ...Data) {
	log.log(log.zl.Fatal(), message, fields...)
}

func (log Logger) log(evt *zerolog.Event, message string, fields ...Data) {
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
