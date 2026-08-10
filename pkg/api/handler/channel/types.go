// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package channel

import (
	"context"
	"time"
)

// Channel represents a notification channel definition managed
// through the CRUD API at /api/v1/prism/channels. The Config field carries
// a channel-type-specific JSON payload interpreted by the channel factory
// selected by Type. The open-source edition supports the "webhook" type.
type Channel struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`   // webhook (CE), extensible via SPI
	Config     string     `json:"config"` // JSON-encoded channel.Config payload
	Enabled    bool       `json:"enabled"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Service defines the operations for managing notification channels.
// The concrete implementation is injected via the WithChannelService
// RouteOption; when omitted, the handler package falls back to an in-memory
// implementation.
type Service interface {
	// ListChannels returns a page of notification channels and the total count.
	ListChannels(ctx context.Context, page, size int) ([]Channel, int64, error)
	// GetChannel returns a single notification channel by ID.
	GetChannel(ctx context.Context, id int64) (*Channel, error)
	// CreateChannel creates a new notification channel from the given request.
	CreateChannel(ctx context.Context, req *Channel) (*Channel, error)
	// UpdateChannel updates an existing notification channel identified by ID.
	UpdateChannel(ctx context.Context, id int64, req *Channel) (*Channel, error)
	// DeleteChannel deletes a notification channel by ID.
	DeleteChannel(ctx context.Context, id int64) error
	// TestChannel sends a test notification through the channel identified by
	// ID and returns an error describing any delivery failure. A nil error
	// means the test notification was delivered successfully.
	TestChannel(ctx context.Context, id int64) error
}
