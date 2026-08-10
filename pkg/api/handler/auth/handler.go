// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/tickraft/tickraft/pkg/api"
	"github.com/tickraft/tickraft/pkg/api/httputil"
	"github.com/tickraft/tickraft/pkg/errdefs"
)

// Handler exposes authentication and API-key management endpoints.
// It is injected via the WithAuthService RouteOption and registered on
// the /api/v1/auth route group.
type Handler struct {
	svc Service
}

// NewHandler creates a new auth Handler backed by the given service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// loginRequest is the request body for the login endpoint.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// tokenData is the response data for endpoints that return tokens.
type tokenData struct {
	AccessToken        string `json:"access_token,omitempty"`
	RefreshToken       string `json:"refresh_token,omitempty"`
	MustChangePassword bool   `json:"must_change_password,omitempty"`
	MFARequired        bool   `json:"mfa_required,omitempty"`
	MFATicket          string `json:"mfa_ticket,omitempty"`
}

// refreshRequest is the request body for the refresh endpoint.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// changePasswordRequest is the request body for the change-password endpoint.
type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// createAPIKeyRequest is the request body for the create-apikey endpoint.
type createAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiredAt *time.Time `json:"expired_at,omitempty"`
}

// apiKeyData is the response data for the create-apikey endpoint.
type apiKeyData struct {
	RawKey   string     `json:"raw_key"`
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Prefix   string     `json:"key_prefix"`
	Status   int        `json:"status"`
	CreateAt time.Time  `json:"created_at"`
	ExpireAt *time.Time `json:"expired_at,omitempty"`
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(ctx context.Context, c *app.RequestContext) {
	var req loginRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	if req.Username == "" || req.Password == "" {
		api.FailWithCode(c, 400, errdefs.CodeBadRequest, "username and password are required")
		return
	}

	tokenPair, err := h.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		api.Fail(c, err)
		return
	}

	api.Success(c, tokenData{
		AccessToken:        tokenPair.AccessToken,
		RefreshToken:       tokenPair.RefreshToken,
		MustChangePassword: tokenPair.MustChangePassword,
		MFARequired:        tokenPair.MFARequired,
		MFATicket:          tokenPair.MFATicket,
	})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(ctx context.Context, c *app.RequestContext) {
	claims, ok := api.GetUserClaims(c)
	if !ok || claims == nil {
		api.FailWithCode(c, 401, errdefs.CodeUnauthorized, "unauthorized")
		return
	}

	// Extract access token JTI from the validated claims.
	accessJTI := claims.JTI
	var accessExpireAt time.Time
	if accessJTI != "" {
		accessExpireAt = time.Now().Add(2 * time.Hour)
	}

	var refreshJTI string
	var refreshExpireAt time.Time

	if err := h.svc.Logout(ctx, accessJTI, refreshJTI, accessExpireAt, refreshExpireAt); err != nil {
		api.Fail(c, err)
		return
	}

	api.Success(c, nil)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(ctx context.Context, c *app.RequestContext) {
	var req refreshRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	if req.RefreshToken == "" {
		api.FailWithCode(c, 400, errdefs.CodeBadRequest, "refresh_token is required")
		return
	}

	tokenPair, err := h.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		api.Fail(c, err)
		return
	}

	api.Success(c, tokenData{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// ChangePassword handles POST /api/v1/auth/password.
func (h *Handler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	claims, ok := api.GetUserClaims(c)
	if !ok || claims == nil {
		api.FailWithCode(c, 401, errdefs.CodeUnauthorized, "unauthorized")
		return
	}

	var req changePasswordRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		api.FailWithCode(c, 400, errdefs.CodeBadRequest, "old_password and new_password are required")
		return
	}

	if err := h.svc.ChangePassword(ctx, claims.UID, req.OldPassword, req.NewPassword, claims.JTI); err != nil {
		api.Fail(c, err)
		return
	}

	api.Success(c, nil)
}

// CreateAPIKey handles POST /api/v1/auth/apikeys.
func (h *Handler) CreateAPIKey(ctx context.Context, c *app.RequestContext) {
	claims, ok := api.GetUserClaims(c)
	if !ok || claims == nil {
		api.FailWithCode(c, 401, errdefs.CodeUnauthorized, "unauthorized")
		return
	}

	var req createAPIKeyRequest
	if !api.BindAndValidate(c, &req) {
		return
	}

	if req.Name == "" {
		api.FailWithCode(c, 400, errdefs.CodeBadRequest, "name is required")
		return
	}

	rawKey, info, err := h.svc.CreateAPIKey(ctx, req.Name, req.ExpiredAt)
	if err != nil {
		api.Fail(c, err)
		return
	}

	_ = claims

	api.Success(c, apiKeyData{
		RawKey:   rawKey,
		ID:       info.ID,
		Name:     info.Name,
		Prefix:   info.KeyPrefix,
		Status:   info.Status,
		CreateAt: info.CreatedAt,
		ExpireAt: info.ExpiredAt,
	})
}

// ListAPIKeys handles GET /api/v1/auth/apikeys.
func (h *Handler) ListAPIKeys(ctx context.Context, c *app.RequestContext) {
	page, size := httputil.ParsePaging(c)
	keys, total, err := h.svc.ListAPIKeys(ctx, page, size)
	if err != nil {
		api.Fail(c, err)
		return
	}

	api.SuccessPage(c, keys, total, page, size)
}

// RevokeAPIKey handles DELETE /api/v1/auth/apikeys/:id.
func (h *Handler) RevokeAPIKey(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	if idStr == "" {
		api.FailWithCode(c, 400, errdefs.CodeBadRequest, "api key id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.FailWithCode(c, 400, errdefs.CodeBadRequest, "invalid api key id")
		return
	}

	if err := h.svc.RevokeAPIKey(ctx, id); err != nil {
		api.Fail(c, err)
		return
	}

	api.Success(c, nil)
}
