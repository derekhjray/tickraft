// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package prism

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tickraft/tickraft/pkg/api/handler"
	"github.com/tickraft/tickraft/pkg/api/handler/channel"
	"github.com/tickraft/tickraft/pkg/api/httputil"
	"github.com/tickraft/tickraft/pkg/errdefs"
	prismalert "github.com/tickraft/tickraft/pkg/prism/alert"
	prismchannel "github.com/tickraft/tickraft/pkg/prism/channel"
	"github.com/tickraft/tickraft/pkg/prism/channel/webhook"
)

// ChannelService implements channel.Service using the prism channel
// store.
type ChannelService struct {
	store *prismchannel.Store
}

// NewChannelService creates a ChannelService backed by the given store.
func NewChannelService(store *prismchannel.Store) *ChannelService {
	return &ChannelService{store: store}
}

// ListChannels returns a page of notification channels and the total count.
func (s *ChannelService) ListChannels(ctx context.Context, page, size int) ([]channel.Channel, int64, error) {
	page, size = httputil.ClampPaging(page, size)
	models, total, err := s.store.List(ctx, page, size)
	if err != nil {
		return nil, 0, mapChannelStoreError(err)
	}
	channels := make([]channel.Channel, 0, len(models))
	for _, m := range models {
		channels = append(channels, channelModelToHandler(m))
	}
	return channels, total, nil
}

// GetChannel returns a single notification channel by ID.
func (s *ChannelService) GetChannel(ctx context.Context, id int64) (*channel.Channel, error) {
	m, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, mapChannelStoreError(err)
	}
	h := channelModelToHandler(m)
	return &h, nil
}

// CreateChannel creates a new notification channel from the given request.
func (s *ChannelService) CreateChannel(ctx context.Context, req *channel.Channel) (*channel.Channel, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	if req.Name == "" || req.Type == "" || req.Config == "" {
		return nil, handler.ErrInvalidRequest
	}
	m := channelHandlerToModel(req)
	if err := s.store.Create(ctx, m); err != nil {
		return nil, mapChannelStoreError(err)
	}
	h := channelModelToHandler(m)
	return &h, nil
}

// UpdateChannel updates an existing notification channel identified by ID.
func (s *ChannelService) UpdateChannel(ctx context.Context, id int64, req *channel.Channel) (*channel.Channel, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, mapChannelStoreError(err)
	}
	m := channelHandlerToModel(req)
	m.ID = existing.ID
	m.CreatedAt = existing.CreatedAt
	if err := s.store.Update(ctx, m); err != nil {
		return nil, mapChannelStoreError(err)
	}
	h := channelModelToHandler(m)
	return &h, nil
}

// DeleteChannel deletes a notification channel by ID.
func (s *ChannelService) DeleteChannel(ctx context.Context, id int64) error {
	if err := s.store.DeleteByID(ctx, id); err != nil {
		return mapChannelStoreError(err)
	}
	return nil
}

// TestChannel sends a synthetic alert event through the channel identified by
// ID and returns an error describing any delivery failure.
func (s *ChannelService) TestChannel(ctx context.Context, id int64) error {
	m, err := s.store.GetByID(ctx, id)
	if err != nil {
		return mapChannelStoreError(err)
	}
	ch, err := buildChannelFromRecord(m)
	if err != nil {
		return handler.NewServiceError(http.StatusBadRequest, errdefs.CodeBadRequest, fmt.Sprintf("build channel: %v", err))
	}
	evt := prismalert.Event{
		Type:      prismalert.TypeMetric,
		Timestamp: time.Now(),
		Violations: []prismalert.Violation{
			{
				Kind:    prismalert.ViolationKindMetric,
				Message: "Test notification from tickraft",
			},
		},
	}
	return ch.Send(ctx, evt)
}

// channelModelToHandler converts a prismchannel.Record persistence model into
// the handler-layer Channel DTO.
func channelModelToHandler(m *prismchannel.Record) channel.Channel {
	return channel.Channel{
		ID:         m.ID,
		Name:       m.Name,
		Type:       m.Type,
		Config:     m.Config,
		Enabled:    m.Enabled,
		LastUsedAt: m.LastUsedAt,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// channelHandlerToModel converts a handler-layer Channel DTO into
// a prismchannel.Record persistence model ready for Create/Update.
func channelHandlerToModel(ch *channel.Channel) *prismchannel.Record {
	return &prismchannel.Record{
		Name:    ch.Name,
		Type:    ch.Type,
		Config:  ch.Config,
		Enabled: ch.Enabled,
	}
}

// mapChannelStoreError translates a channel store error into a handler-level
// ServiceError suitable for the API response layer.
func mapChannelStoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, prismchannel.ErrChannelNotFound) {
		return handler.ErrChannelNotFound
	}
	if errors.Is(err, errdefs.ErrNotFound) {
		return handler.ErrChannelNotFound
	}
	return handler.NewServiceError(http.StatusInternalServerError, errdefs.CodeInternal, err.Error())
}

// buildChannelFromRecord parses the record's Config JSON, looks up a
// registered channel factory by type, and constructs a runtime
// prismalert.Channel. When no factory is registered for the "webhook" type,
// the built-in webhook constructor is used as a fallback.
func buildChannelFromRecord(m *prismchannel.Record) (prismalert.Channel, error) {
	var cfg prismchannel.Config
	if err := json.Unmarshal([]byte(m.Config), &cfg); err != nil {
		return nil, fmt.Errorf("parse channel config: %w", err)
	}
	normalizedType := normalizeType(m.Type)
	if factory := prismchannel.LookupFactory(normalizedType); factory != nil {
		return factory(cfg)
	}
	if normalizedType == "webhook" {
		return buildWebhookChannel(cfg)
	}
	return nil, fmt.Errorf("unsupported channel type: %s", m.Type)
}

// buildWebhookChannel constructs a webhook prismalert.Channel from a
// prismchannel.Config.
func buildWebhookChannel(cfg prismchannel.Config) (prismalert.Channel, error) {
	whCfg := webhook.Config{
		URL:     cfg.URL,
		Headers: cfg.Headers,
	}
	if cfg.Timeout != "" {
		d, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, fmt.Errorf("parse webhook timeout: %w", err)
		}
		whCfg.Timeout = d
	}
	ch, err := webhook.New(whCfg)
	if err != nil {
		return nil, fmt.Errorf("create webhook channel: %w", err)
	}
	return ch, nil
}

// normalizeType lowercases a channel type name for case-insensitive factory
// lookup.
func normalizeType(t string) string {
	return strings.ToLower(t)
}
