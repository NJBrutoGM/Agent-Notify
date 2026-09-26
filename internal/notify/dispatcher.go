package notify

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/agent-notify/agent-notify/internal/config"
)

type Dispatcher struct {
	sound    *SoundNotifier
	tts      *TTSNotifier
	desktop  *DesktopNotifier
	assetsDir string
}

func NewDispatcher(cfg *config.Config, assetsDir string) *Dispatcher {
	return &Dispatcher{
		sound:    NewSoundNotifier(&SoundConfig{Enabled: cfg.Methods.Sound, Volume: cfg.Sound.Volume, Events: cfg.Sound.Events}, assetsDir),
		tts:      NewTTSNotifier(&TTSConfig{Enabled: cfg.Methods.TTS, Voice: cfg.TTS.Voice, Rate: cfg.TTS.Rate, Events: cfg.TTS.Events}),
		desktop:  NewDesktopNotifier(&DesktopConfig{Enabled: cfg.Methods.Desktop, Title: cfg.Desktop.Title, Events: cfg.Desktop.Events}),
		assetsDir: assetsDir,
	}
}

type NotifyOptions struct {
	Event       string
	Message     string
	Title       string
	SoundFile   string
	Methods     []string
	Volume      float64
	Async       bool
}

func (d *Dispatcher) Notify(opts NotifyOptions) error {
	if opts.Async {
		go d.doNotify(opts)
		return nil
	}
	return d.doNotify(opts)
}

func (d *Dispatcher) doNotify(opts NotifyOptions) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	methods := opts.Methods
	if len(methods) == 0 {
		methods = []string{"sound", "tts", "desktop"}
	}

	for _, method := range methods {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			var err error
			switch m {
			case "sound":
				err = d.sound.Notify(opts.Event, opts.SoundFile)
			case "tts":
				err = d.tts.Notify(opts.Event, opts.Message)
			case "desktop":
				err = d.desktop.Notify(opts.Event, opts.Title, opts.Message)
			}
			if err != nil {
				errChan <- fmt.Errorf("%s: %w", m, err)
			}
		}(method)
	}

	wg.Wait()
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (d *Dispatcher) NotifyContext(ctx context.Context, opts NotifyOptions) error {
	if opts.Async {
		go func() {
			_ = d.doNotify(opts)
		}()
		return nil
	}
	return d.doNotify(opts)
}