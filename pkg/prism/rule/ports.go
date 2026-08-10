// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package rule

import "context"

// Lister is the consumer-side read port over the rule persistence
// layer, consumed by Engine.Reload. It is defined at the consumer side
// per code-architecture.md §4.3.2 ("accept interfaces, return
// structs"): the producer (store.go) returns a concrete *Store, and
// the Engine accepts only the narrow read shape it needs so tests can
// substitute a stub that implements ListEnabled alone.
type Lister interface {
	// ListEnabled returns enabled rules for the given scene, ordered
	// for deterministic evaluation. See Store.ListEnabled for the full
	// semantics.
	ListEnabled(ctx context.Context, tenantID int64, scene Scene) ([]Record, error)
}
