# Agent-Notify Specification

## Overview
Cross-platform notification system for CLI agents. Supports sounds, TTS, and desktop notifications via CLI command or background daemon.

## Goals
- **Non-blocking**: Agents notify without breaking flow
- **Cross-platform**: Linux, macOS, Windows
- **Multiple methods**: Sound files, TTS, desktop notifications
- **Flexible integration**: CLI command + optional daemon
- **Configurable**: YAML config + environment variables

## Architecture

### Components
1. **agent-notify** - CLI command for direct notification
2. **agent-notifyd** - Background daemon (optional, lower latency)
3. **Shared library** - Notification backends, config, event types

### Communication
- **CLI**: Direct execution, fire-and-forget
- **Daemon IPC**: Unix socket (Linux/macOS) + Named pipe (Windows) + HTTP (all platforms)

## Event Types (Predefined)
| Event | Description | Default Sound | Default TTS |
|-------|-------------|---------------|-------------|
| `work_done` | Task completed successfully | `success.wav` | "Work completed" |
| `attention_needed` | Agent needs human input | `alert.wav` | "Attention required" |
| `error` | Error occurred | `error.wav` | "Error occurred" |
| `warning` | Warning condition | `warning.wav` | "Warning" |
| `info` | Informational | `info.wav` | "Info" |
| `custom` | User-defined | Configurable | Custom message |

## Configuration

### Config File: `~/.config/agent-notify/config.yaml`
```yaml
# Notification methods to enable
methods:
  sound: true
  tts: true
  desktop: true

# Sound settings
sound:
  enabled: true
  volume: 0.8
  # Custom sounds per event (relative to sounds/ dir or absolute path)
  events:
    work_done: "success.wav"
    attention_needed: "alert.wav"
    error: "error.wav"
    warning: "warning.wav"
    info: "info.wav"

# TTS settings
tts:
  enabled: true
  voice: ""  # Empty = system default
  rate: 1.0  # Speech rate
  events:
    work_done: "Work completed"
    attention_needed: "Attention required"
    error: "Error occurred"
    warning: "Warning"
    info: "Information"

# Desktop notification settings
desktop:
  enabled: true
  title: "Agent Notify"
  events:
    work_done: "Work completed"
    attention_needed: "Attention needed"
    error: "Error"
    warning: "Warning"
    info: "Info"

# Daemon settings
daemon:
  enabled: false
  socket_path: "~/.local/share/agent-notify/agent-notify.sock"
  http_port: 8765
  http_host: "127.0.0.1"
```

### Environment Variables (override config)
```
AGENT_NOTIFY_METHODS=sound,tts,desktop
AGENT_NOTIFY_SOUND_VOLUME=0.8
AGENT_NOTIFY_TTS_VOICE=...
AGENT_NOTIFY_DAEMON_ENABLED=true
AGENT_NOTIFY_DAEMON_SOCKET=...
AGENT_NOTIFY_DAEMON_HTTP_PORT=8765
```

## CLI Usage

### Direct Command (Agent-Friendly)
```bash
# Notify with predefined event
agent-notify work_done
agent-notify attention_needed
agent-notify error "Build failed: missing dependency"
agent-notify warning "Deprecated API usage"
agent-notify info "Starting deployment"

# Custom event
agent-notify custom --sound custom.wav --tts "Custom message" --title "Custom"

# Options
agent-notify work_done --method sound,tts    # Override methods
agent-notify work_done --volume 0.5          # Override volume
agent-notify work_done --no-daemon           # Force direct (skip daemon)
agent-notify --daemon-status                 # Check daemon status
agent-notify --version
```

### User-Friendly Interface (Simple Flags)
```bash
# TTS - speak text
agent-notify -t "Hey there I am User"
agent-notify --tts "Build completed successfully"

# Sound - play named sound
agent-notify -s success
agent-notify --sound alert

# Desktop notification
agent-notify -n "Title" -m "Message body"
agent-notify --notify "Title" --message "Message body"

# Combined
agent-notify -t "Done" -s success -n "Complete" -m "All tasks finished"

# Options
agent-notify -t "Hello" --volume 0.5 --voice "Samantha"
agent-notify -s error --method sound,desktop
```

### Daemon Control
```bash
agent-notifyd start      # Start daemon
agent-notifyd stop       # Stop daemon
agent-notifyd restart    # Restart daemon
agent-notifyd status     # Show daemon status
```

## Daemon API

### Unix Socket / Named Pipe
```
POST /notify
Content-Type: application/json

{
  "event": "work_done",
  "message": "optional override",
  "methods": ["sound", "tts"],
  "volume": 0.8
}
```

### HTTP API
```
POST http://127.0.0.1:8765/notify
Content-Type: application/json

{
  "event": "work_done",
  "message": "optional override",
  "methods": ["sound", "tts"],
  "volume": 0.8
}

GET http://127.0.0.1:8765/health
GET http://127.0.0.1:8765/events
```

## Notification Backends

### Sound
- **Linux**: `aplay`, `paplay`, `ffplay` (fallback chain)
- **macOS**: `afplay`
- **Windows**: `powershell -c "(New-Object Media.SoundPlayer 'path').PlaySync()"`

### TTS
- **Linux**: `espeak-ng`, `festival`, `spd-say` (fallback chain)
- **macOS**: `say`
- **Windows**: `powershell -c "Add-Type -AssemblyName System.Speech; (New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak('text')"`

### Desktop Notifications
- **Linux**: `notify-send` (libnotify)
- **macOS**: `osascript -e 'display notification "msg" with title "title"'`
- **Windows**: `powershell` with `Windows.UI.Notifications` or `BurntToast` fallback

## Default Sounds (bundled)
Located in `assets/sounds/`:
- `success.wav` - Pleasant completion sound
- `alert.wav` - Attention-getting sound
- `error.wav` - Error sound
- `warning.wav` - Warning sound
- `info.wav` - Subtle info sound

## Project Structure
```
agent-notify/
├── cmd/
│   ├── agent-notify/      # CLI command
│   └── agent-notifyd/     # Daemon
├── internal/
│   ├── config/            # Config loading (YAML + env)
│   ├── notify/            # Notification backends
│   │   ├── sound.go
│   │   ├── tts.go
│   │   ├── desktop.go
│   │   └── dispatcher.go  # Routes to enabled backends
│   ├── daemon/            # Daemon logic
│   │   ├── server.go      # HTTP + Unix socket server
│   │   └── client.go      # Client for CLI to talk to daemon
│   └── events/            # Event definitions
├── assets/
│   └── sounds/            # Bundled default sounds
├── configs/
│   └── config.yaml.example
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yaml       # Cross-platform releases
└── README.md
```

## Build & Distribution
- **Go 1.22+**
- **goreleaser** for cross-platform binaries
- **Homebrew tap**, **Scoop bucket**, **AUR package**, **Deb/RPM packages**
- Single binary per platform: `agent-notify`, `agent-notifyd`

## Non-Functional Requirements
- CLI startup < 50ms (no daemon)
- Daemon response < 10ms
- Zero runtime dependencies (static binary)
- Works offline (no network required for direct mode)
- Graceful degradation if backend unavailable