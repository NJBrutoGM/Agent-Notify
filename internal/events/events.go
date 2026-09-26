package events

const (
	EventWorkDone        = "work_done"
	EventAttentionNeeded = "attention_needed"
	EventError           = "error"
	EventWarning         = "warning"
	EventInfo            = "info"
	EventCustom          = "custom"
)

var PredefinedEvents = []string{
	EventWorkDone,
	EventAttentionNeeded,
	EventError,
	EventWarning,
	EventInfo,
	EventCustom,
}

func IsPredefined(event string) bool {
	for _, e := range PredefinedEvents {
		if e == event {
			return true
		}
	}
	return false
}

func GetDefaultSound(event string) string {
	switch event {
	case EventWorkDone:
		return "success.wav"
	case EventAttentionNeeded:
		return "alert.wav"
	case EventError:
		return "error.wav"
	case EventWarning:
		return "warning.wav"
	case EventInfo:
		return "info.wav"
	default:
		return "info.wav"
	}
}

func GetDefaultTTS(event string) string {
	switch event {
	case EventWorkDone:
		return "Work completed"
	case EventAttentionNeeded:
		return "Attention required"
	case EventError:
		return "Error occurred"
	case EventWarning:
		return "Warning"
	case EventInfo:
		return "Information"
	default:
		return "Notification"
	}
}

func GetDefaultDesktop(event string) string {
	switch event {
	case EventWorkDone:
		return "Work completed"
	case EventAttentionNeeded:
		return "Attention needed"
	case EventError:
		return "Error"
	case EventWarning:
		return "Warning"
	case EventInfo:
		return "Info"
	default:
		return "Notification"
	}
}