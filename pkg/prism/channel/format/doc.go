// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package format provides a convenience factory around the i18n Formatter
// for use by notification channels in the tickraft alerting pipeline.
//
// Channels typically need to render an [alert.Event] into a localized
// [i18n.FormattedMessage] before delivering it. While the i18n package
// exposes the Formatter interface directly, channels benefit from a small
// facade that:
//
//   - Constructs a default Formatter from a Registry without each caller
//     repeating the NewDefaultFormatter boilerplate.
//   - Provides Build and BuildWithOpts helpers that pick sensible defaults
//     (Locale=i18n.DefaultLocale, Style=detailed) so simple channels can render with a single
//     call.
//
// The factory is stateless and safe for concurrent use. Extended channels
// may wrap the same Formatter to add channel-specific fields without
// reimplementing the rendering pipeline.
package format
