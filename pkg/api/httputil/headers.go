// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package httputil

// HTTP header constants. All custom headers use the X-Tickraft- prefix per
// the global naming convention (00_tickraft_global.md §四.3.1).
const (
	// HeaderRequestID is the request tracing header, set by the RequestID
	// middleware and echoed in every response.
	HeaderRequestID = "X-Tickraft-Request-Id"

	// HeaderAPIKey is the API key authentication header, validated by the
	// APIKeyAuth middleware.
	HeaderAPIKey = "X-Tickraft-API-Key"
)
