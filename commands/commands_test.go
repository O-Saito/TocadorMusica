package commands

import (
	"sort"
	"testing"
)

func TestGet_Registered(t *testing.T) {
	cmd := Get("add")
	if cmd == nil {
		t.Error("Expected add command, got nil")
	}
	if cmd.Name() != "add" {
		t.Errorf("Name() = %v, want add", cmd.Name())
	}
}

func TestGet_Unknown(t *testing.T) {
	cmd := Get("unknowncommand")
	if cmd != nil {
		t.Errorf("Expected nil for unknown command, got %v", cmd)
	}
}

func TestList_ContainsAll(t *testing.T) {
	list := List()

	expected := []string{"add", "volume", "skip", "pause", "resume", "stop", "queue", "quit"}

	if len(list) != len(expected) {
		t.Errorf("List length = %d, want %d", len(list), len(expected))
	}

	names := make(map[string]bool)
	for _, cmd := range list {
		names[cmd.Name()] = true
	}

	for _, name := range expected {
		if !names[name] {
			t.Errorf("Expected command %q not found in list", name)
		}
	}
}

func TestList_Sorted(t *testing.T) {
	list := List()

	names := make([]string, len(list))
	for i, cmd := range list {
		names[i] = cmd.Name()
	}

	if !sort.StringsAreSorted(names) {
		t.Errorf("List is not sorted: %v", names)
	}
}
