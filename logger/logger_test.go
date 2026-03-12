package logger

import (
	"strings"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected [INFO] in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected 'test message' in output, got: %s", output)
	}
}

func TestLogger_Debug(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected [DEBUG] in output, got: %s", output)
	}
}

func TestLogger_Warn(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l.Warn("warn message")

	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected [WARN] in output, got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got: %s", output)
	}
}

func TestLogger_LogFormat(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l.Info("test message")

	output := buf.String()

	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("expected bracketed format, got: %s", output)
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestLogger_WithProfile(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l = l.WithProfile("perfil1")
	l.Info("test")

	output := buf.String()
	if !strings.Contains(output, "[perfil1]") {
		t.Errorf("expected [perfil1] in output, got: %s", output)
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("warn"))

	l.Debug("debug should not appear")
	l.Info("info should not appear")
	l.Warn("warn should appear")
	l.Error("error should appear")

	output := buf.String()

	if strings.Contains(output, "debug should not appear") {
		t.Errorf("debug should be filtered, got: %s", output)
	}
	if strings.Contains(output, "info should not appear") {
		t.Errorf("info should be filtered, got: %s", output)
	}
	if !strings.Contains(output, "warn should appear") {
		t.Errorf("warn should appear, got: %s", output)
	}
	if !strings.Contains(output, "error should appear") {
		t.Errorf("error should appear, got: %s", output)
	}
}

func TestLogger_ConcurrentAccess(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(i int) {
			l.Info("message %d", i)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 100 {
		t.Errorf("expected 100 lines, got %d", len(lines))
	}
}

func TestLogger_DefaultLevel(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf))

	l.Debug("debug")
	l.Info("info")

	output := buf.String()

	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected default level DEBUG, got: %s", output)
	}
}

func TestLogger_KVFormatting(t *testing.T) {
	buf := &strings.Builder{}
	l := New(WithOutput(buf), WithLevel("debug"))

	l.Info("user action", "user_id", 123, "action", "login")

	output := buf.String()
	if !strings.Contains(output, "user_id=123") {
		t.Errorf("expected key=value format, got: %s", output)
	}
	if !strings.Contains(output, "action=login") {
		t.Errorf("expected key=value format, got: %s", output)
	}
}
