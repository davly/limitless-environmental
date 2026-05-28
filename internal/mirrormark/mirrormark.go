// Package mirrormark implements the cohort L43 Mirror-Mark v1 receipt
// algorithm — byte-identical to foundation/pkg/mirrormark and to every
// cohort Go port (R151 KAT-AS-COHORT-INVARIANT-CROSS-SUBSTRATE-PIN).
//
// limitless-environmental is the UK Environment Agency + Environment
// Act 2021 + EU IED compliance flagship. The Mirror-Mark surface lets
// permit-decision payloads + BNG (biodiversity net gain) calculations
// be tamper-stamped on the way out of the runtime — a regulator (EA /
// SEPA / NRW / NIEA / EU Commission) with the corpus SHA + key can
// cold-verify a permit decision without trusting the upstream binary.
//
// Mark format (byte-identical to foundation/pkg/mirrormark):
//
//	"lore@v1:" + base64url( corpusSHA[:8] || hmacSHA256(0x01 || corpusSHA || payload, key) )
//
// Resulting in a fixed 62-character string.
package mirrormark

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// MarkVersion is the 1-byte tag prefixing the HMAC input.
const MarkVersion byte = 0x01

// MarkPrefix is the documented header-value prefix.
const MarkPrefix = "lore@v1:"

// MarkCorpusPrefixLen is the corpus-SHA prefix length (8 bytes).
const MarkCorpusPrefixLen = 8

// MarkBodyLen is the unencoded length of the mark body (40 bytes).
const MarkBodyLen = MarkCorpusPrefixLen + sha256.Size

var (
	// ErrUnknownMarkVersion — mark missing canonical prefix.
	ErrUnknownMarkVersion = errors.New("mirrormark: unknown mark version (missing 'lore@v1:' prefix)")
	// ErrMalformedMark — base64url decode failed or wrong body length.
	ErrMalformedMark = errors.New("mirrormark: malformed mark (base64url decode failed or wrong body length)")
	// ErrCorpusMismatch — corpus prefix in mark != supplied corpus SHA.
	ErrCorpusMismatch = errors.New("mirrormark: corpus prefix mismatch (mark signed by different corpus)")
	// ErrSignatureMismatch — HMAC digest mismatch (payload or key wrong).
	ErrSignatureMismatch = errors.New("mirrormark: HMAC signature mismatch (payload tampered or wrong key)")
)

// Sign returns the canonical Mirror-Mark string for the given payload.
// Byte-identical to foundation/pkg/mirrormark.Sign.
func Sign(corpusSHA [sha256.Size]byte, payload []byte, key []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte{MarkVersion})
	_, _ = mac.Write(corpusSHA[:])
	_, _ = mac.Write(payload)
	digest := mac.Sum(nil)

	body := make([]byte, 0, MarkBodyLen)
	body = append(body, corpusSHA[:MarkCorpusPrefixLen]...)
	body = append(body, digest...)

	return MarkPrefix + base64.RawURLEncoding.EncodeToString(body)
}

// Verify cold-checks a Mirror-Mark against the caller's (corpus,
// payload, key) triple. Returns nil on match; one of the typed
// sentinel errors on any failure. Both byte-comparisons use hmac.Equal
// (constant-time) — timing-safe.
func Verify(mark string, corpusSHA [sha256.Size]byte, payload []byte, key []byte) error {
	if len(mark) < len(MarkPrefix) || mark[:len(MarkPrefix)] != MarkPrefix {
		return ErrUnknownMarkVersion
	}
	body, err := base64.RawURLEncoding.DecodeString(mark[len(MarkPrefix):])
	if err != nil {
		return ErrMalformedMark
	}
	if len(body) != MarkBodyLen {
		return ErrMalformedMark
	}
	corpusPrefix := body[:MarkCorpusPrefixLen]
	digest := body[MarkCorpusPrefixLen:]
	if !hmac.Equal(corpusPrefix, corpusSHA[:MarkCorpusPrefixLen]) {
		return ErrCorpusMismatch
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte{MarkVersion})
	_, _ = mac.Write(corpusSHA[:])
	_, _ = mac.Write(payload)
	want := mac.Sum(nil)
	if !hmac.Equal(digest, want) {
		return ErrSignatureMismatch
	}
	return nil
}
