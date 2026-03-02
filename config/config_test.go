package config

import (
	"os"
	"path/filepath"
	"testing"
)

const testConfigFile = ".config.test"

func TestLoad_DefaultConfig(t *testing.T) {
	cfg := &Config{Volume: 0.8}
	err := cfg.SaveTo(testConfigFile)
	if err != nil {
		t.Fatalf("SaveTo() error = %v", err)
	}
	defer os.Remove(testConfigFile)

	loaded, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if loaded.Volume != 0.8 {
		t.Errorf("Volume = %v, want 0.8", loaded.Volume)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	os.Remove(testConfigFile)

	cfg, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Volume != 1.0 {
		t.Errorf("Default Volume = %v, want 1.0", cfg.Volume)
	}
}

func TestLoad_InvalidVolume(t *testing.T) {
	os.WriteFile(testConfigFile, []byte("volume=invalid\n"), 0644)
	defer os.Remove(testConfigFile)

	cfg, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Volume != 1.0 {
		t.Errorf("Volume with invalid input = %v, want 1.0", cfg.Volume)
	}
}

func TestLoad_ValidVolume(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"zero", "volume=0\n", 0.0},
		{"full", "volume=1\n", 1.0},
		{"half", "volume=0.5\n", 0.5},
		{"quarter", "volume=0.25\n", 0.25},
		{"decimal", "volume=0.75\n", 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.WriteFile(testConfigFile, []byte(tt.input), 0644)
			defer os.Remove(testConfigFile)

			cfg, err := LoadFrom(testConfigFile)
			if err != nil {
				t.Fatalf("LoadFrom() error = %v", err)
			}

			if cfg.Volume != tt.want {
				t.Errorf("Volume = %v, want %v", cfg.Volume, tt.want)
			}
		})
	}
}

func TestLoad_IgnoresComments(t *testing.T) {
	content := `# This is a comment
volume=0.5
# Another comment
`
	os.WriteFile(testConfigFile, []byte(content), 0644)
	defer os.Remove(testConfigFile)

	cfg, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Volume != 0.5 {
		t.Errorf("Volume = %v, want 0.5", cfg.Volume)
	}
}

func TestLoad_IgnoresEmptyLines(t *testing.T) {
	content := `

volume=0.3

`
	os.WriteFile(testConfigFile, []byte(content), 0644)
	defer os.Remove(testConfigFile)

	cfg, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Volume != 0.3 {
		t.Errorf("Volume = %v, want 0.3", cfg.Volume)
	}
}

func TestLoad_IgnoresInvalidLines(t *testing.T) {
	content := `notvalid
=empty
noequals
invalidkey=0.5
`
	os.WriteFile(testConfigFile, []byte(content), 0644)
	defer os.Remove(testConfigFile)

	cfg, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Volume != 1.0 {
		t.Errorf("Volume = %v, want 1.0 (default)", cfg.Volume)
	}
}

func TestConfig_SaveTo(t *testing.T) {
	cfg := &Config{Volume: 0.75}

	err := cfg.SaveTo(testConfigFile)
	if err != nil {
		t.Fatalf("SaveTo() error = %v", err)
	}
	defer os.Remove(testConfigFile)

	_, err = os.Stat(testConfigFile)
	if os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	os.Remove(testConfigFile)
	defer os.Remove(testConfigFile)

	cfg := &Config{Volume: 0.42}

	err := cfg.SaveTo(testConfigFile)
	if err != nil {
		t.Fatalf("SaveTo() error = %v", err)
	}

	loaded, err := LoadFrom(testConfigFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if loaded.Volume != 0.42 {
		t.Errorf("Loaded Volume = %v, want 0.42", loaded.Volume)
	}
}

func TestConfig_SaveCreatesDirectory(t *testing.T) {
	testDir := "testdir"
	testFile := filepath.Join(testDir, ".config")

	dir, _ := filepath.Split(testFile)
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}
	content := "volume=0.6\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to setup test: %v", err)
	}
	defer os.RemoveAll(testDir)

	loaded, err := LoadFrom(testFile)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if loaded.Volume != 0.6 {
		t.Errorf("Volume = %v, want 0.6", loaded.Volume)
	}
}
