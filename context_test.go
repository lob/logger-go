package logger

import (
	"context"
	"testing"
)

func TestWithContext(t *testing.T) {
	id := "1234"
	log := New("").ID(id)

	ctx := log.WithContext(context.Background())

	newLog, ok := ctx.Value(key{}).(Logger)
	if !ok {
		t.Fatal("context value should be a Logger")
	}
	if log.id != newLog.id {
		t.Errorf("got: %s, wanted: %s", newLog.id, log.id)
	}
}

func TestFromContext(t *testing.T) {
	id := "4321"
	log := New("").ID(id)

	ctx := log.WithContext(context.Background())

	newLog := FromContext(ctx)
	if log.id != newLog.id {
		t.Errorf("got: %s, wanted: %s", newLog.id, log.id)
	}

	log = FromContext(context.Background())
	if log.id == id {
		t.Error("expected a new logger")
	}
}
