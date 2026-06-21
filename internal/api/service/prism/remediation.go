// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package prism

import (
	"context"
	"errors"
	"net/http"

	"github.com/tickraft/tickraft/pkg/api/handler"
	"github.com/tickraft/tickraft/pkg/api/handler/remediation"
	"github.com/tickraft/tickraft/pkg/api/httputil"
	"github.com/tickraft/tickraft/pkg/errdefs"
	prismremediation "github.com/tickraft/tickraft/pkg/prism/remediation"
)

// RemediationService implements remediation.Service using the
// prism remediation store.
type RemediationService struct {
	store *prismremediation.Store
}

// NewRemediationService creates a RemediationService backed by the given
// store.
func NewRemediationService(store *prismremediation.Store) *RemediationService {
	return &RemediationService{store: store}
}

// ListRules returns a page of remediation rules and the total count.
func (s *RemediationService) ListRules(ctx context.Context, page, size int) ([]remediation.Rule, int64, error) {
	page, size = httputil.ClampPaging(page, size)
	models, total, err := s.store.List(ctx, page, size)
	if err != nil {
		return nil, 0, mapRemediationStoreError(err)
	}
	rules := make([]remediation.Rule, 0, len(models))
	for _, m := range models {
		rules = append(rules, remediationModelToHandler(m))
	}
	return rules, total, nil
}

// GetRule returns a single remediation rule by ID.
func (s *RemediationService) GetRule(ctx context.Context, id int64) (*remediation.Rule, error) {
	m, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, mapRemediationStoreError(err)
	}
	h := remediationModelToHandler(m)
	return &h, nil
}

// CreateRule creates a new remediation rule from the given request.
func (s *RemediationService) CreateRule(ctx context.Context, req *remediation.Rule) (*remediation.Rule, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	if req.Name == "" || req.TriggerEventType == "" || req.ExecutorType == "" {
		return nil, handler.ErrInvalidRequest
	}
	m := remediationHandlerToModel(req)
	if err := s.store.Create(ctx, m); err != nil {
		return nil, mapRemediationStoreError(err)
	}
	h := remediationModelToHandler(m)
	return &h, nil
}

// UpdateRule updates an existing remediation rule identified by ID.
func (s *RemediationService) UpdateRule(ctx context.Context, id int64, req *remediation.Rule) (*remediation.Rule, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, mapRemediationStoreError(err)
	}
	m := remediationHandlerToModel(req)
	m.ID = existing.ID
	m.CreatedAt = existing.CreatedAt
	if err := s.store.Update(ctx, m); err != nil {
		return nil, mapRemediationStoreError(err)
	}
	h := remediationModelToHandler(m)
	return &h, nil
}

// DeleteRule deletes a remediation rule by ID.
func (s *RemediationService) DeleteRule(ctx context.Context, id int64) error {
	if err := s.store.DeleteByID(ctx, id); err != nil {
		return mapRemediationStoreError(err)
	}
	return nil
}

// remediationModelToHandler converts a prismremediation.Rule persistence model
// into the handler-layer Rule DTO.
func remediationModelToHandler(m *prismremediation.Rule) remediation.Rule {
	return remediation.Rule{
		ID:                      m.ID,
		Name:                    m.Name,
		Description:             m.Description,
		AssetID:                 m.AssetID,
		TriggerEventType:        m.TriggerEventType,
		ConditionExpr:           m.ConditionExpr,
		ExecutorType:            m.ExecutorType,
		ExecutorConfig:          m.ExecutorConfig,
		Cooldown:                m.Cooldown,
		CircuitBreakerThreshold: m.CircuitBreakerThreshold,
		Enabled:                 m.Enabled,
		Status:                  m.Status,
		LastRunAt:               m.LastRunAt,
		CreatedAt:               m.CreatedAt,
		UpdatedAt:               m.UpdatedAt,
	}
}

// remediationHandlerToModel converts a handler-layer Rule DTO into
// a prismremediation.Rule persistence model ready for Create/Update.
func remediationHandlerToModel(r *remediation.Rule) *prismremediation.Rule {
	return &prismremediation.Rule{
		Name:                    r.Name,
		Description:             r.Description,
		AssetID:                 r.AssetID,
		TriggerEventType:        r.TriggerEventType,
		ConditionExpr:           r.ConditionExpr,
		ExecutorType:            r.ExecutorType,
		ExecutorConfig:          r.ExecutorConfig,
		Cooldown:                r.Cooldown,
		CircuitBreakerThreshold: r.CircuitBreakerThreshold,
		Enabled:                 r.Enabled,
	}
}

// mapRemediationStoreError translates a remediation store error into a
// handler-level ServiceError suitable for the API response layer.
func mapRemediationStoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, prismremediation.ErrRuleNotFound) {
		return handler.ErrRemediationRuleNotFound
	}
	if errors.Is(err, errdefs.ErrNotFound) {
		return handler.ErrRemediationRuleNotFound
	}
	return handler.NewServiceError(http.StatusInternalServerError, errdefs.CodeInternal, err.Error())
}
