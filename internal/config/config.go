package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job defines a monitored cron job entry.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Timeout  time.Duration `yaml:"timeout"`
	AlertOn  []string      `yaml:"alert_on"` // "missed", "failed"
}

// AlertConfig holds alerting backend configuration.
type AlertConfig struct {
	Email   *EmailAlert   `yaml:"email,omitempty"`
	Slack   *SlackAlert   `yaml:"slack,omitempty"`
}

// EmailAlert holds SMTP alert settings.
type EmailAlert struct {
	SMTPHost string   `yaml:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
}

// SlackAlert holds Slack webhook settings.
type SlackAlert struct {
	WebhookURL string `yaml:"webhook_url"`
}

// Config is the top-level cronwatch configuration.
type Config struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Jobs          []Job         `yaml:"jobs"`
	Alerts        AlertConfig   `yaml:"alerts"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = time.Minute
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	seen := make(map[string]bool)
	for _, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job missing name")
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q missing schedule", j.Name)
		}
		if seen[j.Name] {
			return fmt.Errorf("duplicate job name %q", j.Name)
		}
		seen[j.Name] = true
	}
	return nil
}
