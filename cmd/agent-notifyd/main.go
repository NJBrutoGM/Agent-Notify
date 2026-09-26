package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/agent-notify/agent-notify/internal/config"
	"github.com/agent-notify/agent-notify/internal/daemon"
	"github.com/agent-notify/agent-notify/internal/notify"
)

var (
	version = "dev"
)

func main() {
	fs := flag.NewFlagSet("agent-notifyd", flag.ExitOnError)
	assetsDir := fs.String("assets", "", "Assets directory")
	_ = fs.Bool("foreground", false, "Run in foreground")
	fs.Parse(os.Args[1:])

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	// Override daemon enabled from CLI
	cfg.Daemon.Enabled = true

	// Determine assets directory
	assets := *assetsDir
	if assets == "" {
		exe, _ := os.Executable()
		assets = filepath.Join(filepath.Dir(exe), "assets")
		if _, err := os.Stat(assets); os.IsNotExist(err) {
			assets = filepath.Join(filepath.Dir(exe), "..", "assets")
		}
	}

	// Create dispatcher
	dispatcher := notify.NewDispatcher(cfg, assets)

	// Create and start server
	server := daemon.NewServer(&cfg.Daemon, dispatcher)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("agent-notifyd started (HTTP: %s:%d, Socket: %s)\n",
		cfg.Daemon.HTTPHost, cfg.Daemon.HTTPPort, cfg.Daemon.SocketPath)

	// Wait for signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
	server.Stop()
}