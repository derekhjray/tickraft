// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

// Package cache provides concrete caching implementations with TTL support.
//
// The package ships two implementations:
//   - LRUCache: an in-memory LRU cache with per-entry TTL expiration.
//   - BboltCache: a persistent cache backed by an embedded bbolt database.
//
// # Choosing an implementation
//
// Use LRUCache when:
//   - Cache data is purely a performance optimization and can be lost on restart.
//   - Low latency and high throughput are required.
//   - The working set fits comfortably in memory.
//
// Use BboltCache when:
//   - Cache data must survive process restarts.
//   - The cache is used to persist derived or expensive-to-recompute values.
//   - Memory pressure makes an on-disk store preferable.
//
// Both implementations are safe for concurrent use and honor the same TTL
// semantics (per-entry expiration, checked on read).
//
// # Usage
//
// Create an LRU cache and use it:
//
//	c := cache.NewLRU(1024, 5*time.Minute)
//	c.Set("key", []byte("value"))
//	val, ok := c.Get("key")
//
// Create a persistent bbolt cache and use it:
//
//	c, err := cache.NewBbolt("/var/lib/tickraft/cache.db", 5*time.Minute)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//	c.Set("key", []byte("value"))
//	val, ok := c.Get("key")
package cache
