// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/tickraft/tickraft/internal/api/service/prism"
	"github.com/tickraft/tickraft/pkg/prism/alert"
	"github.com/tickraft/tickraft/pkg/prism/channel"
	"github.com/tickraft/tickraft/pkg/prism/governance"
	"github.com/tickraft/tickraft/pkg/prism/remediation"
	"github.com/tickraft/tickraft/pkg/prism/rule"
)

// startPrismEngine builds and starts the prism engine, subscribing it to
// telemetry alert events on the shared event bus. It creates the rule
// store (pkg/prism/rule) and alert record store, wires an OnAlert
// callback that persists alert records, registers the rule engine via
// rule.Register so metric alerts are filtered by expr-lang rules, and
// registers notification channels parsed from channelConfig. The
// engines, rule store, and record store are stored on the runtime so
// that startAPIServer can wire the prismAlertService into the API routes.
func startPrismEngine(ctx context.Context, rt *runtime,
	notificationPoolSize int, channelConfig string,
) (stopFunc, error) {
	bus := rt.eventBus()

	// Create the rule store and alert record store, ensuring their tables
	// exist. The rule store wraps a Compiler so Create/Update validate
	// expressions before they reach the database.
	ruleStore := rule.NewStore(rt.dbc, rule.NewCompiler())
	if err := ruleStore.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate rule table: %w", err)
	}
	recordStore := alert.NewRecordStore(rt.dbc)
	if err := alert.Migrate(ctx, rt.dbc); err != nil {
		return nil, fmt.Errorf("migrate alert tables: %w", err)
	}

	// Create the notification channel store and ensure its table exists.
	// The store persists channel definitions managed through the CRUD API
	// at /api/v1/prism/channels and is consumed by startAPIServer when
	// wiring the channel service into the API routes.
	channelStore := channel.NewStore(rt.dbc)
	if err := channelStore.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate channel table: %w", err)
	}

	// Create the remediation rule store and ensure its table exists.
	// The store persists self-healing rule definitions managed through the
	// CRUD API at /api/v1/prism/remediation/rules and is consumed by
	// startAPIServer when wiring the remediation rule service into the API
	// routes.
	remediationRuleStore := remediation.NewStore(rt.dbc)
	if err := remediationRuleStore.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate remediation rule table: %w", err)
	}

	// OnAlert callback persists alert records to the record store. Errors
	// are logged but do not suppress channel notification.
	onAlert := func(ctx context.Context, evt alert.Event) {
		if err := prism.RecordAlert(ctx, recordStore, evt); err != nil {
			rt.logger.Warn("persist alert record",
				zap.String("type", string(evt.Type)),
				zap.Int64("asset_id", evt.AssetID),
				zap.Error(err),
			)
		}
	}

	engine, err := alert.New(
		alert.WithEventBus(bus),
		alert.WithLogger(rt.logger),
		alert.WithNotificationPoolSize(notificationPoolSize),
		// baseline governance: suppress exact-duplicate alerts
		// within a short window. The callers may inject the full
		// governance chain (silence -> aggregator -> suppressor -> storm) in
		// place of this single guard at startup.
		alert.WithGuards(governance.NewDedup(60*time.Second, rt.logger)),
		alert.WithOnAlert(onAlert),
	)
	if err != nil {
		return nil, fmt.Errorf("create prism engine: %w", err)
	}

	if channelConfig != "" {
		channels, err := LoadChannels(channelConfig)
		if err != nil {
			return nil, fmt.Errorf("load alert channels from %q: %w", channelConfig, err)
		}
		for _, ch := range channels {
			engine.AddChannel(ch)
		}
	}

	// Register the rule engine: compiles static rules, pulls dynamic
	// rules from the store, and injects the MetricMatcher into the prism
	// engine so metric alerts are filtered before dispatch. The returned
	// *rule.Engine is stored on the runtime so the alert service can
	// Reload after rule CRUD operations.
	ruleEng, err := rule.Register(ctx, engine, rule.Config{
		Logger:     rt.logger,
		Store:      ruleStore,
		AssetStore: rt.assetStore,
	})
	if err != nil {
		return nil, fmt.Errorf("register rule engine: %w", err)
	}

	if err := engine.Start(ctx); err != nil {
		return nil, fmt.Errorf("start prism: %w", err)
	}

	// Store on the runtime so startAPIServer can build the alert service.
	rt.alertEngine = engine
	rt.ruleEngine = ruleEng
	rt.ruleStore = ruleStore
	rt.alertRecordStore = recordStore
	rt.channelStore = channelStore
	rt.remediationRuleStore = remediationRuleStore

	rt.logger.Info("prism engine started",
		zap.Int("channels", len(engine.Channels())),
		zap.Int("rules", len(engine.Rules())),
	)

	return func(ctx context.Context) error {
		if ruleEng != nil {
			_ = ruleEng.Stop(ctx) // best-effort: cancel reload loop before stopping engine
		}
		return engine.Stop(ctx)
	}, nil
}
