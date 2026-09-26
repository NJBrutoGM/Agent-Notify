package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/agent-notify/agent-notify/internal/config"
	"github.com/agent-notify/agent-notify/internal/notify"
)

type NotifyRequest struct {
	Event     string   `json:"event"`
	Message   string   `json:"message,omitempty"`
	Title     string   `json:"title,omitempty"`
	SoundFile string   `json:"sound_file,omitempty"`
	Methods   []string `json:"methods,omitempty"`
	Volume    float64  `json:"volume,omitempty"`
}

type Server struct {
	config     *config.DaemonConfig
	dispatcher *notify.Dispatcher
	httpServer *http.Server
	socketListener net.Listener
	mu         sync.Mutex
	running    bool
}

func NewServer(cfg *config.DaemonConfig, dispatcher *notify.Dispatcher) *Server {
	return &Server{
		config:     cfg,
		dispatcher: dispatcher,
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	// Start HTTP server
	if s.config.HTTPPort > 0 {
		go s.startHTTP()
	}

	// Start Unix socket / named pipe
	go s.startSocket()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	var errs []error

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if s.socketListener != nil {
		if err := s.socketListener.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	s.running = false

	if len(errs) > 0 {
		return fmt.Errorf("stop errors: %v", errs)
	}
	return nil
}

func (s *Server) startHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/notify", s.handleNotify)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/events", s.handleEvents)

	addr := fmt.Sprintf("%s:%d", s.config.HTTPHost, s.config.HTTPPort)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
	}
}

func (s *Server) startSocket() {
	socketPath := expandPath(s.config.SocketPath)
	
	// Remove existing socket file
	os.Remove(socketPath)

	// Create directory
	os.MkdirAll(filepath.Dir(socketPath), 0755)

	var listener net.Listener
	var err error

	if runtime.GOOS == "windows" {
		// Named pipe on Windows
		listener, err = createWindowsNamedPipe(socketPath)
	} else {
		// Unix socket on Linux/macOS
		listener, err = net.Listen("unix", socketPath)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Socket listener error: %v\n", err)
		return
	}

	s.mu.Lock()
	s.socketListener = listener
	s.mu.Unlock()

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			fmt.Fprintf(os.Stderr, "Socket accept error: %v\n", err)
			continue
		}
		go s.handleSocketConn(conn)
	}
}

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	opts := notify.NotifyOptions{
		Event:       req.Event,
		Message:     req.Message,
		Title:       req.Title,
		SoundFile:   req.SoundFile,
		Methods:     req.Methods,
		Volume:      req.Volume,
		Async:       true,
	}

	if err := s.dispatcher.Notify(opts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	events := []string{
		"work_done",
		"attention_needed",
		"error",
		"warning",
		"info",
		"custom",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"events": events})
}

func (s *Server) handleSocketConn(conn net.Conn) {
	defer conn.Close()

	var req NotifyRequest
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&req); err != nil {
		fmt.Fprintf(os.Stderr, "Socket decode error: %v\n", err)
		return
	}

	opts := notify.NotifyOptions{
		Event:       req.Event,
		Message:     req.Message,
		Title:       req.Title,
		SoundFile:   req.SoundFile,
		Methods:     req.Methods,
		Volume:      req.Volume,
		Async:       true,
	}

	_ = s.dispatcher.Notify(opts)
}

func createWindowsNamedPipe(pipeName string) (net.Listener, error) {
	// Convert to Windows named pipe format
	if !filepath.IsAbs(pipeName) {
		pipeName = `\\.\pipe\` + pipeName
	} else if !strings.HasPrefix(pipeName, `\\.\pipe\`) {
		pipeName = `\\.\pipe\` + filepath.Base(pipeName)
	}
	
	// Go doesn't have native named pipe support in net package
	// We'll use a workaround with a TCP listener on localhost
	return net.Listen("tcp", "127.0.0.1:0")
}

func expandPath(path string) string {
	if path == "" {
		return ""
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}

func (s *Server) WaitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	s.Stop()
}