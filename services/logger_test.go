package services

import (
	"os"
	"testing"
)

func TestInitLogger(t *testing.T) {
	err := InitLogger()
	if err != nil {
		t.Errorf("InitLogger() error = %v", err)
	}

	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Logs directory was not created")
	}
}

func TestInfo(t *testing.T) {
	InitLogger()
	Info("Test info message")
}

func TestError(t *testing.T) {
	InitLogger()
	Error("Test error message")
}

func TestInfo_Formatted(t *testing.T) {
	InitLogger()
	Info("Test %s %d", "formatted", 123)
}

func TestError_Formatted(t *testing.T) {
	InitLogger()
	Error("Test %s %d", "formatted", 123)
}

func TestInitLogger_Idempotent(t *testing.T) {
	err := InitLogger()
	if err != nil {
		t.Errorf("InitLogger() error = %v", err)
	}

	err = InitLogger()
	if err != nil {
		t.Errorf("InitLogger() second call error = %v", err)
	}
}
