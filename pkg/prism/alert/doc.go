// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package alert provides the alert rule evaluation engine and notification
// channel dispatcher.
//
// The prism alert engine subscribes to collector alert events (metric and log alerts)
// published on the event bus, evaluates registered alert rules against each
// event, and dispatches matching alerts to registered notification channels
// through a bounded worker pool.
//
// Notification channel implementations live in the [github.com/tickraft/tickraft/pkg/prism/channel]
// package and its sub-packages. Channels are registered programmatically via
// Engine.AddChannel. Use [github.com/tickraft/tickraft/pkg/prism/channel.LoadChannels]
// to construct channels from a configuration string (URL or JSON/YAML file
// path).
package alert
