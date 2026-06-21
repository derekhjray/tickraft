// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package remediation provides an event-driven remediation module for the
// standalone single-process runtime.
//
// This package implements the standalone-runtime adaptation of the
// remediation design described in docs/modules/remediation.md. An
// in-process Manager subscribes to events on the event bus, matches them
// against persisted Rules, and dispatches remediation actions directly
// through the executor.Registry with TriggerType=event.
//
// # Single-Process Adaptation
//
// The standalone runtime runs as a single process. The distributed
// building blocks described by the remediation design — a Redis Stream
// backlog, a separate Worker process, and gRPC control-plane RPCs — are
// intentionally absent here:
//
//   - No Redis Stream: events are delivered in-process via the event bus.
//   - No distributed Worker: the Manager invokes executors directly.
//   - No gRPC: there is no control plane to drive.
//
// As a result this package only depends on the standard library, GORM,
// and other in-process kernel packages (pkg/event, pkg/circuitbreaker).
// To avoid an import cycle through pkg/executor -> pkg/asset -> pkg/db ->
// internal/remediation, the Manager does not import pkg/executor directly;
// instead it consumes the local Registry and Executor mirror interfaces
// defined in manager.go. The caller (typically in internal/service) wraps
// the concrete *executor.Registry in an adapter that satisfies the local
// Registry interface. All dispatching happens inside the Manager's event
// handler, and there is no cross-process communication of any kind.
//
// # Manager
//
// The [Manager] is the entry point of this package. Construct one with
// [New] and the WithStore, WithBus, WithRegistry, WithLogger, and
// WithQuota options, then call [Manager.Start] to subscribe to source
// events and begin dispatching, and [Manager.Stop] to drain in-flight
// dispatches and unsubscribe. Call [Manager.ReloadRules] after rule
// definitions change to refresh the in-memory rule cache without
// restarting the Manager.
//
// # Quota
//
// The runtime enforces a fixed quota of 5 remediation tasks, matching the
// internal/quota table entry keyed by TypeRemediation. Quota enforcement
// is owned by internal/quota; this package only persists rule and record
// rows and does not enforce the ceiling itself.
//
// # Persistence
//
// Rule definitions and execution Records are persisted in two SQLite
// tables, sys_remediation_rule and sys_remediation_record, registered
// with the package's Migrate function (see store.go). The Store interface
// abstracts CRUD access; the default implementation is the GORM-backed
// store returned by NewStore.
package remediation
