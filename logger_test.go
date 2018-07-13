package logger

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

var RELEASE = "test12345"

type logLine struct {
	ID          string `json:"id"`
	Release     string `json:"release"`
	Level       string `json:"level"`
	Host        string `json:"host"`
	Data        Data   `json:"data"`
	Root        string `json:"root"`
	Timestamp   string `json:"timestamp"`
	Nanoseconds int64  `json:"nanoseconds"`
	Message     string `json:"message"`
}

func testLog(t *testing.T, infoLevel string, infoMsg string) {
	id := "testId"

	data, rootData := make(Data), make(Data)
	data["data"], rootData["root"] = "test", "test"

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unexpected error when creating a pipe: %s", err)
	}
	os.Stdout = w

	log := New().ID(id).Data(data).Root(rootData)

	switch infoLevel {
	case "error":
		log.Err(errors.New("test error"))
		log.Error(infoMsg)
	case "warn":
		log.Warn(infoMsg)
	case "debug":
		log.Debug(infoMsg)
	case "fatal":
		log.Fatal(infoMsg)
	default:
		log.Info(infoMsg)
	}

	if err = w.Close(); err != nil {
		t.Fatalf("unexpected error closing write pipe: %s", err)
	}

	os.Stdout = origStdout

	l := logLine{}
	if err = json.NewDecoder(r).Decode(&l); err != nil {
		t.Fatalf("unexpected error when decoding JSON %s", err)
	} else if l.Root == "" {
		t.Error("Root data should be present")
	} else if l.Timestamp == "" {
		t.Error("Timestamp should be present")
	} else if len(l.Data) == 0 {
		t.Error("Data should be present")
	}

	host, err := os.Hostname()
	if err != nil {
		t.Fatalf("unexpected error when getting hostname: %s", err)
	}

	cases := []struct {
		got, want string
	}{
		{l.ID, id},
		{l.Release, RELEASE},
		{l.Level, infoLevel},
		{l.Host, host},
		{l.Message, infoMsg},
	}

	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %s, wanted %s", tc.got, tc.want)
		}
	}
}

func TestLogs(t *testing.T) {
	os.Setenv("RELEASE", RELEASE)
	testLog(t, "info", "info test")
	testLog(t, "error", "error test")
	testLog(t, "debug", "debug test")
	testLog(t, "warn", "warn test")
}
