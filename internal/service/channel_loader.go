// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tickraft/tickraft/pkg/prism/alert"
	"github.com/tickraft/tickraft/pkg/prism/channel"
	"github.com/tickraft/tickraft/pkg/prism/channel/email"
	"github.com/tickraft/tickraft/pkg/prism/channel/webhook"
	"gopkg.in/yaml.v3"
)

// channelConfigFile is the JSON/YAML schema for a channel config file.
type channelConfigFile struct {
	Channels []channel.Config `yaml:"channels" json:"channels"`
}

// LoadChannelsError wraps an error encountered while loading channel
// configuration, retaining the offending config string for diagnostics.
type LoadChannelsError struct {
	// Config is the configuration string that produced the error.
	Config string
	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e *LoadChannelsError) Error() string {
	return fmt.Sprintf("channel: load channels from %q: %v", e.Config, e.Err)
}

// Unwrap returns the underlying error, allowing errors.Is and errors.As
// to traverse into the wrapped error.
func (e *LoadChannelsError) Unwrap() error { return e.Err }

// LoadChannels parses a channel configuration string into alert.Channel
// instances. The config string may be:
//
//   - An HTTP or HTTPS URL: a single webhook.Channel pointing at the URL
//     is returned.
//   - A path to a JSON or YAML file: the file is parsed according to the
//     channelConfigFile schema and one alert.Channel per entry is
//     returned. Supported entry types are the built-in "webhook" and
//     "email", plus any additional types registered via channel.Register.
//
// An empty config string returns nil, nil. Unknown channel types in the
// config file produce an error.
func LoadChannels(config string) ([]alert.Channel, error) {
	config = strings.TrimSpace(config)
	if config == "" {
		return nil, nil
	}
	if isHTTPURL(config) {
		return loadFromHTTPURL(config)
	}
	channels, err := loadChannelsFromFile(config)
	if err != nil {
		return nil, &LoadChannelsError{Config: config, Err: err}
	}
	return channels, nil
}

// loadFromHTTPURL constructs a single webhook channel from a URL string.
func loadFromHTTPURL(url string) ([]alert.Channel, error) {
	ch, err := webhook.New(webhook.Config{URL: url})
	if err != nil {
		return nil, &LoadChannelsError{Config: url, Err: err}
	}
	return []alert.Channel{ch}, nil
}

// loadChannelsFromFile reads and parses a channel config file. The file
// extension determines the format: ".yaml"/".yml" uses YAML, everything
// else uses JSON.
func loadChannelsFromFile(path string) ([]alert.Channel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var cfg channelConfigFile
	if isYAMLFile(path) {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal yaml: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal json: %w", err)
		}
	}
	channels := make([]alert.Channel, 0, len(cfg.Channels))
	for i, entry := range cfg.Channels {
		ch, err := buildChannelFromEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("channel #%d: %w", i, err)
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

// buildChannelFromEntry constructs an alert.Channel from a parsed config
// entry. It first consults the extension SPI registry (populated via
// channel.Register); if no factory is registered for the type, it falls
// back to the built-in webhook and email implementations.
func buildChannelFromEntry(entry channel.Config) (alert.Channel, error) {
	typeKey := strings.ToLower(entry.Type)
	if factory := channel.LookupFactory(typeKey); factory != nil {
		return factory(entry)
	}
	switch typeKey {
	case "webhook":
		return buildWebhook(entry)
	case "email":
		return buildEmail(entry)
	default:
		return nil, fmt.Errorf("unsupported channel type %q (supported: webhook, email)", entry.Type)
	}
}

// buildWebhook constructs a webhook.Channel from a config entry.
func buildWebhook(entry channel.Config) (alert.Channel, error) {
	cfg := webhook.Config{
		URL:     entry.URL,
		Headers: entry.Headers,
	}
	if entry.Timeout != "" {
		d, err := time.ParseDuration(entry.Timeout)
		if err != nil {
			return nil, fmt.Errorf("parse timeout %q: %w", entry.Timeout, err)
		}
		cfg.Timeout = d
	}
	ch, err := webhook.New(cfg)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

// buildEmail constructs an email.Channel from a config entry.
func buildEmail(entry channel.Config) (alert.Channel, error) {
	tlsMode, err := parseTLSMode(entry.TLSMode)
	if err != nil {
		return nil, err
	}
	authType, err := parseAuthType(entry.AuthType)
	if err != nil {
		return nil, err
	}
	cfg := email.Config{
		Host:     entry.Host,
		Port:     entry.Port,
		Username: entry.Username,
		Password: entry.Password,
		From:     entry.From,
		To:       entry.To,
		TLSMode:  tlsMode,
		AuthType: authType,
		HTMLMode: entry.HTMLMode,
	}
	ch, err := email.New(cfg)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

// parseTLSMode maps a string to an email.TLSMode. An empty string defaults
// to TLSModeNone (no TLS).
func parseTLSMode(s string) (email.TLSMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "none":
		return email.TLSModeNone, nil
	case "implicit":
		return email.TLSModeImplicit, nil
	case "starttls":
		return email.TLSModeStartTLS, nil
	default:
		return 0, fmt.Errorf("unknown tls_mode %q (want none, implicit, or starttls)", s)
	}
}

// parseAuthType maps a string to an email.AuthType. An empty string
// defaults to AuthTypePlain.
func parseAuthType(s string) (email.AuthType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "plain":
		return email.AuthTypePlain, nil
	case "login":
		return email.AuthTypeLogin, nil
	case "cram-md5":
		return email.AuthTypeCramMD5, nil
	default:
		return 0, fmt.Errorf("unknown auth_type %q (want plain, login, or cram-md5)", s)
	}
}

// isHTTPURL returns true for strings starting with http:// or https://.
func isHTTPURL(s string) bool {
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// isYAMLFile returns true for paths ending in .yaml or .yml.
func isYAMLFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}
