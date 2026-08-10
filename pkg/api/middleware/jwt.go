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
	"github.com/tickraft/tickraft/pkg/auth"
	"github.com/tickraft/tickraft/pkg/auth/jwt"
	"github.com/tickraft/tickraft/pkg/errdefs"
)

// NewJWTAuth returns a Hertz middleware that validates JWT Bearer tokens using
// the provided JWT manager. If permission is non-empty, the middleware also
// checks that the authenticated user holds that permission via the default
// RBAC policy.
func NewJWTAuth(j *jwt.JWT, permission string) app.HandlerFunc {
	return func(ctx context.Context, arc *app.RequestContext) {
		authHeader := string(arc.GetHeader("Authorization"))
		if authHeader == "" {
			httputil.FailWithCode(arc, http.StatusUnauthorized, errdefs.CodeUnauthorized, "missing authorization header")
			arc.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httputil.FailWithCode(arc, http.StatusUnauthorized, errdefs.CodeUnauthorized, "invalid authorization header format")
			arc.Abort()
			return
		}

		token := parts[1]
		claims, err := j.ValidateToken(token, auth.TokenTypeAccess)
		if err != nil {
			httputil.FailWithCode(arc, http.StatusUnauthorized, errdefs.CodeUnauthorized, "invalid or expired token")
			arc.Abort()
			return
		}

		// jwt.ValidateToken returns *jwt.UserClaims, which is the type
		// expected by httputil.SetUserClaims.
		httputil.SetUserClaims(arc, claims)

		// If a permission is specified, check it.
		if permission != "" {
			if !checkPermission(ctx, claims.Role, permission, "") {
				httputil.FailWithCode(arc, http.StatusForbidden, errdefs.CodeForbidden, "permission denied")
				arc.Abort()
				return
			}
		}

		arc.Next(ctx)
	}
}
