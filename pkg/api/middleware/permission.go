// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package middleware

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/tickraft/tickraft/internal/auth"
	"github.com/tickraft/tickraft/pkg/api/httputil"
	"github.com/tickraft/tickraft/pkg/errdefs"
)

// RequirePermission returns a Hertz middleware that checks if the authenticated
// user has the required permission. This middleware must be used after
// NewJWTAuth or another middleware that sets UserClaims.
//
// Permissions are validated directly from the JWT claims (the role field) via
// the default RBAC policy. This is a critical security measure:
// fail-closed, not fail-open.
func RequirePermission(action string, assetType string) app.HandlerFunc {
	return func(ctx context.Context, arc *app.RequestContext) {
		claims, ok := httputil.GetUserClaims(arc)
		if !ok || claims == nil {
			httputil.FailWithCode(arc, http.StatusUnauthorized, errdefs.CodeUnauthorized, "unauthorized")
			arc.Abort()
			return
		}

		if !checkPermission(ctx, claims.Role, action, assetType) {
			httputil.FailWithCode(arc, http.StatusForbidden, errdefs.CodeForbidden, "permission denied")
			arc.Abort()
			return
		}
		arc.Next(ctx)
	}
}

// checkPermission checks whether the given role permits the specified action on
// the asset type, validating directly from the JWT claims' role field via
// the default RBAC policy.
func checkPermission(c context.Context, role int, action string, assetType string) bool {
	return auth.DefaultPolicy().Check(role, action, assetType)
}
