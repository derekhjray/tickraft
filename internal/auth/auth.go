// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package auth

import (
	"context"
	"fmt"
	"time"

	authcore "github.com/tickraft/tickraft/pkg/auth"
	"github.com/tickraft/tickraft/pkg/auth/jwt"
	"github.com/tickraft/tickraft/pkg/auth/password"
	"github.com/tickraft/tickraft/pkg/user"
)

// Authenticator provides authentication (identity verification).
type Authenticator interface {
	// Login authenticates a user and returns a login result containing the
	// token pair plus policy flags (e.g. MustChangePassword).
	Login(ctx context.Context, username, password string) (*authcore.LoginResult, error)
	// Verify validates a JWT token and returns the user.
	Verify(ctx context.Context, token string) (*user.User, error)
}

// Authorizer provides authorization (permission checking).
type Authorizer interface {
	// Can checks if a user has permission to perform an action on a asset.
	Can(ctx context.Context, user *user.User, action string, assetType string, assetID int64) (bool, error)
}

// Registrar provides user registration.
type Registrar interface {
	// Register creates a new user.
	Register(ctx context.Context, username, password, email string) (*user.User, error)
}

// Policy defines the permission checking strategy.
// The default implementation provides an RBAC policy.
type Policy interface {
	// Check returns whether the user with the given role is allowed to
	// perform the specified action on the asset type.
	Check(role int, action string, assetType string) bool
}

// jwtAuthenticator implements Authenticator using JWT for token operations.
type jwtAuthenticator struct {
	users      user.Store
	blacklist  authcore.BlacklistStore
	jwtMgr     *jwt.JWT
	bcryptCost int
}

// NewAuthenticator creates a new Authenticator backed by the given stores
// and JWT manager. bcryptCost of 0 means bcrypt.DefaultCost will be used.
func NewAuthenticator(users user.Store, blacklist authcore.BlacklistStore, jwtMgr *jwt.JWT, bcryptCost int) Authenticator {
	return &jwtAuthenticator{
		users:      users,
		blacklist:  blacklist,
		jwtMgr:     jwtMgr,
		bcryptCost: bcryptCost,
	}
}

// Login authenticates a user by username and password, returning a login
// result containing the token pair and policy flags on success.
func (a *jwtAuthenticator) Login(ctx context.Context, username, pwd string) (*authcore.LoginResult, error) {
	user, err := a.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, authcore.ErrUnauthorized
	}

	if err := password.Verify(user.PasswordHash, pwd); err != nil {
		return nil, authcore.ErrUnauthorized
	}

	// Build jwt.UserClaims from user data.
	//
	// The runtime is single-tenant; TenantID is left as the zero
	// value. The extended User model embeds user.User and
	// populates TenantID from its own augmented user type before issuing tokens.
	jwtClaims := jwt.UserClaims{
		UID:      user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	tokenPair, err := a.jwtMgr.GenerateTokenPair(jwtClaims)
	if err != nil {
		return nil, fmt.Errorf("auth: generate token pair: %w", err)
	}

	return &authcore.LoginResult{
		TokenPair:          &jwt.TokenPair{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken},
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// Verify validates a JWT token and returns the associated user.
func (a *jwtAuthenticator) Verify(ctx context.Context, token string) (*user.User, error) {
	claims, err := a.jwtMgr.ValidateToken(token, authcore.TokenTypeAccess)
	if err != nil {
		return nil, authcore.ErrUnauthorized
	}

	user, err := a.users.GetByID(ctx, claims.UID)
	if err != nil {
		return nil, authcore.ErrUnauthorized
	}

	return user, nil
}

// userRegistrar implements Registrar for user registration.
type userRegistrar struct {
	users      user.Store
	bcryptCost int
}

// NewRegistrar creates a new Registrar backed by the given user store.
// bcryptCost of 0 means bcrypt.DefaultCost will be used.
func NewRegistrar(users user.Store, bcryptCost int) Registrar {
	return &userRegistrar{
		users:      users,
		bcryptCost: bcryptCost,
	}
}

// Register creates a new user.
func (r *userRegistrar) Register(ctx context.Context, username, pwd, email string) (*user.User, error) {
	// Check if username already exists
	_, err := r.users.GetByUsername(ctx, username)
	if err == nil {
		return nil, authcore.ErrUserExists
	}

	// Hash the password
	hashedPwd, err := password.HashWithCost(pwd, r.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("auth: hash password: %w", err)
	}

	// Create the user record
	now := time.Now()
	id, err := r.users.Create(ctx, username, hashedPwd, email, authcore.RoleDeveloper)
	if err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	return &user.User{
		ID:           id,
		Username:     username,
		PasswordHash: hashedPwd,
		Role:         authcore.RoleDeveloper,
		Email:        email,
		CreatedAt:    now,
	}, nil
}

// rbacPolicy implements Policy with a role-based access control strategy.
// admin: full access to all resources.
// developer: read/write on tasks, devices, alerts; read on others.
// visitor: read-only on all resources.
type rbacPolicy struct {
	// rules maps role -> assetType -> set of allowed actions.
	rules map[int]map[string]map[string]bool
}

// newRBACPolicy creates the default RBAC policy.
func newRBACPolicy() *rbacPolicy {
	p := &rbacPolicy{
		rules: make(map[int]map[string]map[string]bool),
	}

	// Admin: full access
	p.rules[authcore.RoleAdmin] = map[string]map[string]bool{
		"*": {authcore.ActionRead: true, authcore.ActionWrite: true, authcore.ActionDelete: true},
	}

	// Developer: manage tasks, devices, alerts; read others
	devResources := map[string]map[string]bool{
		"task":   {authcore.ActionRead: true, authcore.ActionWrite: true, authcore.ActionDelete: true},
		"device": {authcore.ActionRead: true, authcore.ActionWrite: true, authcore.ActionDelete: false},
		"alert":  {authcore.ActionRead: true, authcore.ActionWrite: true, authcore.ActionDelete: false},
		"*":      {authcore.ActionRead: true, authcore.ActionWrite: false, authcore.ActionDelete: false},
	}
	p.rules[authcore.RoleDeveloper] = devResources

	// Visitor: read-only
	p.rules[authcore.RoleVisitor] = map[string]map[string]bool{
		"*": {authcore.ActionRead: true, authcore.ActionWrite: false, authcore.ActionDelete: false},
	}

	return p
}

// Check returns whether the given role is allowed to perform the action on the asset type.
func (p *rbacPolicy) Check(role int, action string, assetType string) bool {
	resourceRules, ok := p.rules[role]
	if !ok {
		return false
	}

	// Check asset-specific rules first
	if actions, found := resourceRules[assetType]; found {
		return actions[action]
	}

	// Fall back to wildcard rules
	if actions, found := resourceRules["*"]; found {
		return actions[action]
	}

	return false
}

// rbacAuthorizer implements Authorizer using the built-in RBAC policy.
type rbacAuthorizer struct {
	policy Policy
}

// NewAuthorizer creates a new Authorizer backed by the built-in RBAC policy.
// The runtime is single-tenant: all users belong to tenant 0 and
// permission checks are resolved by the default RBAC rules.
func NewAuthorizer() Authorizer {
	return &rbacAuthorizer{
		policy: newRBACPolicy(),
	}
}

// Can checks if a user has permission to perform an action on a asset.
func (a *rbacAuthorizer) Can(ctx context.Context, user *user.User, action string, assetType string, assetID int64) (bool, error) {
	if user == nil {
		return false, nil
	}

	return a.policy.Check(user.Role, action, assetType), nil
}

// DefaultPolicy returns the default RBAC policy.
func DefaultPolicy() Policy {
	return newRBACPolicy()
}
