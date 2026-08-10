// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package region

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const (
	testSecret    = "super-secret-cookie-signing-key-32b"
	prevSecret    = "previous-secret-cookie-signing-key-32"
	testRegion    = "cn"
	testTimestamp = int64(1700000000)
)

func TestSignAndVerifyRoundTrip(t *testing.T) {
	signed := SignCookie(testRegion, testTimestamp, []byte(testSecret))

	region, ts, err := VerifyCookie(signed, []byte(testSecret))
	if err != nil {
		t.Fatalf("VerifyCookie() error = %v, want nil", err)
	}
	if region != testRegion {
		t.Errorf("region = %q, want %q", region, testRegion)
	}
	if ts != testTimestamp {
		t.Errorf("timestamp = %d, want %d", ts, testTimestamp)
	}
}

func TestVerifyCookieRejectsWrongSecret(t *testing.T) {
	signed := SignCookie(testRegion, testTimestamp, []byte(testSecret))

	_, _, err := VerifyCookie(signed, []byte("wrong-secret-key-different-length"))
	if !errors.Is(err, ErrCookieTampered) {
		t.Errorf("VerifyCookie(wrong secret) error = %v, want ErrCookieTampered", err)
	}
}

func TestVerifyCookieTamperedRegion(t *testing.T) {
	signed := SignCookie(testRegion, testTimestamp, []byte(testSecret))

	// Replace the region prefix while keeping the original signature.
	parts := strings.SplitN(signed, ":", 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	tampered := "global:" + parts[1] + ":" + parts[2]

	_, _, err := VerifyCookie(tampered, []byte(testSecret))
	if !errors.Is(err, ErrCookieTampered) {
		t.Errorf("VerifyCookie(tampered region) error = %v, want ErrCookieTampered", err)
	}
}

func TestVerifyCookieTamperedTimestamp(t *testing.T) {
	signed := SignCookie(testRegion, testTimestamp, []byte(testSecret))

	// Replace the timestamp while keeping the original signature.
	parts := strings.SplitN(signed, ":", 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	tampered := parts[0] + ":1700000099:" + parts[2]

	_, _, err := VerifyCookie(tampered, []byte(testSecret))
	if !errors.Is(err, ErrCookieTampered) {
		t.Errorf("VerifyCookie(tampered timestamp) error = %v, want ErrCookieTampered", err)
	}
}

func TestVerifyCookieTamperedSignature(t *testing.T) {
	signed := SignCookie(testRegion, testTimestamp, []byte(testSecret))

	// Flip a character in the HMAC signature.
	tampered := signed[:len(signed)-1]
	if tampered[len(tampered)-1] == '0' {
		tampered += "1"
	} else {
		tampered += "0"
	}

	_, _, err := VerifyCookie(tampered, []byte(testSecret))
	if !errors.Is(err, ErrCookieTampered) {
		t.Errorf("VerifyCookie(tampered signature) error = %v, want ErrCookieTampered", err)
	}
}

func TestVerifyCookieMalformed(t *testing.T) {
	cases := []string{
		"",                        // empty
		"cn",                      // missing parts
		"cn:1700000000",           // missing hmac
		"cn:notanumber:abc",       // bad timestamp
		"cn:1700000000:abc:extra", // too many parts
	}
	for _, tc := range cases {
		_, _, err := VerifyCookie(tc, []byte(testSecret))
		if err == nil {
			t.Errorf("VerifyCookie(%q) error = nil, want non-nil", tc)
		}
	}
}

func TestKeyRotator_VerifyCurrentKey(t *testing.T) {
	kr := NewKeyRotator([]byte(testSecret), []byte(prevSecret), DefaultRotationInterval, DefaultTransitionPeriod)

	signed := kr.Sign(testRegion, testTimestamp)

	region, ts, err := kr.Verify(signed)
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if region != testRegion {
		t.Errorf("region = %q, want %q", region, testRegion)
	}
	if ts != testTimestamp {
		t.Errorf("timestamp = %d, want %d", ts, testTimestamp)
	}
}

func TestKeyRotator_VerifyPreviousKeyDuringTransition(t *testing.T) {
	kr := NewKeyRotator([]byte(testSecret), []byte(prevSecret), DefaultRotationInterval, DefaultTransitionPeriod)

	// A cookie signed with the previous key should be accepted during the
	// transition period (lastRotation was set to now in the constructor).
	prevSigned := SignCookie(testRegion, testTimestamp, []byte(prevSecret))

	region, ts, err := kr.Verify(prevSigned)
	if err != nil {
		t.Fatalf("Verify(previous key) error = %v, want nil", err)
	}
	if region != testRegion {
		t.Errorf("region = %q, want %q", region, testRegion)
	}
	if ts != testTimestamp {
		t.Errorf("timestamp = %d, want %d", ts, testTimestamp)
	}
}

func TestKeyRotator_RejectsPreviousKeyAfterTransition(t *testing.T) {
	transition := time.Hour
	kr := NewKeyRotator([]byte(testSecret), []byte(prevSecret), DefaultRotationInterval, transition)

	// Move lastRotation far enough into the past that the transition period
	// has expired. Same-package access to the unexported field.
	kr.lastRotation = time.Now().Add(-2 * transition)

	prevSigned := SignCookie(testRegion, testTimestamp, []byte(prevSecret))

	_, _, err := kr.Verify(prevSigned)
	if err == nil {
		t.Error("Verify(previous key after transition) error = nil, want non-nil")
	}
}

func TestKeyRotator_CurrentKeyAlwaysAcceptedAfterTransition(t *testing.T) {
	transition := time.Hour
	kr := NewKeyRotator([]byte(testSecret), []byte(prevSecret), DefaultRotationInterval, transition)

	// Expire the transition period.
	kr.lastRotation = time.Now().Add(-2 * transition)

	// Current key should still be accepted regardless of transition period.
	currentSigned := kr.Sign(testRegion, testTimestamp)
	_, _, err := kr.Verify(currentSigned)
	if err != nil {
		t.Errorf("Verify(current key after transition) error = %v, want nil", err)
	}
}

func TestKeyRotator_RotateAcceptsOldKeyDuringNewTransition(t *testing.T) {
	kr := NewKeyRotator([]byte(testSecret), []byte(prevSecret), DefaultRotationInterval, DefaultTransitionPeriod)

	// Sign with the current key before rotation.
	oldSigned := kr.Sign(testRegion, testTimestamp)

	// Rotate to a brand new key.
	newSecret := "brand-new-secret-key-after-rotation-32"
	kr.Rotate([]byte(newSecret))

	// The old current key is now the previous key, and we are in a fresh
	// transition period, so the old-signed cookie should still verify.
	region, _, err := kr.Verify(oldSigned)
	if err != nil {
		t.Fatalf("Verify(old key after rotate) error = %v, want nil", err)
	}
	if region != testRegion {
		t.Errorf("region = %q, want %q", region, testRegion)
	}

	// The new current key should also verify.
	newSigned := kr.Sign("global", testTimestamp)
	region, _, err = kr.Verify(newSigned)
	if err != nil {
		t.Fatalf("Verify(new key after rotate) error = %v, want nil", err)
	}
	if region != "global" {
		t.Errorf("region = %q, want %q", region, "global")
	}
}

func TestKeyRotator_NilPreviousOnlyAcceptsCurrent(t *testing.T) {
	kr := NewKeyRotator([]byte(testSecret), nil, DefaultRotationInterval, DefaultTransitionPeriod)

	// Current key works.
	signed := kr.Sign(testRegion, testTimestamp)
	if _, _, err := kr.Verify(signed); err != nil {
		t.Errorf("Verify(current) error = %v, want nil", err)
	}

	// A cookie signed with an unrelated key must fail.
	otherSigned := SignCookie(testRegion, testTimestamp, []byte(prevSecret))
	_, _, err := kr.Verify(otherSigned)
	if err == nil {
		t.Error("Verify(unrelated key) error = nil, want non-nil")
	}
}
