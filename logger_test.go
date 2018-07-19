package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func testLog(t *testing.T, infoLevel string, infoMsg string) {
	id := "testId"

	data, rootData := map[string]interface{}{"data": "test"}, map[string]interface{}{"r1": "test", "r2": "moreTest"}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unexpected error when creating a pipe: %s", err)
	}
	os.Stdout = w

	log := New().ID(id).Data(data).Root(rootData)

	d1, d2, d3, d4 :=
		map[string]interface{}{"1": "1"},
		map[string]interface{}{"2": 2},
		map[string]interface{}{"3": []int{3, 4, 5}},
		map[string]interface{}{"4": map[string]interface{}{"5": 6.5}}
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

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	if err = w.Close(); err != nil {
		t.Fatalf("unexpected error closing write pipe: %s", err)
	}

	os.Stdout = origStdout

	logLine := <-outC

	if !strings.Contains(logLine, fmt.Sprintf(`"id":"%s"`, id)) {
		t.Error("ID is incorrect")
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
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestLogs(t *testing.T) {
	testLog(t, "info", "info test")
	testLog(t, "error", "error test")
	testLog(t, "debug", "debug test")
	testLog(t, "warn", "warn test")
}
