package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/agent-notify/agent-notify/internal/config"
	"github.com/agent-notify/agent-notify/internal/daemon"
	"github.com/agent-notify/agent-notify/internal/events"
	"github.com/agent-notify/agent-notify/internal/notify"
)

var (
	version = "dev"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Handle subcommands
	switch os.Args[1] {
	case "daemon":
		runDaemonCmd(os.Args[2:])
		return
	case "version", "-v", "--version":
		fmt.Printf("agent-notify %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	case "help", "-h", "--help":
		printUsage()
		return
	}

	// Main notify command
	runNotifyCmd(os.Args[1:])
}

func runNotifyCmd(args []string) {
	fs := flag.NewFlagSet("notify", flag.ExitOnError)

	// User-friendly flags
	ttsMsg := fs.String("t", "", "Text to speak (TTS)")
	ttsMsgLong := fs.String("tts", "", "Text to speak (TTS)")
	soundName := fs.String("s", "", "Sound name to play")
	soundNameLong := fs.String("sound", "", "Sound name to play")
	notifyTitle := fs.String("n", "", "Notification title")
	notifyTitleLong := fs.String("notify", "", "Notification title")
	notifyMsg := fs.String("m", "", "Notification message")
	notifyMsgLong := fs.String("message", "", "Notification message")

	// Common flags
	event := fs.String("event", "", "Predefined event (work_done, attention_needed, error, warning, info, custom)")
	methods := fs.String("method", "", "Comma-separated methods: sound,tts,desktop")
	volume := fs.Float64("volume", 0, "Volume 0.0-1.0 (0 = use config)")
	voice := fs.String("voice", "", "TTS voice override")
	noDaemon := fs.Bool("no-daemon", false, "Force direct notification (skip daemon)")
	daemonStatus := fs.Bool("daemon-status", false, "Check daemon status")
	assetsDir := fs.String("assets", "", "Assets directory (for bundled sounds)")

	fs.Parse(args)

	// Handle daemon status check
	if *daemonStatus {
		checkDaemonStatus()
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	// Determine assets directory
	assets := *assetsDir
	if assets == "" {
		// Try to find bundled assets
		exe, _ := os.Executable()
		assets = filepath.Join(filepath.Dir(exe), "assets")
		if _, err := os.Stat(assets); os.IsNotExist(err) {
			assets = filepath.Join(filepath.Dir(exe), "..", "assets")
		}
	}

	// Create dispatcher
	dispatcher := notify.NewDispatcher(cfg, assets)

	// Determine notification options
	opts := notify.NotifyOptions{}

	// Parse user-friendly flags first
	if *ttsMsg != "" || *ttsMsgLong != "" {
		msg := *ttsMsg
		if msg == "" {
			msg = *ttsMsgLong
		}
		opts.Event = events.EventCustom
		opts.Message = msg
		if opts.Methods == nil {
			opts.Methods = []string{"tts"}
		}
	}

	if *soundName != "" || *soundNameLong != "" {
		name := *soundName
		if name == "" {
			name = *soundNameLong
		}
		opts.Event = events.EventCustom
		opts.SoundFile = name
		if opts.Methods == nil {
			opts.Methods = []string{"sound"}
		}
	}

	if *notifyTitle != "" || *notifyTitleLong != "" || *notifyMsg != "" || *notifyMsgLong != "" {
		title := *notifyTitle
		if title == "" {
			title = *notifyTitleLong
		}
		msg := *notifyMsg
		if msg == "" {
			msg = *notifyMsgLong
		}
		opts.Event = events.EventCustom
		opts.Title = title
		opts.Message = msg
		if opts.Methods == nil {
			opts.Methods = []string{"desktop"}
		}
	}

	// Parse event-based flags (positional or -event)
	if fs.NArg() > 0 && *event == "" {
		*event = fs.Arg(0)
	}
	if *event != "" {
		opts.Event = *event
		if opts.Methods == nil {
			opts.Methods = []string{"sound", "tts", "desktop"}
		}
	}

	// Override methods from flag
	if *methods != "" {
		opts.Methods = strings.Split(*methods, ",")
		for i, m := range opts.Methods {
			opts.Methods[i] = strings.TrimSpace(m)
		}
	}

	// Override volume
	if *volume > 0 {
		opts.Volume = *volume
	}

	// Voice override (would need to pass to TTS notifier)
	_ = voice

	// Validate
	if opts.Event == "" {
		fmt.Fprintf(os.Stderr, "Error: No event or message specified\n")
		printUsage()
		os.Exit(1)
	}

	// Try daemon first (unless --no-daemon)
	if !*noDaemon && cfg.Daemon.Enabled {
		client := daemon.NewClient(&cfg.Daemon)
		if client.IsRunning() {
			err = client.Notify(opts.Event, opts.Message, opts.Title, opts.SoundFile, opts.Methods, opts.Volume)
			if err == nil {
				return
			}
			fmt.Fprintf(os.Stderr, "Daemon notify failed, falling back: %v\n", err)
		}
	}

	// Direct notification
	opts.Async = true
	if err := dispatcher.Notify(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Notification failed: %v\n", err)
		os.Exit(1)
	}
}

func runDaemonCmd(args []string) {
	fs := flag.NewFlagSet("daemon", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: agent-notify daemon <start|stop|restart|status>\n")
		os.Exit(1)
	}

	cfg, _ := config.Load()
	client := daemon.NewClient(&cfg.Daemon)

	switch fs.Arg(0) {
	case "start":
		fmt.Println("Daemon start not implemented in CLI (run agent-notifyd directly)")
	case "stop":
		if client.IsRunning() {
			fmt.Println("Daemon stop not implemented in CLI")
		} else {
			fmt.Println("Daemon not running")
		}
	case "restart":
		fmt.Println("Daemon restart not implemented in CLI")
	case "status":
		if client.IsRunning() {
			fmt.Println("Daemon: running")
		} else {
			fmt.Println("Daemon: stopped")
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown daemon command: %s\n", fs.Arg(0))
		os.Exit(1)
	}
}

func checkDaemonStatus() {
	cfg, _ := config.Load()
	client := daemon.NewClient(&cfg.Daemon)
	if client.IsRunning() {
		fmt.Println("Daemon: running")
	} else {
		fmt.Println("Daemon: stopped")
	}
}

func printUsage() {
	fmt.Print(`agent-notify - Cross-platform notification system for CLI agents

Usage:
  agent-notify [event] [options]
  agent-notify -t "text"           # TTS
  agent-notify -s sound_name       # Play sound
  agent-notify -n "title" -m "msg" # Desktop notification
  agent-notify daemon <cmd>        # Daemon control

Predefined Events:
  work_done         Task completed successfully
  attention_needed  Agent needs human input
  error             Error occurred
  warning           Warning condition
  info              Informational
  custom            Custom notification

Options:
  -t, --tts "text"        Text to speak
  -s, --sound "name"      Sound file name (e.g., success, alert, error)
  -n, --notify "title"    Notification title
  -m, --message "msg"     Notification message
  -event "name"           Predefined event name
  -method "sound,tts"     Comma-separated methods
  -volume 0.5             Volume override (0.0-1.0)
  -voice "name"           TTS voice override
  --no-daemon             Force direct notification
  --daemon-status         Check daemon status
  --assets "/path"        Assets directory

Examples:
  agent-notify work_done
  agent-notify error "Build failed"
  agent-notify -t "Hey there I am User"
  agent-notify -s success
  agent-notify -n "Done" -m "All tasks complete"
  agent-notify -t "Complete" -s success -n "Done" -m "Finished"
  agent-notify --daemon-status
`)
}