// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package channel

// Config is a single channel entry in the config file. The Type field
// selects which channel implementation to construct; the remaining fields
// are interpreted by the selected implementation.
//
// Config is exported so that extension SPI factories registered via Register
// can be constructed from a parsed entry by downstream repositories.
type Config struct {
	// Type selects the channel implementation. It is matched
	// case-insensitively against the built-in types ("webhook",
	// "email") and any types registered via Register.
	Type string `yaml:"type" json:"type"`
	// URL is the target endpoint for webhook channels.
	URL string `yaml:"url" json:"url,omitempty"`
	// Timeout is the request timeout for webhook channels, parsed
	// as a time.Duration string (e.g. "10s").
	Timeout string `yaml:"timeout" json:"timeout,omitempty"`
	// Headers carries custom HTTP headers for webhook channels.
	Headers map[string]string `yaml:"headers" json:"headers,omitempty"`
	// Host is the SMTP server hostname for email channels.
	Host string `yaml:"host" json:"host,omitempty"`
	// Port is the SMTP server port for email channels.
	Port int `yaml:"port" json:"port,omitempty"`
	// Username is the SMTP authentication username for email channels.
	Username string `yaml:"username" json:"username,omitempty"`
	// Password is the SMTP authentication password for email channels.
	Password string `yaml:"password" json:"password,omitempty"`
	// From is the sender address for email channels.
	From string `yaml:"from" json:"from,omitempty"`
	// To is the list of recipient addresses for email channels.
	To []string `yaml:"to" json:"to,omitempty"`
	// TLSMode selects the TLS mode for email channels: "none",
	// "implicit", or "starttls".
	TLSMode string `yaml:"tls_mode" json:"tls_mode,omitempty"`
	// AuthType selects the SMTP auth type for email channels: "plain",
	// "login", or "cram-md5".
	AuthType string `yaml:"auth_type" json:"auth_type,omitempty"`
	// HTMLMode sends email bodies as HTML when true.
	HTMLMode bool `yaml:"html_mode" json:"html_mode,omitempty"`
}
