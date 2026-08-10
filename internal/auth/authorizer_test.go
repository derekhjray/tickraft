// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package auth_test

import (
	"context"
	"testing"

	"github.com/tickraft/tickraft/internal/auth"
	authcore "github.com/tickraft/tickraft/pkg/auth"
	"github.com/tickraft/tickraft/pkg/user"
)

// ---------------------------------------------------------------------------
// Authorizer tests
// ---------------------------------------------------------------------------

func TestAuthorizer_DefaultRBAC(t *testing.T) {
	authz := auth.NewAuthorizer()
	ctx := context.Background()

	adminUser := &user.User{ID: 1, Role: authcore.RoleAdmin}
	devUser := &user.User{ID: 2, Role: authcore.RoleDeveloper}
	visitorUser := &user.User{ID: 3, Role: authcore.RoleVisitor}

	// Admin: full access
	allowed, err := authz.Can(ctx, adminUser, authcore.ActionWrite, "task", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected admin write on task to be allowed with default RBAC")
	}

	// Developer: can write tasks
	allowed, err = authz.Can(ctx, devUser, authcore.ActionWrite, "task", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected developer write on task to be allowed with default RBAC")
	}

	// Developer: cannot delete devices
	allowed, err = authz.Can(ctx, devUser, authcore.ActionDelete, "device", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected developer delete on device to be denied with default RBAC")
	}

	// Visitor: read-only
	allowed, err = authz.Can(ctx, visitorUser, authcore.ActionRead, "task", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected visitor read on task to be allowed with default RBAC")
	}

	allowed, err = authz.Can(ctx, visitorUser, authcore.ActionWrite, "task", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected visitor write on task to be denied with default RBAC")
	}
}

func TestAuthorizer_NilUser_ReturnsFalse(t *testing.T) {
	authz := auth.NewAuthorizer()
	ctx := context.Background()

	allowed, err := authz.Can(ctx, nil, authcore.ActionRead, "task", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected nil user to be denied")
	}
}
