package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	serverURL = flag.String("server", "http://localhost:8080", "SIEM server URL")
	interval  = flag.Duration("interval", 5*time.Second, "Collection interval")
	hostname  = flag.String("hostname", "", "Override hostname")
)

func main() {
	flag.Parse()

	host := *hostname
	if host == "" {
		host, _ = os.Hostname()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collector := NewCollector(host, *serverURL)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	fmt.Printf("Linux agent started: host=%s server=%s\n", host, *serverURL)

	for {
		select {
		case <-ticker.C:
			events := collector.Collect(ctx)
			if len(events) > 0 {
				collector.Send(ctx, events)
			}
		case <-quit:
			fmt.Println("agent stopping")
			return
		}
	}
}

type Event struct {
	Timestamp  time.Time         `json:"timestamp"`
	Host       string            `json:"host"`
	Source     string            `json:"source"`
	EventType  string            `json:"event_type"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
	Fields     map[string]string `json:"fields"`
}

type Collector struct {
	host      string
	serverURL string
	client    *http.Client
}

func NewCollector(host, serverURL string) *Collector {
	return &Collector{
		host:      host,
		serverURL: serverURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Collector) Collect(ctx context.Context) []*Event {
	var events []*Event
	events = append(events, c.collectProcesses(ctx)...)
	events = append(events, c.collectNetworkConnections(ctx)...)
	return events
}

func (c *Collector) collectProcesses(ctx context.Context) []*Event {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var events []*Event
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Check if directory name is all digits (PID)
		isPID := true
		for _, ch := range entry.Name() {
			if ch < '0' || ch > '9' {
				isPID = false
				break
			}
		}
		if !isPID {
			continue
		}

		cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%s/cmdline", entry.Name()))
		if err != nil {
			continue
		}

		events = append(events, &Event{
			Timestamp: time.Now(),
			Host:      c.host,
			Source:    "agent",
			EventType: "process",
			Severity:  "info",
			Message:   fmt.Sprintf("process pid=%s", entry.Name()),
			Fields: map[string]string{
				"pid":     entry.Name(),
				"cmdline": string(cmdline),
			},
		})
	}
	return events
}

func (c *Collector) collectNetworkConnections(ctx context.Context) []*Event {
	// Read /proc/net/tcp for active connections
	data, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		return nil
	}

	return []*Event{{
		Timestamp: time.Now(),
		Host:      c.host,
		Source:    "agent",
		EventType: "network_state",
		Severity:  "info",
		Message:   "network connections snapshot",
		Fields: map[string]string{
			"tcp_entries": fmt.Sprintf("%d", len(bytes.Split(data, []byte("\n")))-2),
		},
	}}
}

func (c *Collector) Send(ctx context.Context, events []*Event) {
	for _, event := range events {
		body, _ := json.Marshal(event)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+"/api/v1/events", bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "send error: %v\n", err)
			continue
		}
		resp.Body.Close()
	}
}
