package linux

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/portable-siem/siem/internal/models"
)

// JournaldCollector reads structured logs from systemd journal.
type JournaldCollector struct {
	hostname string
	units    []string // optional: filter by unit names
}

func NewJournaldCollector(hostname string, units []string) *JournaldCollector {
	return &JournaldCollector{hostname: hostname, units: units}
}

func (c *JournaldCollector) Name() string { return "journald" }

func (c *JournaldCollector) Start(ctx context.Context, out chan<- *models.Event) error {
	args := []string{"-f", "-o", "json", "--no-pager", "-n", "0"}
	for _, u := range c.units {
		args = append(args, "-u", u)
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		_ = cmd.Process.Kill()
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var entry map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		event := parseJournalEntry(entry, c.hostname)
		select {
		case out <- event:
		case <-ctx.Done():
			return nil
		}
	}
	return cmd.Wait()
}

func parseJournalEntry(entry map[string]any, hostname string) *models.Event {
	getString := func(key string) string {
		if v, ok := entry[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	ts := time.Now()
	if usec := getString("__REALTIME_TIMESTAMP"); usec != "" {
		// journald gives microseconds since epoch
		var usecInt int64
		if _, err := fmt.Sscan(usec, &usecInt); err == nil {
			ts = time.Unix(usecInt/1e6, (usecInt%1e6)*1000)
		}
	}

	host := getString("_HOSTNAME")
	if host == "" {
		host = hostname
	}

	fields := map[string]string{
		"unit":     getString("_SYSTEMD_UNIT"),
		"pid":      getString("_PID"),
		"uid":      getString("_UID"),
		"priority": getString("PRIORITY"),
		"comm":     getString("_COMM"),
	}

	severity := priorityToSeverity(getString("PRIORITY"))

	return &models.Event{
		ID:        uuid.New(),
		Timestamp: ts,
		ReceivedAt: time.Now(),
		Host:      host,
		Source:    "journald",
		EventType: getString("_SYSTEMD_UNIT"),
		Severity:  severity,
		Message:   getString("MESSAGE"),
		Fields:    fields,
	}
}

func priorityToSeverity(priority string) models.Severity {
	switch priority {
	case "0", "1", "2": // emerg, alert, crit
		return models.SeverityCritical
	case "3": // error
		return models.SeverityHigh
	case "4": // warning
		return models.SeverityMedium
	case "5", "6": // notice, info
		return models.SeverityInfo
	default:
		return models.SeverityInfo
	}
}
