package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func testGlobalLogger(t *testing.T, infoLevel string, infoMsg string) {
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unexpected error when creating a pipe: %s", err)
	}
	os.Stdout = w

	defaultLogger = New()
	rootData := map[string]interface{}{"r1": "test", "r2": "moreTest"}
	Root(rootData)

	d1, d2, d3, d4 :=
		map[string]interface{}{"1": "1"},
		map[string]interface{}{"2": 2},
		map[string]interface{}{"3": []int{3, 4, 5}},
		map[string]interface{}{"4": map[string]interface{}{"5": 6.5}}

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	switch infoLevel {
	case "error":
		Error(infoMsg, d1, d2, d3, d4)
	case "warn":
		Warn(infoMsg, d1, d2, d3, d4)
	case "debug":
		Debug(infoMsg, d1, d2, d3, d4)
	case "fatal":
		Fatal(infoMsg, d1, d2, d3, d4)
	case "info":
		Info(infoMsg, d1, d2, d3, d4)
	}

	if err = w.Close(); err != nil {
		t.Fatalf("unexpected error closing write pipe: %s", err)
	}

	logLine := <-outC

	if !strings.Contains(logLine, fmt.Sprintf(`"level":"%s"`, infoLevel)) {
		t.Error("Log level is incorrect")
	} else if !strings.Contains(logLine, `"host":`) {
		t.Error("Host is missing")
	} else if !strings.Contains(logLine, `"release":`) {
		t.Error("Release is missing")
	} else if !strings.Contains(logLine, `"nanoseconds":`) {
		t.Error("Nanoseconds is missing")
	} else if !strings.Contains(logLine, `"timestamp":`) {
		t.Error("Timestamp is missing")
	} else if !strings.Contains(logLine, `"r1":"test"`) || !strings.Contains(logLine, `"r2":"moreTest"`) {
		t.Error("Root data is incorrect")
	} else if !strings.Contains(logLine, `"data":{"1":"1","2":2,"3":[3,4,5],"4":{"5":6.5}}`) {
		t.Error("Data is incorrect")
	}
}

func TestGlobalLogs(t *testing.T) {
	testGlobalLogger(t, "info", "info test")
	testGlobalLogger(t, "error", "error test")
	testGlobalLogger(t, "debug", "debug test")
	testGlobalLogger(t, "warn", "warn test")
}
