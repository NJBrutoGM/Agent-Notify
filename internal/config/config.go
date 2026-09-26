package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Methods  MethodsConfig  `yaml:"methods"`
	Sound    SoundConfig    `yaml:"sound"`
	TTS      TTSConfig      `yaml:"tts"`
	Desktop  DesktopConfig  `yaml:"desktop"`
	Daemon   DaemonConfig   `yaml:"daemon"`
}

type MethodsConfig struct {
	Sound   bool `yaml:"sound"`
	TTS     bool `yaml:"tts"`
	Desktop bool `yaml:"desktop"`
}

type SoundConfig struct {
	Enabled bool              `yaml:"enabled"`
	Volume  float64           `yaml:"volume"`
	Events  map[string]string `yaml:"events"`
}

type TTSConfig struct {
	Enabled bool              `yaml:"enabled"`
	Voice   string            `yaml:"voice"`
	Rate    float64           `yaml:"rate"`
	Events  map[string]string `yaml:"events"`
}

type DesktopConfig struct {
	Enabled bool              `yaml:"enabled"`
	Title   string            `yaml:"title"`
	Events  map[string]string `yaml:"events"`
}

type DaemonConfig struct {
	Enabled     bool   `yaml:"enabled"`
	SocketPath  string `yaml:"socket_path"`
	HTTPPort    int    `yaml:"http_port"`
	HTTPHost    string `yaml:"http_host"`
}

var defaultConfig = Config{
	Methods: MethodsConfig{
		Sound:   true,
		TTS:     true,
		Desktop: true,
	},
	Sound: SoundConfig{
		Enabled: true,
		Volume:  0.8,
		Events: map[string]string{
			"work_done":         "success.wav",
			"attention_needed":  "alert.wav",
			"error":             "error.wav",
			"warning":           "warning.wav",
			"info":              "info.wav",
		},
	},
	TTS: TTSConfig{
		Enabled: true,
		Voice:   "",
		Rate:    1.0,
		Events: map[string]string{
			"work_done":         "Work completed",
			"attention_needed":  "Attention required",
			"error":             "Error occurred",
			"warning":           "Warning",
			"info":              "Information",
		},
	},
	Desktop: DesktopConfig{
		Enabled: true,
		Title:   "Agent Notify",
		Events: map[string]string{
			"work_done":         "Work completed",
			"attention_needed":  "Attention needed",
			"error":             "Error",
			"warning":           "Warning",
			"info":              "Info",
		},
	},
	Daemon: DaemonConfig{
		Enabled:    false,
		SocketPath: expandPath("~/.local/share/agent-notify/agent-notify.sock"),
		HTTPPort:   8765,
		HTTPHost:   "127.0.0.1",
	},
}

func Load() (*Config, error) {
	cfg := defaultConfig

	configPath := getConfigPath()
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, err
			}
		}
	}

	applyEnvOverrides(&cfg)
	return &cfg, nil
}

func getConfigPath() string {
	if p := os.Getenv("AGENT_NOTIFY_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agent-notify", "config.yaml")
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AGENT_NOTIFY_METHODS"); v != "" {
		methods := strings.Split(v, ",")
		cfg.Methods.Sound = false
		cfg.Methods.TTS = false
		cfg.Methods.Desktop = false
		for _, m := range methods {
			switch strings.TrimSpace(m) {
			case "sound":
				cfg.Methods.Sound = true
			case "tts":
				cfg.Methods.TTS = true
			case "desktop":
				cfg.Methods.Desktop = true
			}
		}
	}
	if v := os.Getenv("AGENT_NOTIFY_SOUND_VOLUME"); v != "" {
		var vol float64
		if _, err := fmt.Sscanf(v, "%f", &vol); err == nil {
			cfg.Sound.Volume = vol
		}
	}
	if v := os.Getenv("AGENT_NOTIFY_TTS_VOICE"); v != "" {
		cfg.TTS.Voice = v
	}
	if v := os.Getenv("AGENT_NOTIFY_TTS_RATE"); v != "" {
		var rate float64
		if _, err := fmt.Sscanf(v, "%f", &rate); err == nil {
			cfg.TTS.Rate = rate
		}
	}
	if v := os.Getenv("AGENT_NOTIFY_DAEMON_ENABLED"); v != "" {
		cfg.Daemon.Enabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("AGENT_NOTIFY_DAEMON_SOCKET"); v != "" {
		cfg.Daemon.SocketPath = expandPath(v)
	}
	if v := os.Getenv("AGENT_NOTIFY_DAEMON_HTTP_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Daemon.HTTPPort = port
		}
	}
	if v := os.Getenv("AGENT_NOTIFY_DAEMON_HTTP_HOST"); v != "" {
		cfg.Daemon.HTTPHost = v
	}
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (c *Config) GetSoundFile(event string) string {
	if f, ok := c.Sound.Events[event]; ok {
		return f
	}
	return c.Sound.Events["info"]
}

func (c *Config) GetTTSMessage(event string) string {
	if m, ok := c.TTS.Events[event]; ok {
		return m
	}
	return c.TTS.Events["info"]
}

func (c *Config) GetDesktopMessage(event string) string {
	if m, ok := c.Desktop.Events[event]; ok {
		return m
	}
	return c.Desktop.Events["info"]
}