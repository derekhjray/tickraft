// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package auth provides the authentication, authorization, and
// registration implementations used by the standalone runtime.
//
// The Authenticator, Authorizer, Registrar, and Policy interfaces and their
// constructors (NewAuthenticator, NewAuthorizer, NewRegistrar, DefaultPolicy)
// are internal abstractions not consumed by the extended repository.
package auth
