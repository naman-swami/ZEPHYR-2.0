package capsule

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestCanonicalJSON(t *testing.T) {
	data1 := map[string]any{"z": 1, "a": "hello", "m": 42}
	data2 := map[string]any{"a": "hello", "m": 42, "z": 1}

	b1, err := CanonicalJSON(data1)
	if err != nil {
		t.Fatalf("failed canonicalizing data1: %v", err)
	}
	b2, err := CanonicalJSON(data2)
	if err != nil {
		t.Fatalf("failed canonicalizing data2: %v", err)
	}

	if string(b1) != string(b2) {
		t.Errorf("expected deterministic JSON match, got: %s vs %s", string(b1), string(b2))
	}
}

func TestDSSE_SignAndVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed generating ed25519 key: %v", err)
	}

	signer := NewSigner(priv)
	ring := NewKeyRing()
	ring.Add(pub)

	st := &Statement{
		Type:          StatementType,
		PredicateType: SLSAProvenanceV1,
		Subject: []Subject{
			{
				Name: "bin/zephyr",
				Digest: map[string]string{
					"sha256": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				},
			},
		},
		Predicate: ProvenancePredicate{
			BuildDefinition: BuildDefinition{
				BuildType: "https://zephyr.dev/slsa/v1",
				ExternalParameters: map[string]string{
					"task": "build",
				},
			},
			RunDetails: RunDetails{
				Builder: Builder{
					ID: "https://github.com/naman-swami/zephyr",
				},
				InvocationID: NewInvocationID(),
				StartedOn:    time.Now().UTC(),
				FinishedOn:   time.Now().UTC(),
			},
		},
	}

	env, err := signer.SignStatement(st)
	if err != nil {
		t.Fatalf("failed signing DSSE statement: %v", err)
	}

	if env.PayloadType != DSSEPayloadType {
		t.Errorf("unexpected payload type: %s", env.PayloadType)
	}
	if len(env.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(env.Signatures))
	}

	// Verify against statement
	if err := ring.Verify(env, st); err != nil {
		t.Fatalf("expected signature verification to pass, got: %v", err)
	}
}

func TestDSSE_KeyRevocation(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	signer := NewSigner(priv)
	ring := NewKeyRing()
	ring.Add(pub)

	st := &Statement{
		Type:          StatementType,
		PredicateType: SLSAProvenanceV1,
		Subject:       []Subject{{Name: "artifact", Digest: map[string]string{"sha256": "1234"}}},
	}

	env, _ := signer.SignStatement(st)

	// Revoke the key
	ring.Revoke(KeyID(pub))

	// Verify must fail closed!
	if err := ring.Verify(env, st); err == nil {
		t.Errorf("expected verification failure for revoked key, but got success")
	}
}

func TestDSSE_TamperPayload(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	signer := NewSigner(priv)
	ring := NewKeyRing()
	ring.Add(pub)

	st := &Statement{
		Type:          StatementType,
		PredicateType: SLSAProvenanceV1,
		Subject:       []Subject{{Name: "artifact", Digest: map[string]string{"sha256": "1234"}}},
	}

	env, _ := signer.SignStatement(st)

	tamperedStatement := &Statement{
		Type:          StatementType,
		PredicateType: SLSAProvenanceV1,
		Subject:       []Subject{{Name: "tampered_artifact", Digest: map[string]string{"sha256": "9999"}}},
	}

	if err := ring.Verify(env, tamperedStatement); err == nil {
		t.Errorf("expected tamper detection failure, but verification succeeded")
	}
}
