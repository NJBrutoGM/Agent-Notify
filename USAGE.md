# How to Use Agent Notify

## After Setup

Just type `notify` in your terminal:

```bash
notify -t "Hey there I am User"
```

## Quick Reference

### Text-to-Speech (TTS)
```bash
notify -t "Your message here"
notify --tts "Build completed successfully"
notify -t "Task done" --volume 0.5
```

### Sound
```bash
notify -s success
notify --sound alert
notify -s error --method sound,desktop
```

### Desktop Notification
```bash
notify -n "Title" -m "Message body"
notify --notify "Done" --message "All finished"
```

### Combined (TTS + Sound + Desktop)
```bash
notify -t "Complete" -s success -n "Done" -m "All tasks finished"
```

### Predefined Events
```bash
notify work_done          # Task completed
notify attention_needed   # Needs human input
notify error "Build failed"
notify warning "Deprecated API"
notify info "Starting..."
notify custom --t "Custom message" -s success
```

### Daemon
```bash
notify daemon status      # Check daemon status
notify --daemon-status    # Same thing
```

### Start Daemon (for faster notifications)
```bash
agent-notifyd start
notify work_done          # Goes through daemon automatically
```

## Available Sounds
- `success` - Task completed
- `alert` - Attention needed
- `error` - Error occurred
- `warning` - Warning
- `info` - Informational

## Adding Custom Sounds
1. Place `.wav` files in `~/.local/share/agent-notify/assets/sounds/`
2. Or edit config: `nano ~/.config/agent-notify/config.yaml`

## Config File
```bash
nano ~/.config/agent-notify/config.yaml
```

## Troubleshooting
```bash
notify --daemon-status     # Check if daemon is running
notify -t "Test" --no-daemon  # Force direct notification
```