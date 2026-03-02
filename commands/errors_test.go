package commands

import (
	"errors"
	"testing"
)

func TestExitError_Error(t *testing.T) {
	err := &ExitError{}
	if err.Error() != "exit" {
		t.Errorf("Error() = %v, want exit", err.Error())
	}
}

func TestIsExitError(t *testing.T) {
	err := &ExitError{}
	if !IsExitError(err) {
		t.Error("IsExitError() = false, want true for ExitError")
	}
}

func TestIsExitError_NonExitError(t *testing.T) {
	err := errors.New("some error")
	if IsExitError(err) {
		t.Error("IsExitError() = true, want false for non-ExitError")
	}
}

func TestIsExitError_Nil(t *testing.T) {
	var err error
	if IsExitError(err) {
		t.Error("IsExitError() = true, want false for nil")
	}
}
