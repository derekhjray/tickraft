// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/tickraft/tickraft/pkg/api/httputil"
)

// CORS returns a middleware that handles cross-origin requests.
// It reflects the request Origin in Access-Control-Allow-Origin
// and handles preflight OPTIONS requests.
func CORS() app.HandlerFunc {
	return func(ctx context.Context, arc *app.RequestContext) {
		origin := string(arc.GetHeader("Origin"))
		if origin == "" {
			arc.Next(ctx)
			return
		}

		// Set CORS headers
		arc.Header("Access-Control-Allow-Origin", origin)
		arc.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		arc.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,"+httputil.HeaderAPIKey+","+httputil.HeaderRequestID)
		arc.Header("Access-Control-Allow-Credentials", "true")
		arc.Header("Access-Control-Max-Age", "86400")

		// Handle preflight
		if strings.EqualFold(string(arc.Request.Method()), "OPTIONS") {
			arc.Status(http.StatusNoContent)
			arc.Abort()
			return
		}

		arc.Next(ctx)
	}
}
