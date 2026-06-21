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
	"github.com/tickraft/tickraft/pkg/api/handler/alert"
	"github.com/tickraft/tickraft/pkg/api/httputil"
)

// memoryAlertService is an in-memory implementation of alert.Service.
type memoryAlertService struct {
	mu         sync.RWMutex
	rules      map[int64]*alert.Rule
	records    map[int64]*alert.Record
	nextRuleID int64
}

// NewAlertService returns a new in-memory AlertService.
func NewAlertService() alert.Service {
	return &memoryAlertService{
		rules:   make(map[int64]*alert.Rule),
		records: make(map[int64]*alert.Record),
	}
}

// ListRules returns a page of alert rules and the total count.
func (s *memoryAlertService) ListRules(_ context.Context, page, size int) ([]alert.Rule, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]alert.Rule, 0, len(s.rules))
	for _, r := range s.rules {
		all = append(all, *r)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	page, size = httputil.ClampPaging(page, size)
	total := len(all)
	start, end := httputil.PageWindow(page, size, total)
	return all[start:end], int64(total), nil
}

// GetRule returns a single alert rule by ID.
func (s *memoryAlertService) GetRule(_ context.Context, id int64) (*alert.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.rules[id]
	if !ok {
		return nil, handler.ErrRuleNotFound
	}
	cp := *r
	return &cp, nil
}

// CreateRule creates a new alert rule from the given request.
func (s *memoryAlertService) CreateRule(_ context.Context, req *alert.Rule) (*alert.Rule, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextRuleID++
	now := time.Now()
	r := *req
	r.ID = s.nextRuleID
	r.CreatedAt = now
	r.UpdatedAt = now
	s.rules[r.ID] = &r
	cp := r
	return &cp, nil
}

// UpdateRule updates an existing alert rule identified by ID.
func (s *memoryAlertService) UpdateRule(_ context.Context, id int64, req *alert.Rule) (*alert.Rule, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.rules[id]
	if !ok {
		return nil, handler.ErrRuleNotFound
	}
	updated := *req
	updated.ID = id
	updated.CreatedAt = existing.CreatedAt
	updated.UpdatedAt = time.Now()
	s.rules[id] = &updated
	cp := updated
	return &cp, nil
}

// DeleteRule deletes an alert rule by ID.
func (s *memoryAlertService) DeleteRule(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[id]; !ok {
		return handler.ErrRuleNotFound
	}
	delete(s.rules, id)
	return nil
}

// ListRecords returns a page of alert records and the total count. The
// in-memory service does not generate records and always returns an empty
// result.
func (s *memoryAlertService) ListRecords(_ context.Context, _, _ int) ([]alert.Record, int64, error) {
	return []alert.Record{}, 0, nil
}

// GetRecord returns a single alert record by ID. The in-memory service does
// not generate records.
func (s *memoryAlertService) GetRecord(_ context.Context, _ int64) (*alert.Record, error) {
	return nil, handler.ErrRecordNotFound
}

// AcknowledgeRecord transitions the alert record identified by ID to the
// "acknowledged" status. The in-memory service does not generate records.
func (s *memoryAlertService) AcknowledgeRecord(_ context.Context, _ int64) (*alert.Record, error) {
	return nil, handler.ErrRecordNotFound
}

// ResolveRecord transitions the alert record identified by ID to the
// "resolved" status. The in-memory service does not generate records.
func (s *memoryAlertService) ResolveRecord(_ context.Context, _ int64) (*alert.Record, error) {
	return nil, handler.ErrRecordNotFound
}
