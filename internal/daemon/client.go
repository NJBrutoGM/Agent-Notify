package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/agent-notify/agent-notify/internal/config"
)

type Client struct {
	config *config.DaemonConfig
	httpClient *http.Client
}

func NewClient(cfg *config.DaemonConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Notify(event, message, title, soundFile string, methods []string, volume float64) error {
	// Try HTTP first
	if err := c.notifyHTTP(event, message, title, soundFile, methods, volume); err == nil {
		return nil
	}

	// Fallback to socket
	return c.notifySocket(event, message, title, soundFile, methods, volume)
}

func (c *Client) notifyHTTP(event, message, title, soundFile string, methods []string, volume float64) error {
	url := fmt.Sprintf("http://%s:%d/notify", c.config.HTTPHost, c.config.HTTPPort)
	
	reqBody := map[string]interface{}{
		"event":       event,
		"message":     message,
		"title":       title,
		"sound_file":  soundFile,
		"methods":     methods,
		"volume":      volume,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) notifySocket(event, message, title, soundFile string, methods []string, volume float64) error {
	socketPath := expandPath(c.config.SocketPath)
	
	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		// Try named pipe via TCP fallback
		conn, err = net.DialTimeout("tcp", "127.0.0.1:8765", 2*time.Second)
	} else {
		conn, err = net.DialTimeout("unix", socketPath, 2*time.Second)
	}

	if err != nil {
		return err
	}
	defer conn.Close()

	req := map[string]interface{}{
		"event":       event,
		"message":     message,
		"title":       title,
		"sound_file":  soundFile,
		"methods":     methods,
		"volume":      volume,
	}

	encoder := json.NewEncoder(conn)
	return encoder.Encode(req)
}

func (c *Client) IsRunning() bool {
	// Check HTTP
	url := fmt.Sprintf("http://%s:%d/health", c.config.HTTPHost, c.config.HTTPPort)
	resp, err := c.httpClient.Get(url)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	// Check socket
	socketPath := expandPath(c.config.SocketPath)
	if runtime.GOOS != "windows" {
		conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}

	return false
}