package parser

import (
	"regexp"
	"strings"

	"github.com/portable-siem/siem/internal/models"
)

// Enricher adds derived fields to an event after collection.
type Enricher interface {
	Enrich(e *models.Event)
}

// Pipeline chains multiple enrichers.
type Pipeline struct {
	enrichers []Enricher
}

func NewPipeline(enrichers ...Enricher) *Pipeline {
	return &Pipeline{enrichers: enrichers}
}

func (p *Pipeline) Process(e *models.Event) *models.Event {
	for _, en := range p.enrichers {
		en.Enrich(e)
	}
	return e
}

// ---- Enrichers ----

// GeoIPEnricher adds geo location to src_ip fields (stub — wire real MaxMind DB).
type GeoIPEnricher struct{}

func (g *GeoIPEnricher) Enrich(e *models.Event) {
	// TODO: integrate MaxMind GeoLite2
}

// IOCEnricher checks src_ip/dst_ip against known IOC list.
type IOCEnricher struct {
	iocSet map[string]bool
}

func NewIOCEnricher(iocs []string) *IOCEnricher {
	set := make(map[string]bool, len(iocs))
	for _, ioc := range iocs {
		set[ioc] = true
	}
	return &IOCEnricher{iocSet: set}
}

func (ie *IOCEnricher) Enrich(e *models.Event) {
	for _, field := range []string{"src_ip", "dst_ip", "domain", "url"} {
		if val, ok := e.Fields[field]; ok && ie.iocSet[val] {
			e.Tags = append(e.Tags, "ioc_match")
			e.Fields["ioc_match"] = field
			if e.Severity < models.SeverityHigh {
				e.Severity = models.SeverityHigh
			}
			return
		}
	}
}

// MITREEnricher maps known event types to ATT&CK technique IDs.
var mitreMap = map[string][]string{
	"login_failed":        {"T1110"},    // Brute Force
	"login_success":       {"T1078"},    // Valid Accounts
	"sudo":                {"T1548.003"}, // Abuse Elevation Control Mechanism
	"firewall_deny":       {"T1046"},    // Network Service Scanning
	"privilege_escalation": {"T1068"},   // Exploitation for Privilege Escalation
	"user_change":         {"T1136"},    // Create Account
}

type MITREEnricher struct{}

func (m *MITREEnricher) Enrich(e *models.Event) {
	if ids, ok := mitreMap[e.EventType]; ok {
		e.MitreIDs = append(e.MitreIDs, ids...)
	}
}

// NormalizeEnricher lowercases source/event_type and trims whitespace.
type NormalizeEnricher struct{}

func (n *NormalizeEnricher) Enrich(e *models.Event) {
	e.Source = strings.ToLower(strings.TrimSpace(e.Source))
	e.EventType = strings.ToLower(strings.TrimSpace(e.EventType))
	e.Host = strings.ToLower(strings.TrimSpace(e.Host))
}

// FieldExtractEnricher applies regex patterns to extract fields from messages.
type FieldExtractEnricher struct {
	patterns []*FieldPattern
}

type FieldPattern struct {
	Source    string
	EventType string
	Regex     *regexp.Regexp
	Fields    []string // capture group names
}

func NewFieldExtractEnricher() *FieldExtractEnricher {
	return &FieldExtractEnricher{
		patterns: []*FieldPattern{
			{
				Source: "auth",
				Regex:  regexp.MustCompile(`Failed password for (?:invalid user )?(\S+) from ([\d.]+) port (\d+)`),
				Fields: []string{"username", "src_ip", "src_port"},
			},
			{
				Source: "auth",
				Regex:  regexp.MustCompile(`Accepted \S+ for (\S+) from ([\d.]+) port (\d+)`),
				Fields: []string{"username", "src_ip", "src_port"},
			},
			{
				Source: "network_syslog",
				Regex:  regexp.MustCompile(`SRC=([\d.]+).*DST=([\d.]+).*DPT=(\d+)`),
				Fields: []string{"src_ip", "dst_ip", "dst_port"},
			},
		},
	}
}

func (f *FieldExtractEnricher) Enrich(e *models.Event) {
	for _, p := range f.patterns {
		if p.Source != "" && p.Source != e.Source {
			continue
		}
		matches := p.Regex.FindStringSubmatch(e.Message)
		if matches == nil {
			continue
		}
		for i, name := range p.Fields {
			if i+1 < len(matches) {
				e.Fields[name] = matches[i+1]
			}
		}
	}
}
