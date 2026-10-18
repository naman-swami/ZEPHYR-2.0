package capsule

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ============================================================================
// CANONICAL DETERMINISTIC JSON (JCS SUBSET VIA STDLIB)
// ============================================================================

// CanonicalJSON marshals any value into canonical deterministic JSON
// with sorted object keys, no insignificant whitespace, and exact number preservation.
func CanonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var generic any
	if err := dec.Decode(&generic); err != nil {
		return nil, err
	}
	return json.Marshal(generic) // Go stdlib encoding/json sorts map keys deterministically
}

// ============================================================================
// IN-TOTO STATEMENT V1 & SLSA PROVENANCE V1
// ============================================================================

const (
	StatementType     = "https://in-toto.io/Statement/v1"
	SLSAProvenanceV1  = "https://slsa.dev/provenance/v1"
	DSSEPayloadType   = "application/vnd.in-toto+json"
)

type Statement struct {
	Type          string              `json:"_type"`
	Subject       []Subject           `json:"subject"`
	PredicateType string              `json:"predicateType"`
	Predicate     ProvenancePredicate `json:"predicate"`
}

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"` // e.g. {"sha256": hex}
}

type ResourceDescriptor struct {
	URI    string            `json:"uri,omitempty"`
	Digest map[string]string `json:"digest,omitempty"`
	Name   string            `json:"name,omitempty"`
}

type BuildDefinition struct {
	BuildType            string               `json:"buildType"`
	ExternalParameters   map[string]string    `json:"externalParameters"`
	InternalParameters   map[string]string    `json:"internalParameters"`
	ResolvedDependencies []ResourceDescriptor `json:"resolvedDependencies"`
}

type Builder struct {
	ID      string            `json:"id"`
	Version map[string]string `json:"version,omitempty"`
}

type RunDetails struct {
	Builder      Builder              `json:"builder"`
	InvocationID string               `json:"invocationId"`
	StartedOn    time.Time            `json:"startedOn"`
	FinishedOn   time.Time            `json:"finishedOn"`
	Byproducts   []ResourceDescriptor `json:"byproducts,omitempty"`
}

type ProvenancePredicate struct {
	BuildDefinition BuildDefinition `json:"buildDefinition"`
	RunDetails      RunDetails      `json:"runDetails"`
}

func NewInvocationID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // UUID v4
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ============================================================================
// DEAD SIMPLE SIGNING ENVELOPE (DSSE) & PRE-AUTHENTICATION ENCODING (PAE)
// ============================================================================

type DSSEEnvelope struct {
	PayloadType string    `json:"payloadType"`
	Payload     string    `json:"payload"` // Base64-encoded canonical statement
	Signatures  []DSSESig `json:"signatures"`
}

type DSSESig struct {
	KeyID string `json:"keyid"`
	Sig   string `json:"sig"` // Base64-encoded Ed25519 signature over PAE
}

// PAE computes the DSSE Pre-Authentication Encoding:
// "DSSEv1" <len(payloadType)> <payloadType> <len(payload)> <payload>
func PAE(payloadType string, payload []byte) []byte {
	return []byte(fmt.Sprintf("DSSEv1 %d %s %d %s",
		len(payloadType), payloadType, len(payload), payload))
}

// ============================================================================
// KEY RING & SIGNATURE LIFECYCLE
// ============================================================================

type KeyRing struct {
	keys    map[string]ed25519.PublicKey
	revoked map[string]struct{}
}

func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])
}

func NewKeyRing() *KeyRing {
	return &KeyRing{
		keys:    make(map[string]ed25519.PublicKey),
		revoked: make(map[string]struct{}),
	}
}

func (kr *KeyRing) Add(pub ed25519.PublicKey) {
	kr.keys[KeyID(pub)] = pub
}

func (kr *KeyRing) Revoke(keyID string) {
	kr.revoked[keyID] = struct{}{}
}

type Signer struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func NewSigner(priv ed25519.PrivateKey) *Signer {
	return &Signer{
		priv: priv,
		pub:  priv.Public().(ed25519.PublicKey),
	}
}

func (s *Signer) SignStatement(st *Statement) (*DSSEEnvelope, error) {
	if st.Type != StatementType || st.PredicateType != SLSAProvenanceV1 {
		return nil, errors.New("statement must be in-toto v1 with SLSA provenance v1")
	}

	payload, err := CanonicalJSON(st)
	if err != nil {
		return nil, fmt.Errorf("canonicalize statement: %w", err)
	}

	sig := ed25519.Sign(s.priv, PAE(DSSEPayloadType, payload))
	return &DSSEEnvelope{
		PayloadType: DSSEPayloadType,
		Payload:     base64.StdEncoding.EncodeToString(payload),
		Signatures: []DSSESig{
			{
				KeyID: KeyID(s.pub),
				Sig:   base64.StdEncoding.EncodeToString(sig),
			},
		},
	}, nil
}

func (kr *KeyRing) Verify(env *DSSEEnvelope, st *Statement) error {
	if env.PayloadType != DSSEPayloadType {
		return fmt.Errorf("unexpected payload type: %s", env.PayloadType)
	}

	payload, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		return fmt.Errorf("invalid base64 payload: %w", err)
	}

	if len(env.Signatures) == 0 {
		return errors.New("envelope has no signatures")
	}

	verified := false
	for _, s := range env.Signatures {
		if _, isRevoked := kr.revoked[s.KeyID]; isRevoked {
			continue // Skip revoked keys
		}
		pub, ok := kr.keys[s.KeyID]
		if !ok {
			continue // Skip unknown keys
		}

		sig, err := base64.StdEncoding.DecodeString(s.Sig)
		if err != nil {
			continue
		}

		if ed25519.Verify(pub, PAE(env.PayloadType, payload), sig) {
			verified = true
			break
		}
	}

	if !verified {
		return errors.New("no valid signature from a trusted, non-revoked ring key")
	}

	if st != nil {
		want, err := CanonicalJSON(st)
		if err != nil {
			return fmt.Errorf("failed to canonicalize expected statement: %w", err)
		}
		if !bytes.Equal(want, payload) {
			return errors.New("statement does not match signed payload")
		}
	}

	return nil
}
