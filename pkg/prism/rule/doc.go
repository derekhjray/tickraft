// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package rule implements the unified rule matching engine for the
// the tickraft repository. It compiles user-authored
// expressions once into reusable bytecode programs and evaluates them
// concurrently across four scenes: task routing, probe-result
// matching, metric-alert filtering, and remediation selection.
//
// The engine is the foundation shared by the alert
// dispatch path and the callers's self-healing workflows. It
// depends exclusively on the public packages of the tickraft
// repository (pkg/asset, pkg/event, pkg/executor, pkg/prism/alert,
// pkg/task, pkg/telemetry, pkg/types) and never imports any
// internal/ directory.
//
// Scenes:
//   - SceneTask: filters scheduler tasks prior to dispatch.
//   - SceneProbe: matches collector probe results and passive reports.
//   - SceneMetric: pre-filters prism metric and log alert events.
//   - SceneRemediation: selects remediation workflows for triggering
//     alert events.
//
// The Compiler enforces a sandbox (builtin whitelist, MaxNodes, and a
// MaxComparisons limit) so an untrusted rule expression cannot
// exhaust process resources. The default MaxComparisons limit is
// 3; callers can override it via CompilerConfig when constructing a
// Compiler or Engine.
package rule
