package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/portable-siem/siem/internal/models"
)

// SyslogReceiver listens on UDP 514 for syslog messages from routers, firewalls, switches.
type SyslogReceiver struct {
	address string
	port    int
}

func NewSyslogReceiver(address string, port int) *SyslogReceiver {
	return &SyslogReceiver{address: address, port: port}
}

func (s *SyslogReceiver) Name() string { return "syslog-udp" }

func (s *SyslogReceiver) Start(ctx context.Context, out chan<- *models.Event) error {
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("listen udp %s: %w", addr, err)
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buf := make([]byte, 65536)
	for {
		n, src, err := conn.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		raw := string(buf[:n])
		event := parseSyslogMessage(raw, src.(*net.UDPAddr).IP.String())

		select {
		case out <- event:
		case <-ctx.Done():
			return nil
		}
	}
}

// parseSyslogMessage parses RFC 3164 / RFC 5424 syslog messages.
func parseSyslogMessage(raw, srcIP string) *models.Event {
	raw = strings.TrimSpace(raw)
	fields := map[string]string{"src_ip": srcIP}
	severity := models.SeverityInfo
	host := srcIP
	message := raw

	// Parse priority <PRI>
	if len(raw) > 3 && raw[0] == '<' {
		end := strings.Index(raw, ">")
		if end > 0 {
			priStr := raw[1:end]
			var pri int
			fmt.Sscan(priStr, &pri)
			facility := pri >> 3
			level := pri & 0x07
			fields["facility"] = fmt.Sprintf("%d", facility)
			fields["priority"] = fmt.Sprintf("%d", level)
			severity = syslogLevelToSeverity(level)
			raw = raw[end+1:]
		}
	}

	// Try to extract hostname and message from the remainder
	parts := strings.SplitN(raw, " ", 4)
	if len(parts) >= 3 {
		host = parts[1]
		if len(parts) >= 4 {
			message = parts[3]
		}
	}

	// Classify network device events
	eventType := "network"
	if containsAny(message, "DENY", "DROP", "BLOCK", "reject") {
		eventType = "firewall_deny"
		severity = models.SeverityMedium
	} else if containsAny(message, "PERMIT", "ALLOW", "ACCEPT") {
		eventType = "firewall_permit"
	} else if containsAny(message, "VPN", "tunnel", "ipsec") {
		eventType = "vpn"
	} else if containsAny(message, "AUTH", "login", "password") {
		eventType = "auth"
	}

	return &models.Event{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		ReceivedAt: time.Now(),
		Host:       host,
		Source:     "network_syslog",
		EventType:  eventType,
		Severity:   severity,
		Message:    message,
		Raw:        raw,
		Fields:     fields,
	}
}

func syslogLevelToSeverity(level int) models.Severity {
	switch level {
	case 0, 1, 2:
		return models.SeverityCritical
	case 3:
		return models.SeverityHigh
	case 4:
		return models.SeverityMedium
	default:
		return models.SeverityInfo
	}
}

func containsAny(s string, subs ...string) bool {
	upper := strings.ToUpper(s)
	for _, sub := range subs {
		if strings.Contains(upper, strings.ToUpper(sub)) {
			return true
		}
	}
	return false
}
