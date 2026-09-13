package logger

import (
	"testing"
)

func TestNewDoesNotReturnNil(t *testing.T) {
	if New(false) == nil {
		t.Error("New(false) returned nil")
	}
	if New(true) == nil {
		t.Error("New(true) returned nil")
	}
}
