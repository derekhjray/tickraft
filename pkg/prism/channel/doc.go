// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package channel is the home for all notification channel implementations
// in the tickraft alerting pipeline.
//
// A channel is anything that delivers an alert to an external system. The
// channel interface is intentionally not redefined here: every channel
// implementation satisfies the [github.com/tickraft/tickraft/pkg/prism/alert.Channel]
// interface, which requires:
//
//   - Send(ctx context.Context, alert alert.Event) error
//   - Name() string
//
// Reusing the alert.Channel interface keeps the dispatch engine
// (alert.Engine) agnostic to where channels are declared, so channels may
// live in this package without forcing a cyclic dependency on alert.
//
// # Built-in channels
//
// The build ships webhook and email channels. The format
// sub-package provides a convenience Formatter facade backed by the
// built-in i18n bundle for use by channels that need localized rendering.
// Extensions register additional channels via the extension SPI:
//
//   - sms
//   - wecom (WeChat Work)
//   - dingtalk
//   - slack
//   - teams (Microsoft Teams)
//
// # Channel loading
//
// Channel configuration is parsed by the service layer
// (internal/service.LoadChannels), which reads a JSON or YAML config file
// or an HTTP URL and constructs alert.Channel instances. The Config struct
// and Factory type defined in this package are used by the loader;
// extension SPI factories are registered via Register and dispatched by
// LookupFactory.
//
// # Resilience
//
// Channels are expected to fail occasionally (network blips, rate limits,
// upstream outages). The pkg/circuitbreaker package provides a
// three-state (Closed / Open / HalfOpen) circuit breaker that channel
// implementations embed to avoid hammering a degraded endpoint. When the
// breaker is open, Send short-circuits and returns ErrCircuitOpen so
// the prism alert engine can skip retries and log the suppression.
//
// All built-in channels enforce three additional resilience guarantees:
//
//   - Timeout control: each Send is bounded by a configurable Timeout via
//     context.WithTimeout, so a slow upstream cannot block the prism alert
//     engine indefinitely.
//   - Panic isolation: a panic in any sub-operation (marshalling, dial,
//     SMTP/HTTP I/O, formatting) is recovered and returned as a
//     non-retryable SendError, so the prism alert engine is never crashed.
//   - Retry with backoff: transient failures (5xx, network errors, 4xx
//     SMTP) are retried with exponential backoff and full jitter;
//     permanent failures (4xx HTTP, 5xx SMTP, auth errors) fail fast.
//
// Context cancellations and programming errors (e.g. JSON marshal
// failures) do not count against the circuit breaker.
package channel
