package logger

// LOGGER is the global logger
var defaultLogger = New()

// Root adds a map to the list of data that will be displayed at the top level
// of the log
func Root(root Data) {
	defaultLogger.root = append(defaultLogger.root, root)
}

// Info writes a info-level log with a message and any additional data provided
func Info(message string, fields ...Data) {
	defaultLogger.log(defaultLogger.zl.Info(), message, fields...)
}

// Error writes an error-level log with a message and any additional data
// provided
func Error(message string, fields ...Data) {
	defaultLogger.log(defaultLogger.zl.Error(), message, fields...)
}

// Warn writes a warn-level log with a message and any additional data provided
func Warn(message string, fields ...Data) {
	defaultLogger.log(defaultLogger.zl.Warn(), message, fields...)
}

// Debug writes a debug-level log with a message and any additional data
// provided
func Debug(message string, fields ...Data) {
	defaultLogger.log(defaultLogger.zl.Debug(), message, fields...)
}

// Fatal writes a fatal-level log with a message and any additional data
// provided. This will also call os.Exit(1)
func Fatal(message string, fields ...Data) {
	defaultLogger.log(defaultLogger.zl.Fatal(), message, fields...)
}
