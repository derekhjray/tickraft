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

// ScopeAtlas is the JWT region claim value that marks a token as
// scoped for atlas access. Tokens without this value are rejected by
// NewAtlasJWTAuth with a 403 Forbidden.
const ScopeAtlas = "atlas"

// Authorizer checks whether an atlas user holds a specific
// permission. The atlas repository implements this interface
// to provide RBAC checks backed by the role-permission database tables.
//
// If no authorizer is provided to NewAtlasJWTAuth, permission
// checks are skipped and the middleware operates in
// authentication-only mode.
type Authorizer interface {
	// Authorize returns true if the user identified by userID has been
	// granted the given permission code (e.g. "tenant:write").
	Authorize(ctx context.Context, userID int64, permission string) bool
}

// NewAtlasJWTAuth returns a Hertz middleware that validates
// JWT Bearer tokens for the atlas API. The middleware performs the
// following checks in order:
//
//  1. Extracts the Bearer token from the Authorization header.
//  2. Validates the JWT signature, expiry, and blacklist status.
//  3. Verifies that the token's region claim equals "atlas"; otherwise
//     returns 403.
//  4. If permission is non-empty and authorizer is non-nil, checks that the
//     user holds the required permission; otherwise returns 403.
//  5. Writes the validated UserClaims to the request context for downstream
//     handlers.
//
// When authorizer is nil, step 4 is skipped, allowing the default build
// to use this middleware without an RBAC backend.
func NewAtlasJWTAuth(
	jwtMgr *jwt.JWT,
	permission string,
	authorizer Authorizer,
) app.HandlerFunc {
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
		claims, err := jwtMgr.ValidateToken(token, auth.TokenTypeAccess)
		if err != nil {
			httputil.FailWithCode(arc, http.StatusUnauthorized, errdefs.CodeUnauthorized, "invalid or expired token")
			arc.Abort()
			return
		}

		// Enforce atlas scope: only tokens with region="atlas"
		// may access /api/v1/atlas/* routes.
		if claims.Region != ScopeAtlas {
			httputil.FailWithCode(arc, http.StatusForbidden, errdefs.CodeForbidden, "token scope is not atlas")
			arc.Abort()
			return
		}

		// If a permission is required and an authorizer is configured,
		// perform the RBAC check. Fail-closed when the authorizer is set
		// but the check fails.
		if permission != "" && authorizer != nil {
			if !authorizer.Authorize(ctx, claims.UID, permission) {
				httputil.FailWithCode(arc, http.StatusForbidden, errdefs.CodeForbidden, "permission denied")
				arc.Abort()
				return
			}
		}

		// jwt.ValidateToken returns *jwt.UserClaims, which is the type
		// expected by httputil.SetUserClaims.
		httputil.SetUserClaims(arc, claims)
		arc.Next(ctx)
	}
}
