package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	privateKeyLen = 64
	pubKeyLen     = 32
	seedLen
	addressLen = 20
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func GeneratePrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != seedLen {
		panic("invalid seed length, must be 32 bytes")
	}
	return &PrivateKey{key: ed25519.NewKeyFromSeed(seed)}
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, seedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}
	return &PrivateKey{key: ed25519.NewKeyFromSeed(seed)}
}

// Bytes returns the private key as a byte slice.
func (k *PrivateKey) Bytes() []byte {
	return k.key
}

// Sign signs the given message with the private key.
func (k *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		sig: ed25519.Sign(k.key, msg),
	}
}

// PublicKey returns the public key associated with the private key.
func (k *PrivateKey) PublicKey() *PublicKey {
	b := make([]byte, pubKeyLen)
	copy(b, k.key[32:])
	return &PublicKey{key: b}
}

type PublicKey struct {
	key ed25519.PublicKey
}

// Address ...
func (k *PublicKey) Address() Address {
	return Address{addr: k.key[len(k.key)-addressLen:]}
}

// Bytes returns the public key as a byte slice.
func (k *PublicKey) Bytes() []byte {
	return k.key
}

type Signature struct {
	sig []byte
}

// Bytes returns the signature as a byte slice.
func (s *Signature) Bytes() []byte {
	return s.sig
}

// Verify verifies the signature against the given message and public key.
func (s *Signature) Verify(msg []byte, pubKey *PublicKey) bool {
	return ed25519.Verify(pubKey.key, msg, s.sig)
}

type Address struct {
	addr []byte
}

// Bytes returns the address as a byte slice.
func (a Address) Bytes() []byte {
	return a.addr
}

func (a Address) String() string {
	return hex.EncodeToString(a.addr)
}
