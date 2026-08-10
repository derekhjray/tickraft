// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package console provides a reusable SPA static-file server utility for
// serving front-end assets from a dedicated HTTP listener.
//
// [Start] starts a static-file listener for an SPA with history-mode
// fallback: requests that do not match a static file fall back to
// index.html so client-side routing works without per-route server
// configuration. It is a no-op when rootDir is empty, so callers can wire
// it into an errgroup unconditionally and only expose the Console port
// when a separate asset directory is configured.
//
//	err := console.Start(ctx, ":8080", "/var/lib/tickraft/console")
//
// This package does NOT embed front-end assets; each edition (default
// tickraft, extended callers) embeds its own build output in its own
// internal/web package. This package only provides the shared HTTP serving
// utility that both editions can reuse.
package console
