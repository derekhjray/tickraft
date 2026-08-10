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
	"github.com/tickraft/tickraft/pkg/api/handler/remediation"
	"github.com/tickraft/tickraft/pkg/api/httputil"
)

// memoryRemediationRuleService is an in-memory implementation of
// remediation.Service.
type memoryRemediationRuleService struct {
	mu     sync.RWMutex
	rules  map[int64]*remediation.Rule
	nextID int64
}

// NewRemediationRuleService returns a new in-memory RemediationRuleService.
func NewRemediationRuleService() remediation.Service {
	return &memoryRemediationRuleService{rules: make(map[int64]*remediation.Rule)}
}

// ListRules returns a page of remediation rules and the total count.
func (s *memoryRemediationRuleService) ListRules(_ context.Context, page, size int) ([]remediation.Rule, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]remediation.Rule, 0, len(s.rules))
	for _, r := range s.rules {
		all = append(all, *r)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	page, size = httputil.ClampPaging(page, size)
	total := len(all)
	start, end := httputil.PageWindow(page, size, total)
	return all[start:end], int64(total), nil
}

// GetRule returns a single remediation rule by ID.
func (s *memoryRemediationRuleService) GetRule(_ context.Context, id int64) (*remediation.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.rules[id]
	if !ok {
		return nil, handler.ErrRemediationRuleNotFound
	}
	cp := *r
	return &cp, nil
}

// CreateRule creates a new remediation rule from the given request.
func (s *memoryRemediationRuleService) CreateRule(_ context.Context, req *remediation.Rule) (*remediation.Rule, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := time.Now()
	r := *req
	r.ID = s.nextID
	r.CreatedAt = now
	r.UpdatedAt = now
	s.rules[r.ID] = &r
	cp := r
	return &cp, nil
}

// UpdateRule updates an existing remediation rule identified by ID.
func (s *memoryRemediationRuleService) UpdateRule(_ context.Context, id int64, req *remediation.Rule) (*remediation.Rule, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.rules[id]
	if !ok {
		return nil, handler.ErrRemediationRuleNotFound
	}
	updated := *req
	updated.ID = id
	updated.CreatedAt = existing.CreatedAt
	updated.UpdatedAt = time.Now()
	s.rules[id] = &updated
	cp := updated
	return &cp, nil
}

// DeleteRule deletes a remediation rule by ID.
func (s *memoryRemediationRuleService) DeleteRule(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[id]; !ok {
		return handler.ErrRemediationRuleNotFound
	}
	delete(s.rules, id)
	return nil
}
