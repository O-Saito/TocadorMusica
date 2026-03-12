package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `log_level=warn
max_queue_size=100
music_folders=C:\Music,D:\Songs

[default]
volume=0.5
search_results=5
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	global, profile := cfg.GetProfile("default")

	if global.LogLevel != "warn" {
		t.Errorf("expected log_level warn, got %s", global.LogLevel)
	}
	if global.MaxQueueSize != 100 {
		t.Errorf("expected max_queue_size 100, got %d", global.MaxQueueSize)
	}
	if len(global.MusicFolders) != 2 {
		t.Errorf("expected 2 music folders, got %d", len(global.MusicFolders))
	}
	if profile.Volume != 0.5 {
		t.Errorf("expected volume 0.5, got %f", profile.Volume)
	}
	if profile.SearchResults != 5 {
		t.Errorf("expected search_results 5, got %d", profile.SearchResults)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, profile := cfg.GetProfile("unknown-profile")

	if profile.Volume != 0.5 {
		t.Errorf("expected default volume 0.5, got %f", profile.Volume)
	}
	if profile.SearchResults != 10 {
		t.Errorf("expected default search_results 10, got %d", profile.SearchResults)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent_config_file")
	if err != nil {
		t.Error("expected no error for nonexistent file, got:", err)
	}
}

func TestConfig_Validate(t *testing.T) {
	content := `log_level=debug
max_queue_size=500

[default]
volume=0.5
search_results=10
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("validation should pass, got: %v", err)
	}
}

func TestConfig_Validate_InvalidVolume(t *testing.T) {
	content := `[default]
volume=1.5
search_results=10
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := cfg.Validate(); err == nil {
		t.Error("validation should fail for volume > 1.0")
	}
}

func TestConfig_Validate_InvalidLogLevel(t *testing.T) {
	content := `log_level=invalid_level

[default]
volume=0.5
search_results=10
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := cfg.Validate(); err == nil {
		t.Error("validation should fail for invalid log_level")
	}
}

func TestConfig_SetVolume(t *testing.T) {
	content := `[default]
volume=0.5
search_results=10
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg.SetVolume("test-profile", 0.8)

	_, profile := cfg.GetProfile("test-profile")
	if profile.Volume != 0.8 {
		t.Errorf("expected volume 0.8, got %f", profile.Volume)
	}
}

func TestConfig_SetSearchResults(t *testing.T) {
	content := `[default]
volume=0.5
search_results=10
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg.SetSearchResults("test-profile", 20)

	_, profile := cfg.GetProfile("test-profile")
	if profile.SearchResults != 20 {
		t.Errorf("expected search_results 20, got %d", profile.SearchResults)
	}
}

func TestConfig_Save(t *testing.T) {
	content := `log_level=debug
max_queue_size=500

[default]
volume=0.5
search_results=10

[main-perfil]
volume=0.8
search_results=5
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg.SetVolume("main-perfil", 0.9)
	cfg.SetSearchResults("new-profile", 15)

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "volume=0.90") {
		t.Error("expected saved volume=0.90")
	}
	if !strings.Contains(string(data), "[new-profile]") {
		t.Error("expected new-profile section")
	}
	if !strings.Contains(string(data), "search_results=15") {
		t.Error("expected search_results=15")
	}
}

func TestConfig_GetProfile_UsesDefault(t *testing.T) {
	content := `[default]
volume=0.3
search_results=7
`
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, profile := cfg.GetProfile("nonexistent-profile")

	if profile.Volume != 0.3 {
		t.Errorf("expected default volume 0.3, got %f", profile.Volume)
	}
	if profile.SearchResults != 7 {
		t.Errorf("expected default search_results 7, got %d", profile.SearchResults)
	}
}
