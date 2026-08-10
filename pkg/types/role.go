// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package types

// Role is the cross-domain user role for authentication and authorization.
//
// Previously this concept was expressed as bare untyped int constants
// (`RoleVisitor`, `RoleDeveloper`, `RoleAdmin`) in pkg/auth/types.go,
// matched against `model.User.Role int` and `jwt.Claims.Role int`.
// Centralizing the type here gives the compiler a way to detect
// mismatched role values at the call site and documents the role
// semantics in a single location.
//
// Migration note: pkg/auth exposes a type alias and re-exports the
// constants so existing callers using auth.RoleVisitor etc. continue to
// compile unchanged.
type Role int

const (
	// RoleVisitor represents a read-only user role (viewer).
	RoleVisitor Role = 0
	// RoleDeveloper represents a user with read/write access to tasks,
	// devices, and alerts.
	RoleDeveloper Role = 1
	// RoleAdmin represents a user with full access to all resources.
	RoleAdmin Role = 2
)
