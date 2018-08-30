package logger

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Logger holds the zerolog logger and metadata.
type Logger struct {
	zl   zerolog.Logger
	id   string
	err  error
	data map[string]interface{}
	root map[string]interface{}
}

const stackSize = 4 << 10 // 4KB

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func init() {
	zerolog.TimestampFieldName = "timestamp"
}

// New prepares and creates a new Logger instance.
func New() Logger {
	host, _ := os.Hostname()
	release := os.Getenv("RELEASE")

	zl := zerolog.New(os.Stdout).With().Timestamp().Str("host", host)

	if release != "" {
		zl = zl.Str("release", release)
	}

	return Logger{
		zl:   zl.Logger(),
		data: map[string]interface{}{},
		root: map[string]interface{}{},
	}
}

// ID returns a new Logger with the ID set to id.
func (log Logger) ID(id string) Logger {
	log.id = id
	return log
}

// Err returns a new Logger with the error set to err.
func (log Logger) Err(err error) Logger {
	log.err = err
	return log
}

// Data returns a new logger with the new data appended to the old list of data.
func (log Logger) Data(data map[string]interface{}) Logger {
	newData := map[string]interface{}{}
	for k, v := range log.data {
		newData[k] = v
	}
	for k, v := range data {
		newData[k] = v
	}
	log.data = newData
	return log
}

// Root returns a new logger with the root info appended to the old list of root
// info. This root info will be displayed at the top level of the log.
func (log Logger) Root(root map[string]interface{}) Logger {
	newRoot := map[string]interface{}{}
	for k, v := range log.root {
		newRoot[k] = v
	}
	for k, v := range root {
		newRoot[k] = v
	}
	log.root = newRoot
	return log
}

// Info outputs an info-level log with a message and any additional data
// provided.
func (log Logger) Info(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Info(), message, fields...)
}

// Error outputs an error-level log with a message and any additional data
// provided.
func (log Logger) Error(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Error(), message, fields...)
}

// Warn outputs a warn-level log with a message and any additional data
// provided.
func (log Logger) Warn(message string, fields ...map[string]interface{}) {
	log.log(log.zl.Warn(), message, fields...)
}

// Debug outputs a debug-level log with a message and any additional data
// provided.
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
	if len(log.data) != 0 {
		hasData = true
	}

	data := zerolog.Dict().Fields(log.data)
	for _, field := range fields {
		if len(field) != 0 {
			hasData = true
			data = data.Fields(field)
		}
	}

	evt.Fields(log.root)

	if log.id != "" {
		evt = evt.Str("id", log.id)
	}

	if hasData {
		evt = evt.Dict("data", data)
	}

	if log.err != nil {
		stack := make([]byte, stackSize)
		// support pkg/errors stackTracer interface
		if err, ok := log.err.(stackTracer); ok {
			st := err.StackTrace()
			stack = []byte(fmt.Sprintf("%+v", st))
		} else {
			_ = runtime.Stack(stack, true)
		}
		f := map[string]interface{}{"message": log.err, "stack": stack}
		evt = evt.Dict("error", zerolog.Dict().Fields(f))
	}

	evt.Int64("nanoseconds", zerolog.TimestampFunc().UnixNano()).Msg(message)
}
