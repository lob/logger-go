package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Logger holds the zerolog logger and metadata.
type Logger struct {
	zl   zerolog.Logger
	id   string
	err  error
	data Data
	root Data
}

// Data is a type alias so that it's much more concise to add additional data to
// log lines.
type Data map[string]interface{}

const stackSize = 4 << 10 // 4KB

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type Option func(*zerolog.Context)

func init() {
	zerolog.TimestampFieldName = "timestamp"
}

func WithField(key, value string) Option {
	return func(c *zerolog.Context) {
		c.Str(key, value)
	}
}

// New prepares and creates a new Logger instance.
func New(serviceName string, options ...Option) Logger {
	return NewWithWriter(serviceName, os.Stdout, options...)
}

// NewWithWriter prepares and creates a new Logger instance with a specified writer.
func NewWithWriter(serviceName string, w io.Writer, options ...Option) Logger {
	host, _ := os.Hostname()
	if serviceName == "" {
		serviceName = os.Getenv("SERVICE_NAME")
	}

	zl := zerolog.New(w).With().Timestamp()
	zl = zl.Str("host", host).Str("release", os.Getenv("RELEASE"))
	zl = zl.Str("service", serviceName).Str("name", serviceName)

	// List of DataDog Metadata tags
	var ddtags []string

	// If we are in a container, populate ddtags with the containerId
	if containerId, err := getContainerId(); err == nil {
		ddtags = append(ddtags, fmt.Sprintf("container_id:%s", containerId))
	}

	// Add a zerolog field containing our datadog tags in the format "key1:value1,key2:value2,..."
	zl = zl.Str("ddtags", strings.Join(ddtags[:], ","))

	for _, o := range options {
		o(&zl)
	}

	return Logger{
		zl:   zl.Logger(),
		data: Data{},
		root: Data{},
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
func (log Logger) Data(data Data) Logger {
	newData := Data{}
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
func (log Logger) Root(root Data) Logger {
	newRoot := Data{}
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
func (log Logger) Info(message string, fields ...Data) {
	e := log.zl.Info()
	e.Fields(Data{"status": zerolog.LevelInfoValue})
	log.log(e, message, fields...)
}

// Error outputs an error-level log with a message and any additional data
// provided.
func (log Logger) Error(message string, fields ...Data) {
	e := log.zl.Error()
	e.Fields(Data{"status": zerolog.LevelErrorValue})
	log.log(e, message, fields...)
}

// Warn outputs a warn-level log with a message and any additional data
// provided.
func (log Logger) Warn(message string, fields ...Data) {
	e := log.zl.Warn()
	e.Fields(Data{"status": zerolog.LevelWarnValue})
	log.log(e, message, fields...)
}

// Debug outputs a debug-level log with a message and any additional data
// provided.
func (log Logger) Debug(message string, fields ...Data) {
	e := log.zl.Debug()
	e.Fields(Data{"status": zerolog.LevelDebugValue})
	log.log(e, message, fields...)
}

// Fatal outputs a fatal-level log with a message and any additional data
// provided. This will also call os.Exit(1)
func (log Logger) Fatal(message string, fields ...Data) {
	e := log.zl.Fatal()
	e.Fields(Data{"status": zerolog.LevelFatalValue})
	log.log(e, message, fields...)
}

func (log Logger) log(evt *zerolog.Event, message string, fields ...Data) {
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
		var stack []byte
		// support pkg/errors stackTracer interface
		if err, ok := log.err.(stackTracer); ok {
			st := err.StackTrace()
			stack = []byte(fmt.Sprintf("%+v", st))
		} else {
			stack = make([]byte, stackSize)
			n := runtime.Stack(stack, true)
			stack = stack[:n]
		}
		f := Data{"message": log.err, "stack": stack}
		evt = evt.Dict("error", zerolog.Dict().Fields(f))
	}

	evt.Int64("nanoseconds", zerolog.TimestampFunc().UnixNano()).Msg(message)
}

func getContainerId() (string, error) {
	content, err := ioutil.ReadFile("/proc/1/cpuset")
	if err != nil {
		return "", err
	}

	// The format is /namespace/subNamespace/containerId
	// Split the content of the file
	parts := strings.Split(string(content), "/")

	// Pull the last element form the split
	id := parts[len(parts)-1]

	// Remove whitespace (probably just newlines)
	clean_id := strings.TrimSpace(id)

	return clean_id, nil
}
