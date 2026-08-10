// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

// newTestBboltCache creates a BboltCache backed by a temporary file with the
// given default TTL. The temporary path is provided by t.TempDir() and is
// cleaned up automatically when the test finishes.
func newTestBboltCache(t *testing.T, defaultTTL time.Duration) *BboltCache {
	t.Helper()
	path := t.TempDir() + "/cache.db"
	c, err := NewBbolt(path, defaultTTL)
	if err != nil {
		t.Fatalf("failed to create bbolt cache: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})
	return c
}

func TestBboltCache_BasicOperations(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	// Set + Get.
	c.Set(context.Background(), "user:1", []byte("alice"))
	val, ok := c.Get(context.Background(), "user:1")
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if string(val) != "alice" {
		t.Fatalf("expected alice, got %s", val)
	}

	// Has on existing key.
	if !c.Has(context.Background(), "user:1") {
		t.Fatal("expected Has to return true for existing key")
	}

	// Has on missing key.
	if c.Has(context.Background(), "missing") {
		t.Fatal("expected Has to return false for missing key")
	}

	// Delete.
	c.Delete(context.Background(), "user:1")
	if _, ok := c.Get(context.Background(), "user:1"); ok {
		t.Fatal("expected cache miss after delete")
	}
	if c.Has(context.Background(), "user:1") {
		t.Fatal("expected Has to return false after delete")
	}

	// Delete on missing key is a no-op.
	c.Delete(context.Background(), "nonexistent")
}

func TestBboltCache_TTLExpiry(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	c.SetWithTTL(context.Background(), "session:abc", []byte("token123"), 50*time.Millisecond)

	// Should be present before expiration.
	val, ok := c.Get(context.Background(), "session:abc")
	if !ok {
		t.Fatal("expected cache hit before TTL expiration")
	}
	if string(val) != "token123" {
		t.Fatalf("expected token123, got %s", val)
	}

	// Wait for expiration.
	time.Sleep(100 * time.Millisecond)

	// Should be missing after expiration.
	if _, ok := c.Get(context.Background(), "session:abc"); ok {
		t.Fatal("expected cache miss after TTL expiration")
	}

	// Has should also return false after expiration.
	if c.Has(context.Background(), "session:abc") {
		t.Fatal("expected Has to return false after TTL expiration")
	}
}

func TestBboltCache_DefaultTTL(t *testing.T) {
	// Use a short default TTL to verify Set applies it.
	c := newTestBboltCache(t, 50*time.Millisecond)

	c.Set(context.Background(), "key", []byte("value"))

	// Should be present immediately.
	if _, ok := c.Get(context.Background(), "key"); !ok {
		t.Fatal("expected cache hit before default TTL expiration")
	}

	// Wait for the default TTL to elapse.
	time.Sleep(100 * time.Millisecond)

	if _, ok := c.Get(context.Background(), "key"); ok {
		t.Fatal("expected cache miss after default TTL expiration")
	}
}

func TestBboltCache_GetWithTTL(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	// Missing key.
	_, _, ok := c.GetWithTTL(context.Background(), "missing")
	if ok {
		t.Fatal("expected GetWithTTL to return false for missing key")
	}

	// Existing key with TTL.
	c.SetWithTTL(context.Background(), "session:abc", []byte("token"), 5*time.Second)
	val, ttl, ok := c.GetWithTTL(context.Background(), "session:abc")
	if !ok {
		t.Fatal("expected GetWithTTL to return true for existing key")
	}
	if string(val) != "token" {
		t.Fatalf("expected token, got %s", val)
	}
	if ttl <= 0 || ttl > 5*time.Second {
		t.Fatalf("expected TTL between 0 and 5s, got %v", ttl)
	}

	// Expired key.
	c.SetWithTTL(context.Background(), "session:expired", []byte("gone"), 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	_, _, ok = c.GetWithTTL(context.Background(), "session:expired")
	if ok {
		t.Fatal("expected GetWithTTL to return false for expired key")
	}
}

func TestBboltCache_DeleteByPrefix(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	c.Set(context.Background(), "user:1", []byte("alice"))
	c.Set(context.Background(), "user:2", []byte("bob"))
	c.Set(context.Background(), "tenant:1", []byte("acme"))
	c.Set(context.Background(), "tenant:2", []byte("beta"))

	c.DeleteByPrefix(context.Background(), "user:")

	// user keys should be gone.
	if _, ok := c.Get(context.Background(), "user:1"); ok {
		t.Fatal("expected user:1 to be deleted")
	}
	if _, ok := c.Get(context.Background(), "user:2"); ok {
		t.Fatal("expected user:2 to be deleted")
	}

	// tenant keys should remain.
	if val, ok := c.Get(context.Background(), "tenant:1"); !ok || string(val) != "acme" {
		t.Fatalf("expected tenant:1=acme, got %s, ok=%v", val, ok)
	}
	if val, ok := c.Get(context.Background(), "tenant:2"); !ok || string(val) != "beta" {
		t.Fatalf("expected tenant:2=beta, got %s, ok=%v", val, ok)
	}
}

func TestBboltCache_Clear(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	c.Set(context.Background(), "user:1", []byte("alice"))
	c.Set(context.Background(), "user:2", []byte("bob"))
	c.Set(context.Background(), "tenant:1", []byte("acme"))

	if s := c.Size(context.Background()); s != 3 {
		t.Fatalf("expected size 3 before clear, got %d", s)
	}

	c.Clear(context.Background())

	for _, key := range []string{"user:1", "user:2", "tenant:1"} {
		if _, ok := c.Get(context.Background(), key); ok {
			t.Fatalf("expected cache miss for %s after clear", key)
		}
	}

	if s := c.Size(context.Background()); s != 0 {
		t.Fatalf("expected size 0 after clear, got %d", s)
	}
}

func TestBboltCache_Size(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	if s := c.Size(context.Background()); s != 0 {
		t.Fatalf("expected size 0, got %d", s)
	}

	c.Set(context.Background(), "user:1", []byte("alice"))
	c.Set(context.Background(), "user:2", []byte("bob"))

	if s := c.Size(context.Background()); s != 2 {
		t.Fatalf("expected size 2, got %d", s)
	}

	c.Delete(context.Background(), "user:1")

	if s := c.Size(context.Background()); s != 1 {
		t.Fatalf("expected size 1, got %d", s)
	}

	c.Clear(context.Background())

	if s := c.Size(context.Background()); s != 0 {
		t.Fatalf("expected size 0 after clear, got %d", s)
	}
}

func TestBboltCache_Persistence(t *testing.T) {
	path := t.TempDir() + "/cache.db"

	// Create the cache, write data, and close it.
	c, err := NewBbolt(path, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to create bbolt cache: %v", err)
	}
	c.Set(context.Background(), "user:1", []byte("alice"))
	c.Set(context.Background(), "user:2", []byte("bob"))
	if err = c.Close(context.Background()); err != nil {
		t.Fatalf("failed to close cache: %v", err)
	}

	// Reopen the same file and verify data survives.
	c2, err := NewBbolt(path, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to reopen bbolt cache: %v", err)
	}
	defer c2.Close(context.Background())

	if val, ok := c2.Get(context.Background(), "user:1"); !ok || string(val) != "alice" {
		t.Fatalf("expected user:1=alice after reopen, got %s, ok=%v", val, ok)
	}
	if val, ok := c2.Get(context.Background(), "user:2"); !ok || string(val) != "bob" {
		t.Fatalf("expected user:2=bob after reopen, got %s, ok=%v", val, ok)
	}
}

func TestBboltCache_NilValue(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	// Storing an empty byte slice should be retrievable.
	c.Set(context.Background(), "empty", []byte{})
	val, ok := c.Get(context.Background(), "empty")
	if !ok {
		t.Fatal("expected cache hit for empty value")
	}
	if len(val) != 0 {
		t.Fatalf("expected empty value, got %s", val)
	}

	// Storing a nil byte slice should also be retrievable as empty.
	c.Set(context.Background(), "nil", nil)
	val, ok = c.Get(context.Background(), "nil")
	if !ok {
		t.Fatal("expected cache hit for nil value")
	}
	if len(val) != 0 {
		t.Fatalf("expected empty value, got %s", val)
	}
}

func TestBboltCache_Concurrent(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "key:" + string(rune('a'+n%26))
			c.Set(context.Background(), key, []byte{byte(n)})
			_, _ = c.Get(context.Background(), key)
			c.Has(context.Background(), key)
			_, _, _ = c.GetWithTTL(context.Background(), key)
			c.Delete(context.Background(), key)
		}(i)
	}
	wg.Wait()
}

func TestBboltCache_SetUpdatesExistingKey(t *testing.T) {
	c := newTestBboltCache(t, 5*time.Minute)

	c.Set(context.Background(), "user:1", []byte("alice"))
	c.Set(context.Background(), "user:1", []byte("bob"))

	val, ok := c.Get(context.Background(), "user:1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(val) != "bob" {
		t.Fatalf("expected bob, got %s", val)
	}
}

func TestBboltCache_DefaultConstructor(t *testing.T) {
	// Zero defaultTTL should fall back to 5 minutes.
	path := t.TempDir() + "/cache.db"
	c, err := NewBbolt(path, 0)
	if err != nil {
		t.Fatalf("failed to create bbolt cache: %v", err)
	}
	defer c.Close(context.Background())
	if c.defaultTTL != 5*time.Minute {
		t.Fatalf("expected default TTL=5m, got %v", c.defaultTTL)
	}
}

func TestBboltCache_EmptyPathReturnsError(t *testing.T) {
	if _, err := NewBbolt("", 5*time.Minute); err == nil {
		t.Fatal("expected error when path is empty")
	}
}

func TestBboltCache_CloseTwiceIsSafe(t *testing.T) {
	path := t.TempDir() + "/cache.db"
	c, err := NewBbolt(path, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to create bbolt cache: %v", err)
	}
	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}
	// Second Close should be a no-op and return nil.
	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("second Close should be a no-op, got %v", err)
	}
}
