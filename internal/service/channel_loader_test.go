// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tickraft/tickraft/pkg/prism/alert"
	"github.com/tickraft/tickraft/pkg/prism/channel"
)

// ---------------------------------------------------------------------------
// Test fakes (for registry integration tests)
// ---------------------------------------------------------------------------

// fakeChannel is a minimal alert.Channel implementation used to verify
// that the registry dispatches to registered factories.
type fakeChannel struct {
	name string
}

func (f *fakeChannel) Send(_ context.Context, _ alert.Event) error { return nil }
func (f *fakeChannel) Name() string                                { return f.name }

// capturingFactory returns a Factory that records each invocation
// and returns a fakeChannel whose name is derived from the config Type.
func capturingFactory(counter *int32) channel.Factory {
	return func(cfg channel.Config) (alert.Channel, error) {
		atomic.AddInt32(counter, 1)
		return &fakeChannel{name: cfg.Type}, nil
	}
}

// writeConfigFile writes content to a temporary file with the given name
// and returns the full path.
func writeConfigFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
	return path
}

// writeFile is an alias for writeConfigFile for compatibility with
// load tests that use the original helper name.
func writeFile(t *testing.T, name, content string) string {
	return writeConfigFile(t, name, content)
}

// ---------------------------------------------------------------------------
// LoadChannels: empty and URL inputs
// ---------------------------------------------------------------------------

func TestLoadChannelsEmpty(t *testing.T) {
	channels, err := LoadChannels("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channels != nil {
		t.Errorf("expected nil channels, got %v", channels)
	}
}

func TestLoadChannelsWhitespaceOnly(t *testing.T) {
	channels, err := LoadChannels("   \t\n  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channels != nil {
		t.Errorf("expected nil channels, got %v", channels)
	}
}

func TestLoadChannelsHTTPURL(t *testing.T) {
	channels, err := LoadChannels("http://example.com/hook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel name: got %q, want %q", name, "webhook")
	}
}

func TestLoadChannelsHTTPSURL(t *testing.T) {
	channels, err := LoadChannels("https://example.com/hook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel name: got %q, want %q", name, "webhook")
	}
}

// ---------------------------------------------------------------------------
// LoadChannels: JSON file loading
// ---------------------------------------------------------------------------

func TestLoadChannelsJSONWebhook(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "webhook",
      "url": "https://example.com/hook"
    }
  ]
}`
	path := writeFile(t, "channels_webhook.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel name: got %q, want %q", name, "webhook")
	}
}

func TestLoadChannelsJSONEmail(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "from": "alert@example.com",
      "to": ["ops@example.com"]
    }
  ]
}`
	path := writeFile(t, "channels_email.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "email" {
		t.Errorf("channel name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsJSONMixed(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "webhook",
      "url": "https://example.com/hook",
      "timeout": "10s",
      "headers": {"Authorization": "Bearer xxx"}
    },
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "username": "user@example.com",
      "password": "secret",
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "tls_mode": "starttls",
      "auth_type": "plain",
      "html_mode": true
    }
  ]
}`
	path := writeFile(t, "channels_mixed.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel[0] name: got %q, want %q", name, "webhook")
	}
	if name := channels[1].Name(); name != "email" {
		t.Errorf("channel[1] name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsJSONEmptyChannels(t *testing.T) {
	const jsonContent = `{"channels": []}`
	path := writeFile(t, "channels_empty.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("expected 0 channels, got %d", len(channels))
	}
}

// ---------------------------------------------------------------------------
// LoadChannels: YAML file loading
// ---------------------------------------------------------------------------

func TestLoadChannelsYAMLMixed(t *testing.T) {
	const yamlContent = `channels:
  - type: webhook
    url: https://example.com/hook
    timeout: 10s
    headers:
      Authorization: Bearer xxx
  - type: email
    host: smtp.example.com
    port: 587
    username: user@example.com
    password: secret
    from: alert@example.com
    to:
      - ops@example.com
    tls_mode: starttls
    auth_type: plain
    html_mode: true
`
	path := writeFile(t, "channels_mixed.yaml", yamlContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel[0] name: got %q, want %q", name, "webhook")
	}
	if name := channels[1].Name(); name != "email" {
		t.Errorf("channel[1] name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsYAMLExtension(t *testing.T) {
	const yamlContent = `channels:
  - type: webhook
    url: https://example.com/hook
`
	path := writeFile(t, "channels.yml", yamlContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel name: got %q, want %q", name, "webhook")
	}
}

// ---------------------------------------------------------------------------
// LoadChannels: webhook features
// ---------------------------------------------------------------------------

func TestLoadChannelsWebhookCustomHeaders(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "webhook",
      "url": "https://example.com/hook",
      "headers": {
        "Authorization": "Bearer token123",
        "X-Custom": "value"
      }
    }
  ]
}`
	path := writeFile(t, "channels_headers.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel name: got %q, want %q", name, "webhook")
	}
}

func TestLoadChannelsWebhookCustomTimeout(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "webhook",
      "url": "https://example.com/hook",
      "timeout": "15s"
    }
  ]
}`
	path := writeFile(t, "channels_timeout.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "webhook" {
		t.Errorf("channel name: got %q, want %q", name, "webhook")
	}
}

func TestLoadChannelsWebhookInvalidTimeout(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "webhook",
      "url": "https://example.com/hook",
      "timeout": "not-a-duration"
    }
  ]
}`
	path := writeFile(t, "channels_bad_timeout.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for invalid timeout, got nil")
	}
}

// ---------------------------------------------------------------------------
// LoadChannels: email features
// ---------------------------------------------------------------------------

func TestLoadChannelsEmailHTMLMode(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "html_mode": true
    }
  ]
}`
	path := writeFile(t, "channels_email_html.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "email" {
		t.Errorf("channel name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsEmailSTARTTLS(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "tls_mode": "starttls"
    }
  ]
}`
	path := writeFile(t, "channels_email_starttls.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "email" {
		t.Errorf("channel name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsEmailPlainAuth(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "username": "user@example.com",
      "password": "secret",
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "auth_type": "plain"
    }
  ]
}`
	path := writeFile(t, "channels_email_plain.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "email" {
		t.Errorf("channel name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsEmailImplicitTLS(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 465,
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "tls_mode": "implicit"
    }
  ]
}`
	path := writeFile(t, "channels_email_implicit.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if name := channels[0].Name(); name != "email" {
		t.Errorf("channel name: got %q, want %q", name, "email")
	}
}

func TestLoadChannelsEmailLoginAuth(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "username": "user@example.com",
      "password": "secret",
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "auth_type": "login"
    }
  ]
}`
	path := writeFile(t, "channels_email_login.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
}

func TestLoadChannelsEmailCramMD5Auth(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "username": "user@example.com",
      "password": "secret",
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "auth_type": "cram-md5"
    }
  ]
}`
	path := writeFile(t, "channels_email_crammd5.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
}

// ---------------------------------------------------------------------------
// LoadChannels: error cases
// ---------------------------------------------------------------------------

func TestLoadChannelsUnknownType(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "feishu",
      "url": "https://example.com/hook"
    }
  ]
}`
	path := writeFile(t, "channels_unknown.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for unknown channel type, got nil")
	}
}

func TestLoadChannelsFileNotFound(t *testing.T) {
	_, err := LoadChannels("/nonexistent/path/to/channels.json")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestLoadChannelsInvalidJSON(t *testing.T) {
	const invalidJSON = `{invalid json content`
	path := writeFile(t, "channels_invalid.json", invalidJSON)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadChannelsInvalidYAML(t *testing.T) {
	const invalidYAML = `channels:
  - type: webhook
    url: [invalid
    : broken
`
	path := writeFile(t, "channels_invalid.yaml", invalidYAML)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadChannelsEmailMissingHost(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "from": "alert@example.com",
      "to": ["ops@example.com"]
    }
  ]
}`
	path := writeFile(t, "channels_email_no_host.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for email missing host, got nil")
	}
}

func TestLoadChannelsWebhookMissingURL(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "webhook"
    }
  ]
}`
	path := writeFile(t, "channels_webhook_no_url.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for webhook missing url, got nil")
	}
}

func TestLoadChannelsInvalidTLSMode(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "tls_mode": "ssl"
    }
  ]
}`
	path := writeFile(t, "channels_bad_tls.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for invalid tls_mode, got nil")
	}
}

func TestLoadChannelsInvalidAuthType(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "from": "alert@example.com",
      "to": ["ops@example.com"],
      "auth_type": "oauth"
    }
  ]
}`
	path := writeFile(t, "channels_bad_auth.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for invalid auth_type, got nil")
	}
}

func TestLoadChannelsEmailPasswordWithoutUsername(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "password": "secret",
      "from": "alert@example.com",
      "to": ["ops@example.com"]
    }
  ]
}`
	path := writeFile(t, "channels_email_pw_no_user.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for password without username, got nil")
	}
}

// ---------------------------------------------------------------------------
// LoadChannelsError
// ---------------------------------------------------------------------------

func TestLoadChannelsErrorWrapping(t *testing.T) {
	const configPath = "/nonexistent/path/channels.json"
	_, err := LoadChannels(configPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var lce *LoadChannelsError
	if !errors.As(err, &lce) {
		t.Fatalf("expected *LoadChannelsError, got %T: %v", err, err)
	}
	if lce.Config != configPath {
		t.Errorf("Config: got %q, want %q", lce.Config, configPath)
	}
	if lce.Err == nil {
		t.Error("expected non-nil inner error")
	}
}

func TestLoadChannelsErrorErrorString(t *testing.T) {
	lce := &LoadChannelsError{
		Config: "/path/to/file.json",
		Err:    errors.New("read file: no such file"),
	}
	got := lce.Error()
	want := `channel: load channels from "/path/to/file.json": read file: no such file`
	if got != want {
		t.Errorf("Error(): got %q, want %q", got, want)
	}
}

func TestLoadChannelsErrorUnwrap(t *testing.T) {
	inner := errors.New("inner failure")
	lce := &LoadChannelsError{Config: "cfg", Err: inner}
	if unwrapped := lce.Unwrap(); unwrapped != inner {
		t.Errorf("Unwrap(): got %v, want %v", unwrapped, inner)
	}
}

func TestLoadChannelsErrorErrorsIs(t *testing.T) {
	sentinel := errors.New("sentinel")
	lce := &LoadChannelsError{Config: "cfg", Err: sentinel}
	if !errors.Is(lce, sentinel) {
		t.Error("errors.Is(lce, sentinel): got false, want true")
	}
}

// ---------------------------------------------------------------------------
// parseTLSMode
// ---------------------------------------------------------------------------

func TestParseTLSMode(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"none", 0},
		{"NONE", 0},
		{" None ", 0},
		{"implicit", 1},
		{"IMPLICIT", 1},
		{"starttls", 2},
		{"StartTLS", 2},
	}
	for _, tt := range tests {
		got, err := parseTLSMode(tt.input)
		if err != nil {
			t.Errorf("parseTLSMode(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if int(got) != tt.want {
			t.Errorf("parseTLSMode(%q): got %d, want %d", tt.input, int(got), tt.want)
		}
	}
}

func TestParseTLSModeInvalid(t *testing.T) {
	_, err := parseTLSMode("ssl")
	if err == nil {
		t.Fatal("expected error for invalid tls_mode, got nil")
	}
}

// ---------------------------------------------------------------------------
// parseAuthType
// ---------------------------------------------------------------------------

func TestParseAuthType(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"plain", 0},
		{"PLAIN", 0},
		{" plain ", 0},
		{"login", 1},
		{"LOGIN", 1},
		{"cram-md5", 2},
		{"CRAM-MD5", 2},
	}
	for _, tt := range tests {
		got, err := parseAuthType(tt.input)
		if err != nil {
			t.Errorf("parseAuthType(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if int(got) != tt.want {
			t.Errorf("parseAuthType(%q): got %d, want %d", tt.input, int(got), tt.want)
		}
	}
}

func TestParseAuthTypeInvalid(t *testing.T) {
	_, err := parseAuthType("oauth")
	if err == nil {
		t.Fatal("expected error for invalid auth_type, got nil")
	}
}

// ---------------------------------------------------------------------------
// isHTTPURL / isYAMLFile
// ---------------------------------------------------------------------------

func TestIsHTTPURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"HTTP://example.com", true},
		{"Https://example.com", true},
		{"ftp://example.com", false},
		{"/path/to/file.json", false},
		{"example.com", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isHTTPURL(tt.input); got != tt.want {
			t.Errorf("isHTTPURL(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsYAMLFile(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"channels.yaml", true},
		{"channels.yml", true},
		{"channels.YAML", true},
		{"channels.YML", true},
		{"channels.json", false},
		{"channels.txt", false},
		{"channels", false},
		{"/path/to/channels.yaml", true},
	}
	for _, tt := range tests {
		if got := isYAMLFile(tt.input); got != tt.want {
			t.Errorf("isYAMLFile(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Multiple channels and ordering
// ---------------------------------------------------------------------------

func TestLoadChannelsMultipleWebhooks(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook1"},
    {"type": "webhook", "url": "https://example.com/hook2"},
    {"type": "webhook", "url": "https://example.com/hook3"}
  ]
}`
	path := writeFile(t, "channels_multi_webhook.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(channels))
	}
	for i, ch := range channels {
		if name := ch.Name(); name != "webhook" {
			t.Errorf("channel[%d] name: got %q, want %q", i, name, "webhook")
		}
	}
}

func TestLoadChannelsErrorOnSecondChannel(t *testing.T) {
	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook"},
    {"type": "feishu", "url": "https://example.com/hook"}
  ]
}`
	path := writeFile(t, "channels_second_bad.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Registry integration tests (moved from pkg/prism/channel/registry_test.go)
// ---------------------------------------------------------------------------

func TestRegisterCustomFactory(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var calls int32
	channel.Register("sms", capturingFactory(&calls))

	const jsonContent = `{
  "channels": [
    {"type": "sms", "url": "https://example.com/sms"}
  ]
}`
	path := writeConfigFile(t, "channels_sms.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if got := channels[0].Name(); got != "sms" {
		t.Errorf("channel name: got %q, want %q", got, "sms")
	}
	if calls != 1 {
		t.Errorf("factory call count: got %d, want 1", calls)
	}
}

func TestRegisterFactoryReceivesConfig(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var captured channel.Config
	channel.Register("dingtalk", func(cfg channel.Config) (alert.Channel, error) {
		captured = cfg
		return &fakeChannel{name: "dingtalk"}, nil
	})

	const jsonContent = `{
  "channels": [
    {
      "type": "dingtalk",
      "url": "https://oapi.dingtalk.com/robot/send",
      "headers": {"Authorization": "Bearer secret"}
    }
  ]
}`
	path := writeConfigFile(t, "channels_dingtalk.json", jsonContent)
	if _, err := LoadChannels(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Type != "dingtalk" {
		t.Errorf("captured Type: got %q, want %q", captured.Type, "dingtalk")
	}
	if captured.URL != "https://oapi.dingtalk.com/robot/send" {
		t.Errorf("captured URL: got %q", captured.URL)
	}
	if captured.Headers["Authorization"] != "Bearer secret" {
		t.Errorf("captured Headers[Authorization]: got %q", captured.Headers["Authorization"])
	}
}

func TestRegisterFallbackToBuiltinWebhook(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook"}
  ]
}`
	path := writeConfigFile(t, "channels_builtin_webhook.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if got := channels[0].Name(); got != "webhook" {
		t.Errorf("channel name: got %q, want %q", got, "webhook")
	}
}

func TestRegisterFallbackToBuiltinEmail(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	const jsonContent = `{
  "channels": [
    {
      "type": "email",
      "host": "smtp.example.com",
      "port": 587,
      "from": "alert@example.com",
      "to": ["ops@example.com"]
    }
  ]
}`
	path := writeConfigFile(t, "channels_builtin_email.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if got := channels[0].Name(); got != "email" {
		t.Errorf("channel name: got %q, want %q", got, "email")
	}
}

func TestRegisterUnknownTypeStillErrors(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	const jsonContent = `{
  "channels": [
    {"type": "feishu", "url": "https://example.com/hook"}
  ]
}`
	path := writeConfigFile(t, "channels_unknown.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error for unregistered type, got nil")
	}
}

func TestRegisterMultipleFactories(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var smsCalls, dingtalkCalls int32
	channel.Register("sms", capturingFactory(&smsCalls))
	channel.Register("dingtalk", capturingFactory(&dingtalkCalls))

	const jsonContent = `{
  "channels": [
    {"type": "sms", "url": "https://example.com/sms"},
    {"type": "dingtalk", "url": "https://example.com/dingtalk"},
    {"type": "sms", "url": "https://example.com/sms2"}
  ]
}`
	path := writeConfigFile(t, "channels_multi.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(channels))
	}
	if got := channels[0].Name(); got != "sms" {
		t.Errorf("channel[0] name: got %q, want %q", got, "sms")
	}
	if got := channels[1].Name(); got != "dingtalk" {
		t.Errorf("channel[1] name: got %q, want %q", got, "dingtalk")
	}
	if got := channels[2].Name(); got != "sms" {
		t.Errorf("channel[2] name: got %q, want %q", got, "sms")
	}
	if smsCalls != 2 {
		t.Errorf("sms factory call count: got %d, want 2", smsCalls)
	}
	if dingtalkCalls != 1 {
		t.Errorf("dingtalk factory call count: got %d, want 1", dingtalkCalls)
	}
}

func TestRegisterMixedBuiltinAndCustom(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var slackCalls int32
	channel.Register("slack", capturingFactory(&slackCalls))

	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook"},
    {"type": "slack", "url": "https://hooks.slack.com/services/x"},
    {"type": "email", "host": "smtp.example.com", "port": 587, "from": "a@b.com", "to": ["c@d.com"]}
  ]
}`
	path := writeConfigFile(t, "channels_mixed_registry.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(channels))
	}
	if got := channels[0].Name(); got != "webhook" {
		t.Errorf("channel[0] name: got %q, want %q", got, "webhook")
	}
	if got := channels[1].Name(); got != "slack" {
		t.Errorf("channel[1] name: got %q, want %q", got, "slack")
	}
	if got := channels[2].Name(); got != "email" {
		t.Errorf("channel[2] name: got %q, want %q", got, "email")
	}
	if slackCalls != 1 {
		t.Errorf("slack factory call count: got %d, want 1", slackCalls)
	}
}

func TestRegisterTypeCaseInsensitive(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var calls int32
	channel.Register("slack", capturingFactory(&calls))

	const jsonContent = `{
  "channels": [
    {"type": "SLACK", "url": "https://hooks.slack.com/services/x"},
    {"type": "Slack", "url": "https://hooks.slack.com/services/y"}
  ]
}`
	path := writeConfigFile(t, "channels_case.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if calls != 2 {
		t.Errorf("factory call count: got %d, want 2", calls)
	}
}

func TestRegisterOverridesBuiltinWebhook(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var calls int32
	channel.Register("webhook", capturingFactory(&calls))

	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook"}
  ]
}`
	path := writeConfigFile(t, "channels_override_webhook.json", jsonContent)
	channels, err := LoadChannels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	if calls != 1 {
		t.Errorf("override factory call count: got %d, want 1", calls)
	}
}

func TestRegisterOverrideThenResetRestoresBuiltin(t *testing.T) {
	var calls int32
	channel.Register("webhook", capturingFactory(&calls))

	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook"}
  ]
}`
	path := writeConfigFile(t, "channels_reset.json", jsonContent)

	if _, err := LoadChannels(path); err != nil {
		t.Fatalf("unexpected error during override: %v", err)
	}
	if calls != 1 {
		t.Fatalf("override factory call count: got %d, want 1", calls)
	}

	channel.ResetRegistryForTest()

	if _, err := LoadChannels(path); err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if calls != 1 {
		t.Errorf("override factory call count after reset: got %d, want 1", calls)
	}
}

func TestRegisterFactoryErrorPropagation(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	factoryErr := errors.New("factory boom")
	channel.Register("broken", func(_ channel.Config) (alert.Channel, error) {
		return nil, factoryErr
	})

	const jsonContent = `{
  "channels": [
    {"type": "broken", "url": "https://example.com/x"}
  ]
}`
	path := writeConfigFile(t, "channels_broken.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error from factory, got nil")
	}
	if !errors.Is(err, factoryErr) {
		t.Errorf("expected error to wrap factoryErr, got %v", err)
	}
}

func TestRegisterFactoryErrorOnSecondChannel(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	factoryErr := errors.New("second channel failed")
	callCount := 0
	channel.Register("intermittent", func(_ channel.Config) (alert.Channel, error) {
		callCount++
		if callCount == 2 {
			return nil, factoryErr
		}
		return &fakeChannel{name: "intermittent"}, nil
	})

	const jsonContent = `{
  "channels": [
    {"type": "intermittent", "url": "https://example.com/1"},
    {"type": "intermittent", "url": "https://example.com/2"}
  ]
}`
	path := writeConfigFile(t, "channels_intermittent.json", jsonContent)
	_, err := LoadChannels(path)
	if err == nil {
		t.Fatal("expected error on second channel, got nil")
	}
	if !errors.Is(err, factoryErr) {
		t.Errorf("expected error to wrap factoryErr, got %v", err)
	}
}

func TestRegisterConcurrentRegisterAndLoad(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	const jsonContent = `{
  "channels": [
    {"type": "webhook", "url": "https://example.com/hook"}
  ]
}`
	path := writeConfigFile(t, "channels_concurrent.json", jsonContent)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			channel.Register(fmt.Sprintf("race-type-%d", n), capturingFactory(new(int32)))
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = LoadChannels(path)
		}()
	}
	wg.Wait()
}

func TestRegisterOverwriteFactory(t *testing.T) {
	t.Cleanup(channel.ResetRegistryForTest)

	var firstCalls, secondCalls int32
	channel.Register("overwritable", capturingFactory(&firstCalls))
	channel.Register("overwritable", capturingFactory(&secondCalls))

	const jsonContent = `{
  "channels": [
    {"type": "overwritable", "url": "https://example.com/x"}
  ]
}`
	path := writeConfigFile(t, "channels_overwrite.json", jsonContent)
	if _, err := LoadChannels(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstCalls != 0 {
		t.Errorf("first factory call count: got %d, want 0", firstCalls)
	}
	if secondCalls != 1 {
		t.Errorf("second factory call count: got %d, want 1", secondCalls)
	}
}
