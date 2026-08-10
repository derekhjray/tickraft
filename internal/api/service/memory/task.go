// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package memory provides in-memory implementations of the handler-layer
// Service interfaces for testing and development debugging only.
//
// These implementations are NOT used in production code paths. The
// production server registers concrete service implementations (backed by
// the scheduler engine, prism engine, or database) via router options.
// When a service option is omitted, the corresponding route group is
// simply not registered rather than falling back to these stubs.
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/tickraft/tickraft/pkg/api/handler"
	"github.com/tickraft/tickraft/pkg/api/handler/task"
	"github.com/tickraft/tickraft/pkg/api/httputil"
)

// memoryTaskService is an in-memory implementation of task.Service.
type memoryTaskService struct {
	mu     sync.RWMutex
	tasks  map[int64]*task.Task
	nextID int64
}

// NewTaskService returns a new in-memory TaskService.
func NewTaskService() task.Service {
	return &memoryTaskService{tasks: make(map[int64]*task.Task)}
}

// ListTasks returns a page of tasks matching the given filter and the total
// count. A zero-value Filter returns all tasks.
func (s *memoryTaskService) ListTasks(_ context.Context, page, size int, filter task.Filter) ([]task.Task, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matched := make([]task.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if !matchTaskFilter(t, filter) {
			continue
		}
		matched = append(matched, *t)
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })

	page, size = httputil.ClampPaging(page, size)
	total := len(matched)
	start, end := httputil.PageWindow(page, size, total)
	return matched[start:end], int64(total), nil
}

// GetTask returns a single task by ID.
func (s *memoryTaskService) GetTask(_ context.Context, id int64) (*task.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, handler.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}

// CreateTask creates a new task from the given request.
func (s *memoryTaskService) CreateTask(_ context.Context, req *task.Task) (*task.Task, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := time.Now()
	t := *req
	t.ID = s.nextID
	t.CreatedAt = now
	t.UpdatedAt = now
	s.tasks[t.ID] = &t
	cp := t
	return &cp, nil
}

// UpdateTask updates an existing task identified by ID.
func (s *memoryTaskService) UpdateTask(_ context.Context, id int64, req *task.Task) (*task.Task, error) {
	if req == nil {
		return nil, handler.ErrInvalidRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.tasks[id]
	if !ok {
		return nil, handler.ErrTaskNotFound
	}
	updated := *req
	updated.ID = id
	updated.CreatedAt = existing.CreatedAt
	updated.UpdatedAt = time.Now()
	s.tasks[id] = &updated
	cp := updated
	return &cp, nil
}

// DeleteTask deletes a task by ID.
func (s *memoryTaskService) DeleteTask(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return handler.ErrTaskNotFound
	}
	delete(s.tasks, id)
	return nil
}

// TriggerTask triggers an immediate execution of a task.
func (s *memoryTaskService) TriggerTask(_ context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.tasks[id]; !ok {
		return handler.ErrTaskNotFound
	}
	return nil
}

// PauseTask pauses a task by removing it from the scheduling wheel.
func (s *memoryTaskService) PauseTask(_ context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.tasks[id]; !ok {
		return handler.ErrTaskNotFound
	}
	return nil
}

// ResumeTask resumes a paused task by re-adding it to the scheduling wheel.
func (s *memoryTaskService) ResumeTask(_ context.Context, id int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.tasks[id]; !ok {
		return handler.ErrTaskNotFound
	}
	return nil
}

// ListExecutions returns a page of executions for a task and the total count.
// The in-memory service does not track executions and always returns an empty
// result.
func (s *memoryTaskService) ListExecutions(_ context.Context, _ int64, _, _ int) ([]task.Execution, int64, error) {
	return []task.Execution{}, 0, nil
}

// GetExecution returns a single execution record by ID. The in-memory service
// does not track executions.
func (s *memoryTaskService) GetExecution(_ context.Context, _ int64) (*task.Execution, error) {
	return nil, handler.ErrExecutionNotFound
}

// CopyTask creates a new task by cloning the configuration of an existing
// task identified by id. The new task is assigned a fresh ID and the given
// name; an empty name defaults to "<source name> (copy)".
func (s *memoryTaskService) CopyTask(_ context.Context, id int64, newName string) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	source, ok := s.tasks[id]
	if !ok {
		return nil, handler.ErrTaskNotFound
	}
	s.nextID++
	cp := *source
	cp.ID = s.nextID
	if newName == "" {
		newName = source.Name + " (copy)"
	}
	cp.Name = newName
	now := time.Now()
	cp.CreatedAt = now
	cp.UpdatedAt = now
	s.tasks[cp.ID] = &cp
	result := cp
	return &result, nil
}

// GetExecutionStats returns aggregated execution statistics for the given
// time range. The in-memory service does not track executions and always
// returns a zero-valued result.
func (s *memoryTaskService) GetExecutionStats(_ context.Context, _, _ time.Time) (task.ExecutionStats, error) {
	return task.ExecutionStats{}, nil
}

// matchTaskFilter reports whether t matches the optional filtering criteria.
// A zero-value filter matches all tasks.
func matchTaskFilter(t *task.Task, filter task.Filter) bool {
	if filter.Group != "" && t.Group != filter.Group {
		return false
	}
	if len(filter.Tags) > 0 && !hasAnyTag(t.Tags, filter.Tags) {
		return false
	}
	return true
}

// hasAnyTag reports whether have contains at least one entry in want.
func hasAnyTag(have, want []string) bool {
	for _, w := range want {
		for _, h := range have {
			if h == w {
				return true
			}
		}
	}
	return false
}
