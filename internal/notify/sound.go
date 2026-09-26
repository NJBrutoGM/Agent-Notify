package notify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type SoundNotifier struct {
	config *SoundConfig
	assetsDir string
}

type SoundConfig struct {
	Enabled bool
	Volume  float64
	Events  map[string]string
}

func NewSoundNotifier(cfg *SoundConfig, assetsDir string) *SoundNotifier {
	return &SoundNotifier{
		config:    cfg,
		assetsDir: assetsDir,
	}
}

func (s *SoundNotifier) Notify(event, customSound string) error {
	if !s.config.Enabled {
		return nil
	}

	soundFile := customSound
	if soundFile == "" {
		soundFile = s.config.Events[event]
	}
	if soundFile == "" {
		soundFile = s.config.Events["info"]
	}

	path := s.resolveSoundPath(soundFile)
	if path == "" {
		return fmt.Errorf("sound file not found: %s", soundFile)
	}

	return s.playSound(path)
}

func (s *SoundNotifier) resolveSoundPath(name string) string {
	if filepath.IsAbs(name) {
		if _, err := os.Stat(name); err == nil {
			return name
		}
		return ""
	}

	// Check assets dir first (bundled sounds)
	if s.assetsDir != "" {
		path := filepath.Join(s.assetsDir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check config dir
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "agent-notify", "sounds", name)
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Check current directory
	if _, err := os.Stat(name); err == nil {
		return name
	}

	return ""
}

func (s *SoundNotifier) playSound(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = s.linuxPlayCmd(path)
	case "darwin":
		cmd = exec.Command("afplay", path)
	case "windows":
		cmd = s.windowsPlayCmd(path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if s.config.Volume < 1.0 && runtime.GOOS == "linux" {
		// Volume control via pactl/pactl for pulseaudio
	}
	
	return cmd.Run()
}

func (s *SoundNotifier) linuxPlayCmd(path string) *exec.Cmd {
	// Try multiple players in order
	players := [][]string{
		{"paplay", "--volume=" + fmt.Sprintf("%d", int(s.config.Volume*65536)), path},
		{"aplay", path},
		{"ffplay", "-nodisp", "-autoexit", "-volume", fmt.Sprintf("%d", int(s.config.Volume*100)), path},
	}

	for _, args := range players {
		if _, err := exec.LookPath(args[0]); err == nil {
			return exec.Command(args[0], args[1:]...)
		}
	}
	
	// Fallback to aplay
	return exec.Command("aplay", path)
}

func (s *SoundNotifier) windowsPlayCmd(path string) *exec.Cmd {
	// Use PowerShell to play sound asynchronously
	psScript := fmt.Sprintf(`
		$player = New-Object System.Media.SoundPlayer
		$player.SoundLocation = "%s"
		$player.Play()
	`, strings.ReplaceAll(path, `\`, `\\`))
	return exec.Command("powershell", "-NoProfile", "-Command", psScript)
}