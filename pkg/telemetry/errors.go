// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

import "errors"

var (
	// ErrAssetNotFound is returned when a asset is not found.
	ErrAssetNotFound = errors.New("telemetry: asset not found")
	// ErrInvalidConfig is returned when a collect configuration is invalid.
	ErrInvalidConfig = errors.New("telemetry: invalid config")
	// ErrValidationFailed is returned when a telemetry fails validation.
	ErrValidationFailed = errors.New("telemetry: validation failed")
	// ErrTenantMismatch is returned when the telemetry tenant does not match the asset tenant.
	ErrTenantMismatch = errors.New("telemetry: tenant mismatch")
	// ErrMetricLimitExceeded is returned when the number of metrics in a telemetry exceeds the limit.
	ErrMetricLimitExceeded = errors.New("telemetry: metric limit exceeded")
	// ErrLogLimitExceeded is returned when a log body exceeds the size limit.
	ErrLogLimitExceeded = errors.New("telemetry: log limit exceeded")
)
