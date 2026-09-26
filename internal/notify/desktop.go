package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type DesktopNotifier struct {
	config *DesktopConfig
}

type DesktopConfig struct {
	Enabled bool
	Title   string
	Events  map[string]string
}

func NewDesktopNotifier(cfg *DesktopConfig) *DesktopNotifier {
	return &DesktopNotifier{config: cfg}
}

func (d *DesktopNotifier) Notify(event, customTitle, customMessage string) error {
	if !d.config.Enabled {
		return nil
	}

	title := customTitle
	if title == "" {
		title = d.config.Title
	}

	message := customMessage
	if message == "" {
		message = d.config.Events[event]
	}
	if message == "" {
		message = d.config.Events["info"]
	}

	return d.showNotification(title, message)
}

func (d *DesktopNotifier) showNotification(title, message string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = d.linuxNotifyCmd(title, message)
	case "darwin":
		cmd = d.macOSNotifyCmd(title, message)
	case "windows":
		cmd = d.windowsNotifyCmd(title, message)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}

func (d *DesktopNotifier) linuxNotifyCmd(title, message string) *exec.Cmd {
	// Try notify-send (libnotify)
	if _, err := exec.LookPath("notify-send"); err == nil {
		return exec.Command("notify-send", title, message)
	}
	
	// Fallback: try using dbus directly
	return exec.Command("echo", "Desktop notifications not available: install libnotify-bin")
}

func (d *DesktopNotifier) macOSNotifyCmd(title, message string) *exec.Cmd {
	// Escape quotes for osascript
	escapedTitle := strings.ReplaceAll(title, `"`, `\"`)
	escapedMessage := strings.ReplaceAll(message, `"`, `\"`)
	
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escapedMessage, escapedTitle)
	return exec.Command("osascript", "-e", script)
}

func (d *DesktopNotifier) windowsNotifyCmd(title, message string) *exec.Cmd {
	// Try BurntToast module first (modern Windows notifications)
	psScript := fmt.Sprintf(`
		try {
			Import-Module BurntToast -ErrorAction Stop
			New-BurntToastNotification -Text "%s", "%s"
		} catch {
			# Fallback to legacy balloon tip
			$obj = New-Object System.Windows.Forms.NotifyIcon
			$obj.Icon = [System.Drawing.SystemIcons]::Information
			$obj.Visible = $true
			$obj.ShowBalloonTip(5000, "%s", "%s", [System.Windows.Forms.ToolTipIcon]::Info)
		}
	`, 
		strings.ReplaceAll(title, `"`, `\"`),
		strings.ReplaceAll(message, `"`, `\"`),
		strings.ReplaceAll(title, `"`, `\"`),
		strings.ReplaceAll(message, `"`, `\"`),
	)
	return exec.Command("powershell", "-NoProfile", "-Command", psScript)
}