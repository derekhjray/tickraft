// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package httpx

import (
	"net/http"
	"sync"
	"time"
)

// defaultClient is the lazily-initialized shared pooled client used by
// [DefaultClient]. It is shared across all callers that do not need
// per-call tuning, so that idle connections are reused process-wide.
var (
	defaultClient     *http.Client
	defaultClientOnce sync.Once
)

// DefaultClient returns a process-wide [*http.Client] backed by a
// pooled transport with package defaults. The same instance is returned
// on every call so that idle connections are reused across callers,
// which is the primary benefit of connection pooling.
//
// Callers must not mutate the returned client's Transport or Timeout.
// Use [NewPoolClient] when a dedicated client with custom tuning is
// required.
func DefaultClient() *http.Client {
	defaultClientOnce.Do(func() {
		defaultClient = NewPoolClient(Config{Timeout: DefaultTimeout})
	})
	return defaultClient
}

// Preset groups a [Config] with a documented intent so that callers can
// select a tuned client without spelling out every field. Presets are
// plain values; callers may copy and modify them before passing to
// [NewPoolClient].
type Preset struct {
	// Name is a short human-readable label for logging and diagnostics.
	Name string
	// Config is the pooled-client configuration.
	Config Config
}

// Presets for common outbound-traffic profiles.
var (
	// PresetDefault is the general-purpose preset: 30s end-to-end
	// timeout, 100 idle connections, 10 per host, 90s idle lifetime.
	PresetDefault = Preset{
		Name: "default",
		Config: Config{
			Timeout:             DefaultTimeout,
			MaxIdleConns:        DefaultMaxIdleConns,
			MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
			IdleConnTimeout:     DefaultIdleConnTimeout,
		},
	}

	// PresetFast is for low-latency control-plane calls (e.g. metadata
	// services, internal API-to-API). A short timeout surfaces
	// failures quickly without exhausting the connection pool.
	PresetFast = Preset{
		Name: "fast",
		Config: Config{
			Timeout:             5 * time.Second,
			MaxIdleConns:        DefaultMaxIdleConns,
			MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
			IdleConnTimeout:     DefaultIdleConnTimeout,
		},
	}

	// PresetLong is for outbound calls that may legitimately take longer
	// (e.g. webhook delivery to slow endpoints, large report uploads).
	PresetLong = Preset{
		Name: "long",
		Config: Config{
			Timeout:             60 * time.Second,
			MaxIdleConns:        DefaultMaxIdleConns,
			MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
			IdleConnTimeout:     DefaultIdleConnTimeout,
		},
	}
)

// NewFromPreset builds a pooled client from a [Preset]. The preset's
// Config is copied, so callers may safely mutate the returned client.
func NewFromPreset(p Preset) *http.Client {
	return NewPoolClient(p.Config)
}
