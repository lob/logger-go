package logger

import (
	"testing"
)

func TestGlobalLogs(t *testing.T) {
	testLogger(t, "info", "info test", true)
	testLogger(t, "error", "error test", true)
	testLogger(t, "debug", "debug test", true)
	testLogger(t, "warn", "warn test", true)
}
