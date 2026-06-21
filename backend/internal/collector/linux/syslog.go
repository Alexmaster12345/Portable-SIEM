package linux

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/portable-siem/siem/internal/models"
)

// FileCollector tails a log file and emits events.
type FileCollector struct {
	name     string
	path     string
	hostname string
	source   string
	parser   func(line string) *models.Event
}

func NewAuthLogCollector(hostname string) *FileCollector {
	return &FileCollector{
		name:     "auth.log",
		path:     "/var/log/auth.log",
		hostname: hostname,
		source:   "auth",
		parser:   parseAuthLine(hostname),
	}
}

func NewSyslogCollector(hostname string) *FileCollector {
	return &FileCollector{
		name:     "syslog",
		path:     "/var/log/syslog",
		hostname: hostname,
		source:   "syslog",
		parser:   parseSyslogLine(hostname),
	}
}

func NewSecureCollector(hostname string) *FileCollector {
	return &FileCollector{
		name:     "secure",
		path:     "/var/log/secure",
		hostname: hostname,
		source:   "auth",
		parser:   parseAuthLine(hostname),
	}
}

func (c *FileCollector) Name() string { return c.name }

func (c *FileCollector) Start(ctx context.Context, out chan<- *models.Event) error {
	f, err := os.Open(c.path)
	if err != nil {
		return fmt.Errorf("open %s: %w", c.path, err)
	}
	defer f.Close()

	// Seek to end so we only read new lines
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}

		if event := c.parser(line); event != nil {
			select {
			case out <- event:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func parseAuthLine(hostname string) func(string) *models.Event {
	return func(line string) *models.Event {
		event := &models.Event{
			ID:         uuid.New(),
			Timestamp:  time.Now(),
			ReceivedAt: time.Now(),
			Host:       hostname,
			Source:     "auth",
			Severity:   models.SeverityInfo,
			Message:    line,
			Fields:     map[string]string{},
		}

		// Detect SSH events
		if contains(line, "Failed password") {
			event.EventType = "login_failed"
			event.Severity = models.SeverityMedium
			if ip := extractIP(line); ip != "" {
				event.Fields["src_ip"] = ip
			}
			if user := extractUser(line); user != "" {
				event.Fields["username"] = user
			}
		} else if contains(line, "Accepted password") || contains(line, "Accepted publickey") {
			event.EventType = "login_success"
			event.Severity = models.SeverityInfo
			if ip := extractIP(line); ip != "" {
				event.Fields["src_ip"] = ip
			}
			if user := extractUser(line); user != "" {
				event.Fields["username"] = user
			}
		} else if contains(line, "sudo:") {
			event.EventType = "sudo"
			event.Severity = models.SeverityMedium
		} else if contains(line, "useradd") || contains(line, "userdel") {
			event.EventType = "user_change"
			event.Severity = models.SeverityHigh
		} else {
			event.EventType = "auth"
		}

		return event
	}
}

func parseSyslogLine(hostname string) func(string) *models.Event {
	return func(line string) *models.Event {
		return &models.Event{
			ID:         uuid.New(),
			Timestamp:  time.Now(),
			ReceivedAt: time.Now(),
			Host:       hostname,
			Source:     "syslog",
			EventType:  "syslog",
			Severity:   models.SeverityInfo,
			Message:    line,
			Fields:     map[string]string{},
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func extractIP(line string) string {
	// Simple extraction of "from X.X.X.X" or "from X.X.X.X port"
	const fromStr = "from "
	idx := 0
	for i := 0; i <= len(line)-len(fromStr); i++ {
		if line[i:i+len(fromStr)] == fromStr {
			idx = i + len(fromStr)
			end := idx
			for end < len(line) && (line[end] == '.' || (line[end] >= '0' && line[end] <= '9')) {
				end++
			}
			return line[idx:end]
		}
	}
	return ""
}

func extractUser(line string) string {
	const forStr = "for "
	for i := 0; i <= len(line)-len(forStr); i++ {
		if line[i:i+len(forStr)] == forStr {
			start := i + len(forStr)
			end := start
			for end < len(line) && line[end] != ' ' {
				end++
			}
			return line[start:end]
		}
	}
	return ""
}
