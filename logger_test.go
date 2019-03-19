package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

type FakeWriter struct{}

func (fl *FakeWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (fl *FakeWriter) Close() error {
	return nil
}

func TestNewWithWriter(t *testing.T) {
	logger := NewWithWriter(&FakeWriter{})

	assert.NotEmpty(t, logger)
	assert.NotEmpty(t, logger.zl)
}

func testLogger(t *testing.T, infoLevel string, infoMsg string, global bool) {
	var id string
	var data Data
	var log Logger
	rootData := Data{"r1": "test", "r2": "moreTest"}

	origStdout := os.Stdout
	defer func() {
		os.Stdout = origStdout
		defaultLogger = New()
	}()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unexpected error when creating a pipe: %s", err)
	}
	os.Stdout = w

	if global {
		defaultLogger = New()
	} else {
		id = "testId"
		data = Data{"data": "test"}
		var e error
		if infoLevel == "error" {
			e = errors.New("pkg error")
		} else {
			e = fmt.Errorf("runtime error")
		}
		log = New().ID(id).Err(e).Data(data).Data(data).Root(rootData).Root(rootData)
	}

	d1, d2, d3, d4 :=
		Data{"1": "1"},
		Data{"2": 2},
		Data{"3": []int{3, 4, 5}},
		Data{"4": Data{"5": 6.5}}

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	if global {
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
	} else {
		switch infoLevel {
		case "error":
			log.Error(infoMsg, d1, d2, d3, d4)
		case "warn":
			log.Warn(infoMsg, d1, d2, d3, d4)
		case "debug":
			log.Debug(infoMsg, d1, d2, d3, d4)
		case "fatal":
			log.Fatal(infoMsg, d1, d2, d3, d4)
		default:
			log.Info(infoMsg, d1, d2, d3, d4)
		}
	}

	if err = w.Close(); err != nil {
		t.Fatalf("unexpected error closing write pipe: %s", err)
	}

	logLine := <-outC
	if global {
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
		} else if !strings.Contains(logLine, `"data":{"1":"1","2":2,"3":[3,4,5],"4":{"5":6.5}}`) {
			t.Error("Data is incorrect")
		}
	} else {
		if !strings.Contains(logLine, fmt.Sprintf(`"id":"%s"`, id)) {
			t.Error("ID is incorrect")
		} else if !strings.Contains(logLine, `"error":{`) {
			t.Error("Error is incorrect")
		} else if !strings.Contains(logLine, fmt.Sprintf(`"level":"%s"`, infoLevel)) {
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
		} else if !strings.Contains(logLine, `"data":{"data":"test","1":"1","2":2,"3":[3,4,5],"4":{"5":6.5}}`) {
			t.Error("Data is incorrect")
		}
	}
}

func TestLogs(t *testing.T) {
	testLogger(t, "info", "info test", false)
	testLogger(t, "error", "error test", false)
	testLogger(t, "debug", "debug test", false)
	testLogger(t, "warn", "warn test", false)
}
