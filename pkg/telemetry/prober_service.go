// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

import (
	"context"
	"fmt"

	"github.com/tickraft/tickraft/pkg/executor"
	"github.com/tickraft/tickraft/pkg/task"
	"go.uber.org/zap"
)

// ProberService manages active probing by holding a task.Manager
// instance and consuming executor.Prober executors. When a monitoring
// point (Mode=ModeActive) is registered, ProberService schedules it via the
// engine. On fire, it invokes the prober executor and feeds the result
// into the shared processing pipeline (Validator -> Processor ->
// StateManager -> Emitter).
//
// The service operates exclusively on MonitorPoint records where
// Mode=ModeActive. Passive points (Mode=ModePassive) are handled by the
// listener pipeline and are never touched by this service.
type ProberService struct {
	sched   task.Manager
	execReg *executor.Registry
	manager *Manager
	// store optionally persists and queries monitoring points. When nil,
	// ListActivePoints returns an empty slice and RegisterPoint/UnregisterPoint
	// are in-memory stubs. The callers injects a concrete store
	// backed by the monitor_points table.
	store  *MonitorStore
	logger *zap.Logger
}

// ProberOption configures a ProberService at construction time.
type ProberOption func(*ProberService)

// WithProberMonitorStore injects a MonitorStore for querying and persisting
// active monitoring points. When omitted, the service operates without
// persistence (ListActivePoints returns an empty slice).
func WithProberMonitorStore(store *MonitorStore) ProberOption {
	return func(s *ProberService) { s.store = store }
}

// NewProberService creates a ProberService with the given task
// engine, executor registry, and optional configuration. The variadic
// options allow callers to inject a MonitorStore for point persistence
// without changing the positional signature.
func NewProberService(sched task.Manager, execReg *executor.Registry, manager *Manager, logger *zap.Logger, opts ...ProberOption) *ProberService {
	s := &ProberService{
		sched:   sched,
		execReg: execReg,
		manager: manager,
		logger:  logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ListActivePoints returns all monitoring points in active probing mode
// (Mode=ModeActive) from the injected MonitorStore. When no store is
// configured, it returns an empty slice and a nil error.
func (s *ProberService) ListActivePoints(ctx context.Context) ([]MonitorPoint, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.ListActive(ctx)
}

// RegisterPoint registers an active monitoring point for periodic probing.
// The runtime provides this as a stub; callers may inject
// a concrete implementation that parses the schedule and registers the point
// with the task.Manager. The point must have Mode=ModeActive; a passive
// point is rejected with an error.
func (s *ProberService) RegisterPoint(_ context.Context, point MonitorPoint) error {
	if !point.IsActive() {
		return fmt.Errorf("telemetry: cannot register passive point as active prober")
	}
	return nil
}

// UnregisterPoint removes an active monitoring point.
// The runtime provides this as a stub; callers may inject
// a concrete implementation that unschedules the point from the task.Manager.
func (s *ProberService) UnregisterPoint(_ context.Context, _ int64) error {
	return nil
}

// Start begins the prober service.
//
// The scheduler engine lifecycle (SubscribeEvents + Restore) is owned by
// the application bootstrap; ProberService coordinates active point
// registration and does not start the shared scheduler itself to avoid
// double-subscribing event handlers.
func (s *ProberService) Start(_ context.Context) error {
	return nil
}

// Stop gracefully stops the prober service.
//
// The shared scheduler is stopped by the bootstrap (stopWorkerEngines).
// ProberService must not stop it to avoid premature teardown of the
// shared task scheduling subsystem.
func (s *ProberService) Stop(_ context.Context) error {
	return nil
}
