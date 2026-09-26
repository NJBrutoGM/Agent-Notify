# Agent Notify

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://img.shields.io/badge/Go-Reference-blue)](https://pkg.go.dev/github.com/agent-notify/agent-notify)

Cross-platform notification system for CLI agents. Use sounds, TTS, and desktop notifications without breaking your workflow.

## Quick Start

```bash
# Install
go install github.com/agent-notify/agent-notify/cmd/agent-notify@latest
go install github.com/agent-notify/agent-notify/cmd/agent-notifyd@latest

# Notify with TTS
agent-notify -t "Hey there I am User"

# Play a sound
agent-notify -s success

# Desktop notification
agent-notify -n "Done" -m "All tasks finished"

# Predefined event
agent-notify work_done
```

## Usage

**TTS**
```bash
agent-notify -t "Your message here"
agent-notify --tts "Build completed"
```

**Sound**
```bash
agent-notify -s success
agent-notify --sound alert
```

**Desktop**
```bash
agent-notify -n "Title" -m "Message"
agent-notify --notify "Done" --message "All done"
```

**Combined**
```bash
agent-notify -t "Complete" -s success -n "Done" -m "Finished"
```

**Predefined Events**
```bash
agent-notify work_done          # Task completed
agent-notify attention_needed   # Needs human input
agent-notify error "Build failed"
agent-notify warning "Deprecated API"
agent-notify info "Starting..."
```

## Daemon

```bash
agent-notifyd start      # Start background daemon
agent-notifyd status     # Check status
agent-notifyd stop       # Stop daemon
```

Or via CLI:
```bash
agent-notify daemon status
```

## Configuration

Edit `~/.config/agent-notify/config.yaml`:

```yaml
methods:
  sound: true
  tts: true
  desktop: true

sound:
  volume: 0.8
  events:
    work_done: "success.wav"
    error: "error.wav"

tts:
  voice: ""
  rate: 1.0

desktop:
  title: "Agent Notify"
```

Or use environment variables:
```bash
AGENT_NOTIFY_METHODS=sound,tts
AGENT_NOTIFY_SOUND_VOLUME=0.5
```

## Daemon API

```bash
# Send notification via HTTP
curl -X POST http://127.0.0.1:8765/notify \
  -H "Content-Type: application/json" \
  -d '{"event": "work_done", "message": "Build complete"}'

# Health check
curl http://127.0.0.1:8765/health
```

## Building

```bash
make build          # Current platform
make build-all      # Linux, macOS, Windows
make install-user   # Install to ~/.local/bin
```

## Supported Platforms

|          | Sound | TTS | Desktop |
|----------|-------|-----|---------|
| **Linux**    | aplay/paplay | espeak-ng | notify-send |
| **macOS**    | afplay | say | osascript |
| **Windows**  | PowerShell | PowerShell | BurntToast |

## License

MIT
