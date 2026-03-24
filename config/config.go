package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrInvalidVolume        = errors.New("volume must be between 0.0 and 1.0")
	ErrInvalidLogLevel      = errors.New("log_level must be debug, info, warn, or error")
	ErrInvalidSearchResults = errors.New("search_results must be positive")
	ErrInvalidMaxQueueSize  = errors.New("max_queue_size must be positive")
	ErrInvalidSampleRate    = errors.New("sample_rate must be positive")
)

type ProfileConfig struct {
	Volume        float64
	SearchResults int
	AutoPlay      bool
	CustomData    map[string]map[string]string
}

type GlobalConfig struct {
	LogLevel     string
	MaxQueueSize int
	SampleRate   int
	MusicFolders []string
	CustomData   map[string]map[string]string
}

type Config interface {
	GetProfile(profileName string) (GlobalConfig, ProfileConfig)
	SetVolume(profileName string, volume float64)
	SetSearchResults(profileName string, results int)
	SetAutoPlay(profileName string, autoPlay bool)
	GetAutoPlay(profileName string) bool
	GetCustomData(profileName string) (map[string]map[string]string, map[string]map[string]string)
	SetCustomData(profileName string, interfaceName string, data map[string]string)
	Save() error
	Validate() error
}

type config struct {
	mu       sync.RWMutex
	path     string
	global   GlobalConfig
	profiles map[string]ProfileConfig
}

const defaultProfileName = "default"

func Load(path string) (Config, error) {
	cfg := &config{
		path: path,
		global: GlobalConfig{
			LogLevel:     "info",
			MaxQueueSize: 500,
			SampleRate:   44100,
			MusicFolders: []string{},
			CustomData:   make(map[string]map[string]string),
		},
		profiles: make(map[string]ProfileConfig),
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err == nil {
		currentSection := ""
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				currentSection = strings.Trim(line, "[]")
				if _, ok := cfg.profiles[currentSection]; !ok {
					cfg.profiles[currentSection] = ProfileConfig{
						Volume:        0.5,
						SearchResults: 10,
						AutoPlay:      true,
						CustomData:    make(map[string]map[string]string),
					}
				}
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if currentSection == "" {
				switch key {
				case "log_level":
					if value != "" {
						cfg.global.LogLevel = value
					}
				case "max_queue_size":
					if v, err := strconv.Atoi(value); err == nil && v > 0 {
						cfg.global.MaxQueueSize = v
					}
				case "sample_rate":
					if v, err := strconv.Atoi(value); err == nil && v > 0 {
						cfg.global.SampleRate = v
					}
				case "music_folders":
					cfg.global.MusicFolders = strings.Split(value, ",")
					for i, folder := range cfg.global.MusicFolders {
						cfg.global.MusicFolders[i] = strings.TrimSpace(folder)
					}
				default:
					if strings.HasPrefix(key, "interface.") {
						ifaceName := strings.TrimPrefix(key, "interface.")
						if ifaceName != "" {
							if cfg.global.CustomData == nil {
								cfg.global.CustomData = make(map[string]map[string]string)
							}
							cfg.global.CustomData[ifaceName] = parseCustomData(value)
						}
					}
				}
			} else {
				profile := cfg.profiles[currentSection]
				switch key {
				case "volume":
					if v, err := strconv.ParseFloat(value, 64); err == nil {
						profile.Volume = v
					}
				case "search_results":
					if v, err := strconv.Atoi(value); err == nil && v > 0 {
						profile.SearchResults = v
					}
				case "auto_play":
					profile.AutoPlay = value == "true"
				default:
					if strings.HasPrefix(key, "interface.") {
						ifaceName := strings.TrimPrefix(key, "interface.")
						if ifaceName != "" {
							if profile.CustomData == nil {
								profile.CustomData = make(map[string]map[string]string)
							}
							profile.CustomData[ifaceName] = parseCustomData(value)
						}
					}
				}
				cfg.profiles[currentSection] = profile
			}
		}
	}

	if _, ok := cfg.profiles[defaultProfileName]; !ok {
		cfg.profiles[defaultProfileName] = ProfileConfig{
			Volume:        0.5,
			SearchResults: 10,
			AutoPlay:      true,
			CustomData:    make(map[string]map[string]string),
		}
	}

	return cfg, nil
}

func parseCustomData(value string) map[string]string {
	result := make(map[string]string)
	if value == "" {
		return result
	}
	pairs := strings.Split(value, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			if key != "" {
				result[key] = val
			}
		}
	}
	return result
}

func serializeCustomData(data map[string]map[string]string) string {
	var parts []string
	for name, kv := range data {
		var pairs []string
		for k, v := range kv {
			pairs = append(pairs, k+":"+v)
		}
		parts = append(parts, "interface."+name+"="+strings.Join(pairs, ";"))
	}
	return strings.Join(parts, "\n")
}

func (c *config) GetCustomData(profileName string) (map[string]map[string]string, map[string]map[string]string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	profile, ok := c.profiles[profileName]
	if !ok {
		profile = c.profiles[defaultProfileName]
	}

	return c.global.CustomData, profile.CustomData
}

func (c *config) SetCustomData(profileName string, interfaceName string, data map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if profileName == "" {
		profileName = defaultProfileName
	}

	profile, ok := c.profiles[profileName]
	if !ok {
		profile = ProfileConfig{
			Volume:        0.5,
			SearchResults: 10,
			AutoPlay:      true,
			CustomData:    make(map[string]map[string]string),
		}
	}

	if profile.CustomData == nil {
		profile.CustomData = make(map[string]map[string]string)
	}
	profile.CustomData[interfaceName] = data
	c.profiles[profileName] = profile
	c.saveLocked()
}

func (c *config) GetProfile(profileName string) (GlobalConfig, ProfileConfig) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	profile, ok := c.profiles[profileName]
	if !ok {
		profile = c.profiles[defaultProfileName]
	}

	return c.global, profile
}

func (c *config) SetVolume(profileName string, volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.profiles[profileName]; !ok {
		profile := c.profiles[defaultProfileName]
		profile.Volume = volume
		c.profiles[profileName] = profile
	} else {
		profile := c.profiles[profileName]
		profile.Volume = volume
		c.profiles[profileName] = profile
	}

	c.saveLocked()
}

func (c *config) SetSearchResults(profileName string, results int) {
	if results < 1 {
		results = 1
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.profiles[profileName]; !ok {
		profile := c.profiles[defaultProfileName]
		profile.SearchResults = results
		c.profiles[profileName] = profile
	} else {
		profile := c.profiles[profileName]
		profile.SearchResults = results
		c.profiles[profileName] = profile
	}

	c.saveLocked()
}

func (c *config) SetAutoPlay(profileName string, autoPlay bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.profiles[profileName]; !ok {
		profile := c.profiles[defaultProfileName]
		profile.AutoPlay = autoPlay
		c.profiles[profileName] = profile
	} else {
		profile := c.profiles[profileName]
		profile.AutoPlay = autoPlay
		c.profiles[profileName] = profile
	}

	c.saveLocked()
}

func (c *config) GetAutoPlay(profileName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	profile, ok := c.profiles[profileName]
	if !ok {
		profile = c.profiles[defaultProfileName]
	}

	return profile.AutoPlay
}

func (c *config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saveLocked()
}

func (c *config) saveLocked() error {
	var lines []string

	lines = append(lines, fmt.Sprintf("log_level=%s", c.global.LogLevel))
	lines = append(lines, fmt.Sprintf("max_queue_size=%d", c.global.MaxQueueSize))
	lines = append(lines, fmt.Sprintf("sample_rate=%d", c.global.SampleRate))
	if len(c.global.MusicFolders) > 0 {
		lines = append(lines, fmt.Sprintf("music_folders=%s", strings.Join(c.global.MusicFolders, ",")))
	}
	if len(c.global.CustomData) > 0 {
		lines = append(lines, serializeCustomData(c.global.CustomData))
	}

	profileNames := make([]string, 0, len(c.profiles))
	for name := range c.profiles {
		profileNames = append(profileNames, name)
	}
	for _, name := range profileNames {
		profile := c.profiles[name]
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("[%s]", name))
		lines = append(lines, fmt.Sprintf("volume=%.2f", profile.Volume))
		lines = append(lines, fmt.Sprintf("search_results=%d", profile.SearchResults))
		lines = append(lines, fmt.Sprintf("auto_play=%t", profile.AutoPlay))
		if len(profile.CustomData) > 0 {
			lines = append(lines, serializeCustomData(profile.CustomData))
		}
	}

	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return os.WriteFile(c.path, []byte(content), 0644)
}

func (c *config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.global.MaxQueueSize <= 0 {
		return ErrInvalidMaxQueueSize
	}

	if c.global.SampleRate <= 0 {
		return ErrInvalidSampleRate
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.global.LogLevel] {
		return ErrInvalidLogLevel
	}

	for _, profile := range c.profiles {
		if profile.Volume < 0.0 || profile.Volume > 1.0 {
			return ErrInvalidVolume
		}
		if profile.SearchResults <= 0 {
			return ErrInvalidSearchResults
		}
	}

	return nil
}
