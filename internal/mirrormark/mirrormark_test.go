package mirrormark

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

// Cohort-canonical KAT-1 mark literal. Byte-identical to every cohort
// Go port (foundation/pkg/mirrormark + nexus + folio + pulse + oracle
// + baseline + iris + canopy + green-anchor + aegis). I52 marathon
// 2026-05-28 — limitless-environmental joins as the UK EA + EU IED
// substrate.
const kat1Mark = "lore@v1:AAAAAAAAAAAjmn0NPxu-Opiu3gHirYGMLbYLcXfALi8BUDWytbfbyg"

// Cohort-canonical KAT-6 mark literal. 0x33 corpus + "hello world" + "iik_hello".
const kat6Mark = "lore@v1:MzMzMzMzMzNDXUcWs_KJVkPQfl3-ykizfhchYGxWCw-IoxKxgijBOw"

// Cohort-canonical KAT-7 mark literal. Identity corpus + pulse JSON.
const kat7Mark = "lore@v1:AAECAwQFBgdXSiwQoZ5vwuA9nIqeZ_2v8tfAsQWV2ow_OiE34Pud_w"

// KAT-1 HMAC-SHA256 digest hex — regulator OpenSSL-reproducible.
const kat1DigestHex = "239a7d0d3f1bbe3a98aede01e2ad818c2db60b7177c02e2f015035b2b5b7dbca"

// TestVerify_KAT1Mark — cohort substrate-parity firewall.
func TestVerify_KAT1Mark(t *testing.T) {
	var zeroCorpus [sha256.Size]byte
	if err := Verify(kat1Mark, zeroCorpus, []byte{}, []byte{}); err != nil {
		t.Fatalf("KAT-1 cohort literal failed Verify: %v\n\nThe limitless-environmental mirrormark algorithm has drifted from the cohort.", err)
	}
}

// TestSign_ProducesKAT1Mark — Sign reproduces published literal.
func TestSign_ProducesKAT1Mark(t *testing.T) {
	var zeroCorpus [sha256.Size]byte
	got := Sign(zeroCorpus, []byte{}, []byte{})
	if got != kat1Mark {
		t.Fatalf("Sign for KAT-1 input drift:\n  got:  %q\n  want: %q", got, kat1Mark)
	}
}

// TestVerify_KAT6Mark — 0x33 corpus + "hello world" + "iik_hello".
func TestVerify_KAT6Mark(t *testing.T) {
	var corpus [sha256.Size]byte
	for i := range corpus {
		corpus[i] = 0x33
	}
	if err := Verify(kat6Mark, corpus, []byte("hello world"), []byte("iik_hello")); err != nil {
		t.Fatalf("KAT-6 cohort literal failed Verify: %v", err)
	}
}

// TestSign_ProducesKAT6Mark — KAT-6 inputs reproduce literal.
func TestSign_ProducesKAT6Mark(t *testing.T) {
	var corpus [sha256.Size]byte
	for i := range corpus {
		corpus[i] = 0x33
	}
	got := Sign(corpus, []byte("hello world"), []byte("iik_hello"))
	if got != kat6Mark {
		t.Fatalf("Sign for KAT-6 input drift:\n  got:  %q\n  want: %q", got, kat6Mark)
	}
}

// TestVerify_KAT7Mark — identity corpus + pulse JSON.
func TestVerify_KAT7Mark(t *testing.T) {
	var corpus [sha256.Size]byte
	for i := range corpus {
		corpus[i] = byte(i)
	}
	payload := []byte(`{"probeId":"https-google","verdict":"red","ms":5000,"error":"connection-timeout"}`)
	key := []byte("iik_pulse_kat_probe_failure")
	if err := Verify(kat7Mark, corpus, payload, key); err != nil {
		t.Fatalf("KAT-7 cohort literal failed Verify: %v", err)
	}
}

// TestKAT1Digest_EmbeddedInKAT1Mark — connects mark literal -> OpenSSL.
func TestKAT1Digest_EmbeddedInKAT1Mark(t *testing.T) {
	encoded := strings.TrimPrefix(kat1Mark, MarkPrefix)
	body, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("KAT-1 mark body not valid base64.RawURLEncoding: %v", err)
	}
	if len(body) != MarkBodyLen {
		t.Fatalf("KAT-1 body length: got %d want %d", len(body), MarkBodyLen)
	}
	gotDigestHex := hex.EncodeToString(body[MarkCorpusPrefixLen:])
	if gotDigestHex != kat1DigestHex {
		t.Fatalf("KAT-1 embedded-digest drift:\n  got:      %s\n  expected: %s", gotDigestHex, kat1DigestHex)
	}
}

// TestSign_RoundtripVerify — happy-path: signed mark round-trips.
func TestSign_RoundtripVerify(t *testing.T) {
	var corpus [sha256.Size]byte
	for i := range corpus {
		corpus[i] = byte(i * 7)
	}
	key := []byte("iik_env_test_key")
	payload := []byte(`{"permit":"EPR/AB1234XY","outcome":"EA_PERMIT_FRESH"}`)
	mark := Sign(corpus, payload, key)
	if !strings.HasPrefix(mark, MarkPrefix) {
		t.Fatalf("missing prefix: %q", mark)
	}
	if err := Verify(mark, corpus, payload, key); err != nil {
		t.Fatalf("Verify rejected fresh mark: %v", err)
	}
}

// TestVerify_RejectsMissingPrefix — non-Mirror-Mark string.
func TestVerify_RejectsMissingPrefix(t *testing.T) {
	var corpus [sha256.Size]byte
	err := Verify("not-a-mark", corpus, []byte{}, []byte("k"))
	if err != ErrUnknownMarkVersion {
		t.Fatalf("missing-prefix: got %v, want ErrUnknownMarkVersion", err)
	}
}

// TestVerify_RejectsMalformedBase64 — invalid base64 body.
func TestVerify_RejectsMalformedBase64(t *testing.T) {
	var corpus [sha256.Size]byte
	err := Verify("lore@v1:!!!not-base64!!!", corpus, []byte{}, []byte("k"))
	if err != ErrMalformedMark {
		t.Fatalf("malformed-base64: got %v, want ErrMalformedMark", err)
	}
}

// TestVerify_RejectsWrongCorpus — corpus A signed, corpus B passed.
func TestVerify_RejectsWrongCorpus(t *testing.T) {
	var corpusA, corpusB [sha256.Size]byte
	for i := range corpusA {
		corpusA[i] = 0x11
		corpusB[i] = 0x22
	}
	key := []byte("k")
	payload := []byte("p")
	markA := Sign(corpusA, payload, key)
	err := Verify(markA, corpusB, payload, key)
	if err != ErrCorpusMismatch {
		t.Fatalf("wrong-corpus: got %v, want ErrCorpusMismatch", err)
	}
}

// TestVerify_RejectsTamperedPayload — payload mutation.
func TestVerify_RejectsTamperedPayload(t *testing.T) {
	var corpus [sha256.Size]byte
	for i := range corpus {
		corpus[i] = 0x44
	}
	key := []byte("k")
	markA := Sign(corpus, []byte("original"), key)
	err := Verify(markA, corpus, []byte("tampered"), key)
	if err != ErrSignatureMismatch {
		t.Fatalf("tampered-payload: got %v, want ErrSignatureMismatch", err)
	}
}

// TestMarkLength_FixedAt62 — every Mirror-Mark v1 is 62 chars.
func TestMarkLength_FixedAt62(t *testing.T) {
	var corpus [sha256.Size]byte
	for i := range corpus {
		corpus[i] = byte(i * 3)
	}
	mark := Sign(corpus, []byte("anything"), []byte("k"))
	if len(mark) != 62 {
		t.Fatalf("Mark length: got %d, want 62", len(mark))
	}
}
