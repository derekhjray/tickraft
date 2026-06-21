// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package rule

import (
	"time"

	"github.com/tickraft/tickraft/pkg/asset"
	"github.com/tickraft/tickraft/pkg/executor"
	"github.com/tickraft/tickraft/pkg/prism/alert"
	"github.com/tickraft/tickraft/pkg/task"
	"github.com/tickraft/tickraft/pkg/telemetry"
)

// TaskView is the expr-lang view of a task.Task used by the
// task matching scene. Field names are exposed to expressions via the
// expr struct tag using snake_case identifiers (e.g. task.priority).
type TaskView struct {
	// ID is the task identifier.
	ID int64 `expr:"id"`
	// TenantID is the tenant that owns the task.
	TenantID int64 `expr:"tenant_id"`
	// AssetID is the associated asset identifier.
	AssetID int64 `expr:"asset_id"`
	// ExecutorType identifies which executor runs the task.
	ExecutorType string `expr:"executor_type"`
	// Priority controls execution order; higher values fire first.
	Priority int `expr:"priority"`
	// Timeout is the maximum execution duration.
	Timeout time.Duration `expr:"timeout"`
	// Metadata holds optional key-value extension data.
	Metadata map[string]string `expr:"metadata"`
}

// AssetView is the expr-lang view of an asset.Asset shared by
// all matching scenes.
type AssetView struct {
	// ID is the asset identifier.
	ID int64 `expr:"id"`
	// TenantID is the tenant that owns the asset.
	TenantID int64 `expr:"tenant_id"`
	// Type categorizes the asset, stored as a plain string so rule
	// expressions can compare it directly to a string literal
	// (e.g. asset.type == "host").
	Type string `expr:"type"`
	// Name is the human-readable asset name.
	Name string `expr:"name"`
	// Status is the current asset status, stored as a plain string so
	// rule expressions can compare it directly to a string literal
	// (e.g. asset.status == "abnormal").
	Status string `expr:"status"`
}

// ResultView is the expr-lang view of an executor.Result used by the
// probe matching scene.
type ResultView struct {
	// Status is the execution result status, stored as a plain string
	// so rule expressions can compare it directly to a string literal.
	Status string `expr:"status"`
	// StatusCode is the protocol-specific status code.
	StatusCode int `expr:"status_code"`
	// Body is the response body or execution output.
	Body string `expr:"body"`
	// ErrorMsg describes any execution error.
	ErrorMsg string `expr:"error_msg"`
	// Duration is the total execution time.
	Duration time.Duration `expr:"duration"`
	// Metrics carries optional numerical metrics.
	Metrics map[string]float64 `expr:"metrics"`
}

// ReportView is the expr-lang view of a telemetry.Telemetry used
// by the probe matching scene.
type ReportView struct {
	// AssetID is the reported asset identifier.
	AssetID int64 `expr:"asset_id"`
	// TenantID is the tenant that owns the reported asset.
	TenantID int64 `expr:"tenant_id"`
	// AssetType categorizes the reported asset, stored as a plain string
	// so rule expressions can compare it to a string literal.
	AssetType string `expr:"asset_type"`
	// SourceType identifies the data source.
	SourceType string `expr:"source_type"`
	// RemoteAddr is the source address of the report.
	RemoteAddr string `expr:"remote_addr"`
	// CollectedAt is when the data was collected.
	CollectedAt time.Time `expr:"collected_at"`
	// Metrics holds extracted numerical metrics.
	Metrics map[string]float64 `expr:"metrics"`
	// LogContent holds log content.
	LogContent string `expr:"log_content"`
	// LogLevel holds the severity level of the log content.
	LogLevel string `expr:"log_level"`
	// Status is the pre-judged status set by Listeners, stored as a
	// plain string so rule expressions can compare it to a string literal.
	Status string `expr:"status"`
}

// AlertView is the expr-lang view of an alert.Event used by
// the metric matching scene. The view is projected from the alert's
// primary violation so a single struct exposes the fields rule
// expressions care about.
type AlertView struct {
	// Type is the alert category, stored as a plain string so rule
	// expressions can compare it directly to a string literal
	// (e.g. `event.type == "metric"`). The canonical alert.Type named
	// type is converted at projection time in toAlertView; storing a
	// plain string here avoids expr-lang's "mismatched types" error
	// when a named type is compared to a string literal.
	Type string `expr:"type"`
	// AssetID is the asset that triggered the alert.
	AssetID int64 `expr:"asset_id"`
	// TenantID is the owning tenant.
	TenantID int64 `expr:"tenant_id"`
	// Timestamp is when the alert was generated.
	Timestamp time.Time `expr:"timestamp"`
	// Severity is the unified severity carried by the primary violation.
	Severity string `expr:"severity"`
	// Metrics holds the related metric values from the primary violation.
	Metrics map[string]float64 `expr:"metrics"`
	// Keyword is the matched log keyword.
	Keyword string `expr:"keyword"`
	// Content is the log line that matched.
	Content string `expr:"content"`
	// Source is the origin of the alert (IP for log, probe_id for heartbeat).
	Source string `expr:"source"`
}

// RemediationView is the expr-lang view of a self-healing rule
// evaluation context. It is projected from a remediation rule's
// EventContext and the triggering alert event.
type RemediationView struct {
	// Type is the alert category that triggered the remediation.
	Type string `expr:"type"`
	// AssetID is the asset targeted by the remediation.
	AssetID int64 `expr:"asset_id"`
	// MetricName is the metric that violated a threshold, projected
	// from the primary violation.
	MetricName string `expr:"metric_name"`
	// MetricValue is the observed metric value, projected from the
	// primary violation.
	MetricValue float64 `expr:"metric_value"`
	// Threshold is the configured threshold, projected from the
	// primary violation.
	Threshold float64 `expr:"threshold"`
	// Severity is the unified severity of the triggering alert.
	Severity string `expr:"severity"`
	// Keyword is the matched log keyword, projected from the primary
	// violation.
	Keyword string `expr:"keyword"`
	// Content is the log line that matched, projected from the primary
	// violation.
	Content string `expr:"content"`
	// SourceIP is the origin address of the triggering alert.
	SourceIP string `expr:"source_ip"`
	// PrevStatus is the asset status before the remediation, used to
	// guard against re-running workflows when the state has already
	// transitioned.
	PrevStatus string `expr:"prev_status"`
	// CurrStatus is the asset status after the alert that triggered
	// the remediation.
	CurrStatus string `expr:"curr_status"`
}

// TaskMatchEnv is the expr-lang evaluation environment for the task
// matching scene.
type TaskMatchEnv struct {
	// Task is the task under evaluation.
	Task TaskView `expr:"task"`
	// Event is an alias of Task exposed under the intuitive top-level
	// name "event". It mirrors Task field-for-field so rule authors can
	// write `event.executor_type == "ssh"` without learning the
	// scene-specific "task" handle. It is purely additive: existing
	// expressions referencing "task" are unaffected.
	Event TaskView `expr:"event"`
	// Asset is the asset associated with the task.
	Asset AssetView `expr:"asset"`
	// Tags are key-value tags derived from the asset metadata.
	Tags map[string]string `expr:"tags"`
}

// ProbeMatchEnv is the expr-lang evaluation environment for the probe
// matching scene.
type ProbeMatchEnv struct {
	// Result is the executor result (zero-value for passive reports).
	Result ResultView `expr:"result"`
	// Report is the collector report context.
	Report ReportView `expr:"report"`
	// Event is an alias of Report exposed under the intuitive top-level
	// name "event". It mirrors Report field-for-field so rule authors
	// can write `event.log_level == "error"` without learning the
	// scene-specific "report" handle. It is purely additive: existing
	// expressions referencing "report" are unaffected.
	Event ReportView `expr:"event"`
	// Asset is the asset associated with the report.
	Asset AssetView `expr:"asset"`
}

// MetricMatchEnv is the expr-lang evaluation environment for the
// metric matching scene.
type MetricMatchEnv struct {
	// Alert is the originating alert event projected from alert.Event.
	Alert AlertView `expr:"alert"`
	// Event is an alias of Alert exposed under the intuitive top-level
	// name "event". It mirrors Alert field-for-field so rule authors
	// can write `event.type == "log"` or `event.metrics["cpu"] > 80`
	// without learning the scene-specific "alert" handle. It is purely
	// additive: existing expressions referencing "alert" are unaffected.
	Event AlertView `expr:"event"`
	// Asset is the asset associated with the alert.
	Asset AssetView `expr:"asset"`
}

// RemediationMatchEnv is the expr-lang evaluation environment for the
// remediation matching scene.
type RemediationMatchEnv struct {
	// Remediation is the projected remediation context.
	Remediation RemediationView `expr:"remediation"`
	// Event is an alias of Remediation exposed under the intuitive
	// top-level name "event". It mirrors Remediation field-for-field so
	// rule authors can write `event.type == "metric"` or
	// `event.metric_value > 80` without learning the scene-specific
	// "remediation" handle. It is purely additive: existing expressions
	// referencing "remediation" are unaffected.
	Event RemediationView `expr:"event"`
	// Asset is the asset associated with the remediation.
	Asset AssetView `expr:"asset"`
}

// toTaskView projects a task.Task into the expr-lang TaskView.
func toTaskView(tk task.Task) TaskView {
	return TaskView{
		ID:           tk.ID,
		TenantID:     tk.TenantID,
		AssetID:      tk.AssetID,
		ExecutorType: tk.ExecutorName,
		Priority:     tk.Priority,
		Timeout:      tk.Timeout,
		Metadata:     tk.Metadata,
	}
}

// toAssetView projects an asset.Asset into the expr-lang
// AssetView.
func toAssetView(res asset.Asset) AssetView {
	return AssetView{
		ID:       res.ID,
		TenantID: res.TenantID,
		Type:     string(res.AssetType),
		Name:     res.Name,
		Status:   string(res.Status),
	}
}

// toResultView projects an executor.Result into the expr-lang
// ResultView.
func toResultView(result executor.Result) ResultView {
	return ResultView{
		Status:     string(result.Status),
		StatusCode: result.StatusCode,
		Body:       result.Body,
		ErrorMsg:   result.ErrorMsg,
		Duration:   result.Duration,
		Metrics:    result.Metrics,
	}
}

// toReportView projects a telemetry.Telemetry into the expr-lang
// ReportView.
func toReportView(report telemetry.Telemetry) ReportView {
	return ReportView{
		AssetID:     report.AssetID,
		TenantID:    report.TenantID,
		AssetType:   string(report.AssetType),
		SourceType:  report.SourceType,
		RemoteAddr:  report.RemoteAddr,
		CollectedAt: report.CollectedAt,
		Metrics:     report.Metrics,
		LogContent:  report.LogContent,
		LogLevel:    report.LogLevel,
		Status:      string(report.Status),
	}
}

// toAlertView projects an alert.Event into the expr-lang AlertView.
// The metric/keyword/content fields are projected from the event's
// primary violation sub-contexts so a single struct exposes the fields
// rule expressions care about. When the event carries no violations the
// projection yields a zero-value AlertView for those fields.
func toAlertView(evt alert.Event) AlertView {
	view := AlertView{
		Type:      string(evt.Type),
		AssetID:   evt.AssetID,
		TenantID:  evt.TenantID,
		Timestamp: evt.Timestamp,
	}
	if primary, ok := evt.PrimaryViolation(); ok {
		view.Severity = primary.Severity
		view.Source = primary.Source
		if primary.Metric != nil {
			view.Metrics = primary.Metric.Metrics
		}
		if primary.Log != nil {
			view.Keyword = primary.Log.Keyword
			view.Content = primary.Log.Content
		}
	}
	return view
}

// toRemediationView projects the supplied fields into the expr-lang
// RemediationView. The metric fields (MetricName, MetricValue,
// Threshold) come from the triggering alert's primary violation metric
// context; the keyword and content fields likewise come from the primary
// violation log context. The status fields describe the asset state
// transition that the remediation may observe or guard against.
func toRemediationView(
	evt alert.Event,
	prevStatus, currStatus string,
) RemediationView {
	view := RemediationView{
		Type:       string(evt.Type),
		AssetID:    evt.AssetID,
		PrevStatus: prevStatus,
		CurrStatus: currStatus,
	}
	if primary, ok := evt.PrimaryViolation(); ok {
		view.Severity = primary.Severity
		view.SourceIP = primary.Source
		if primary.Metric != nil {
			view.MetricName = primary.Metric.Name
			view.MetricValue = primary.Metric.Value
			view.Threshold = primary.Metric.Threshold
		}
		if primary.Log != nil {
			view.Keyword = primary.Log.Keyword
			view.Content = primary.Log.Content
		}
	}
	return view
}
