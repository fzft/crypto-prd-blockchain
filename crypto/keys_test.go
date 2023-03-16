package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	prvKey := GeneratePrivateKey()
	assert.Equal(t, PrivateKeyLen, len(prvKey.Bytes()))

	pubKey := prvKey.PublicKey()
	assert.Equal(t, PubKeyLen, len(pubKey.Bytes()))
}

func TestPrivateKeySign(t *testing.T) {
	prvKey := GeneratePrivateKey()
	msg := []byte("hello world")

	sig := prvKey.Sign(msg)
	assert.True(t, sig.Verify(msg, prvKey.PublicKey()))

	assert.False(t, sig.Verify([]byte("hello world!"), prvKey.PublicKey()))

	prvKey2 := GeneratePrivateKey()
	assert.False(t, sig.Verify(msg, prvKey2.PublicKey()))
}

func TestPublicKeyAddress(t *testing.T) {
	prv := GeneratePrivateKey()
	pub := prv.PublicKey()
	addr := pub.Address()
	assert.Equal(t, AddressLen, len(addr.Bytes()))

	t.Log("address:", addr.String())
}
