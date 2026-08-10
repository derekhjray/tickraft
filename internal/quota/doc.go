// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package quota encodes the fixed resource ceilings the runtime enforces.
//
// # Positioning
//
// The package was split out of pkg/auth (its previous location was
// pkg/auth/quota.go) so that cross-domain callers can reference the
// quota constants without pulling in authentication logic. The split
// also reflects the deployment reality: the runtime has its
// own fixed quota policy, while the callers may have a completely
// independent license-driven quota system in internal/account/quota.go.
// The two are mutually exclusive at the source level (the extended
// edition does not import this package), so keeping the
// quota in pkg/auth created the false impression that auth owns quota
// enforcement.
//
// # Hard constraints
//
//   - This package is internal: it is not importable by the downstream
//     extended repository. The callers defines its own quota
//     types and constants in internal/account/quota.go.
//   - The constants here are the stable public contract of the
//     runtime. Source compilation remains the documented
//     extension point for lifting these ceilings.
package quota
