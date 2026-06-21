// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package taskconv provides bidirectional conversion between the
// handler-layer task DTO (pkg/api/handler/task) and the task domain
// task (pkg/task). It delegates to the canonical conversion functions
// defined in the handler task package so that service-layer code can use a
// dedicated taskconv namespace without duplicating conversion logic.
package taskconv

import (
	"github.com/tickraft/tickraft/pkg/api/handler/task"
	domaintask "github.com/tickraft/tickraft/pkg/task"
)

// DomainTaskToHandler converts a task domain task into a handler-layer Task.
// A nil input returns nil.
func DomainTaskToHandler(t *domaintask.Task) *task.Task {
	return task.DomainTaskToHandler(t)
}

// HandlerToDomainTask converts a handler-layer Task into a task domain task.
// A nil input returns nil.
func HandlerToDomainTask(t *task.Task) *domaintask.Task {
	return task.HandlerToDomainTask(t)
}
