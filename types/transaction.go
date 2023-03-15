package types

import (
	"crypto/sha256"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	pb "github.com/golang/protobuf/proto"
)

// SignTransaction signs the transaction.
func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func HashTransaction(tx *proto.Transaction) []byte {
	t, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(t)
	return hash[:]
}

// VerifyTransaction verifies the transaction.
func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		sig := crypto.SignatureFromBytes(input.Signature)

		// TODO: this is a hack, we should not modify the transaction
		input.Signature = nil
		if !sig.Verify(HashTransaction(tx), crypto.PublicKeyFromBytes(input.PublicKey)) {
			return false
		}
	}
	return true
}
