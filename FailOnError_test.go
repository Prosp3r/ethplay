package main

import (
	"errors"
	"testing"
)

func TestFailOnError(t *testing.T) {
	err := errors.New("Test error Not nil --")
	if FailOnError(err, "test error") != true {
		t.Error("Error logger resolved to false instead of true")
	}
}
