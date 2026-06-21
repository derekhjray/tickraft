// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package alert defines the shared alert domain objects consumed by both the
// prism alert engine (pkg/prism/alert) and the
// governance middleware (pkg/prism/governance).
//
// It lives under internal/ so external editions cannot import it directly:
// they only reference these types through the re-export aliases in
// pkg/prism/alert. Keeping the shared objects here breaks what would otherwise
// be a circular dependency between the alert engine (which consumes
// governance.Middleware) and the governance package (which references the
// alert Event).
package alert
