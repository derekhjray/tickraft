// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/tickraft/tickraft/pkg/api/handler"
	"github.com/tickraft/tickraft/pkg/api/handler/channel"
	"github.com/tickraft/tickraft/pkg/api/httputil"
)

// memoryChannelService is an in-memory implementation of
// channel.Service.
type memoryChannelService struct {
	mu            sync.RWMutex
	channels      map[int64]*channel.Channel
	nextChannelID int64
}

// NewChannelService returns a new in-memory ChannelService.
func NewChannelService() channel.Service {
	return &memoryChannelService{channels: make(map[int64]*channel.Channel)}
}

// ListChannels returns a page of notification channels and the total count.
func (s *memoryChannelService) ListChannels(_ context.Context, page, size int) ([]channel.Channel, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]channel.Channel, 0, len(s.channels))
	for _, c := range s.channels {
		all = append(all, *c)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	page, size = httputil.ClampPaging(page, size)
	total := len(all)
	start, end := httputil.PageWindow(page, size, total)
	return all[start:end], int64(total), nil
}

// GetChannel returns a single notification channel by ID.
func (s *memoryChannelService) GetChannel(_ context.Context, id int64) (*channel.Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.channels[id]
	if !ok {
		return nil, handler.ErrChannelNotFound
	}
	cp := *c
	return &cp, nil
}

// CreateChannel creates a new notification channel from the given request.
func (s *memoryChannelService) CreateChannel(_ context.Context, req *channel.Channel) (*channel.Channel, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextChannelID++
	now := time.Now()
	c := *req
	c.ID = s.nextChannelID
	c.CreatedAt = now
	c.UpdatedAt = now
	s.channels[c.ID] = &c
	cp := c
	return &cp, nil
}

// UpdateChannel updates an existing notification channel identified by ID.
func (s *memoryChannelService) UpdateChannel(_ context.Context, id int64, req *channel.Channel) (*channel.Channel, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.channels[id]
	if !ok {
		return nil, handler.ErrChannelNotFound
	}
	updated := *req
	updated.ID = id
	updated.CreatedAt = existing.CreatedAt
	updated.UpdatedAt = time.Now()
	s.channels[id] = &updated
	cp := updated
	return &cp, nil
}

// DeleteChannel deletes a notification channel by ID.
func (s *memoryChannelService) DeleteChannel(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.channels[id]; !ok {
		return handler.ErrChannelNotFound
	}
	delete(s.channels, id)
	return nil
}

// TestChannel sends a test notification through the channel identified by ID.
// The in-memory service performs no actual delivery; it only validates that
// the channel exists.
func (s *memoryChannelService) TestChannel(_ context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.channels[id]; !ok {
		return handler.ErrChannelNotFound
	}
	return nil
}
