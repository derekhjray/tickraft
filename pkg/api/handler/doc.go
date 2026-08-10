// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package handler hosts the HTTP handlers for the tickraft API
// server.
//
// The package is organized around a single RegisterRoutes entry point that
// wires every endpoint onto an [*api.Server] via a set of
// [RouteOption] values. Each option injects either a middleware
// (app.HandlerFunc) or a service implementation, keeping the handler package
// free of any direct dependency on pkg/auth, pkg/auth/jwt, or pkg/cache.
//
// # Service SPI
//
// Each domain module exposes a service interface and an in-memory default:
//
//   - auth.Service: authentication and API-key management.
//   - task.Service: scheduled task CRUD and execution history. The
//     in-memory default (NewMemoryTaskService) is suitable for the
//     runtime; NewSchedulerTaskService wires the real scheduler
//     engine and persistent stores.
//   - AlertService: alert rule and record management. The in-memory default
//     (NewMemoryAlertService) is the default;
//     NewPrismAlertService wires the real alert engine and persistent
//     stores.
//   - SystemService: system configuration and info.
//   - telemetry.Service: telemetry collection task CRUD.
//
// Extended builds inject real services through the corresponding
// With*Service option before [RegisterRoutes] is called.
//
// # Handler types
//
// Handler-local request/response types (Task, Execution, AlertRule,
// AlertRecord, SystemConfig, telemetry.Task, Telemetry, etc.) are defined in
// types.go. They decouple the API contract from the internal model package
// so the handler package stays testable without a database.
//
// # Purity
//
// The handler package depends only on the Go standard library, hertz,
// pkg/api, and the tickraft packages it receives as injected
// dependencies. It does not import any third-party observability, RPC, or
// caching client libraries.
package handler
