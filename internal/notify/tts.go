package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type TTSNotifier struct {
	config *TTSConfig
}

type TTSConfig struct {
	Enabled bool
	Voice   string
	Rate    float64
	Events  map[string]string
}

func NewTTSNotifier(cfg *TTSConfig) *TTSNotifier {
	return &TTSNotifier{config: cfg}
}

func (t *TTSNotifier) Notify(event, customMessage string) error {
	if !t.config.Enabled {
		return nil
	}

	message := customMessage
	if message == "" {
		message = t.config.Events[event]
	}
	if message == "" {
		message = t.config.Events["info"]
	}

	return t.speak(message)
}

func (t *TTSNotifier) speak(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = t.linuxSpeakCmd(text)
	case "darwin":
		cmd = t.macOSSpeakCmd(text)
	case "windows":
		cmd = t.windowsSpeakCmd(text)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}

func (t *TTSNotifier) linuxSpeakCmd(text string) *exec.Cmd {
	// Try espeak-ng first (best quality), then spd-say, then festival
	if _, err := exec.LookPath("espeak-ng"); err == nil {
		args := []string{"-v", t.getVoice("en")}
		if t.config.Rate != 1.0 {
			args = append(args, "-s", fmt.Sprintf("%d", int(175*t.config.Rate)))
		}
		args = append(args, text)
		return exec.Command("espeak-ng", args...)
	}
	if _, err := exec.LookPath("spd-say"); err == nil {
		args := []string{"-r", fmt.Sprintf("%d", int((t.config.Rate-1.0)*100)), text}
		if t.config.Voice != "" {
			args = append([]string{"-v", t.config.Voice}, args...)
		}
		return exec.Command("spd-say", args...)
	}
	if _, err := exec.LookPath("festival"); err == nil {
		return exec.Command("festival", "--tts", "--pipe", text)
	}
	return exec.Command("echo", "TTS not available: install espeak-ng, spd-say, or festival")
}

func (t *TTSNotifier) macOSSpeakCmd(text string) *exec.Cmd {
	args := []string{text}
	if t.config.Voice != "" {
		args = append([]string{"-v", t.config.Voice}, args...)
	}
	if t.config.Rate != 1.0 {
		args = append([]string{"-r", fmt.Sprintf("%d", int(180*t.config.Rate))}, args...)
	}
	return exec.Command("say", args...)
}

func (t *TTSNotifier) windowsSpeakCmd(text string) *exec.Cmd {
	psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Speech
		$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
		%s
		%s
		$synth.Speak("%s")
	`, 
		t.getVoiceScript(),
		t.getRateScript(),
		strings.ReplaceAll(text, `"`, `\"`),
	)
	return exec.Command("powershell", "-NoProfile", "-Command", psScript)
}

func (t *TTSNotifier) getVoice(lang string) string {
	if t.config.Voice != "" {
		return t.config.Voice
	}
	return lang
}

func (t *TTSNotifier) getVoiceScript() string {
	if t.config.Voice != "" {
		return fmt.Sprintf(`$synth.SelectVoice("%s")`, t.config.Voice)
	}
	return ""
}

func (t *TTSNotifier) getRateScript() string {
	if t.config.Rate != 1.0 {
		return fmt.Sprintf(`$synth.Rate = %d`, int((t.config.Rate-1.0)*10))
	}
	return ""
}