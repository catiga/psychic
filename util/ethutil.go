package util

import (
	ml "eli/middleware"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func Verify(signatureStr string, hashStr string, verifyAddr string) bool {

	signature := hexutil.MustDecode(signatureStr)
	signature[64] -= 27
	hash := crypto.Keccak256Hash([]byte(hashStr))

	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)

	if err != nil {
		ml.Log.Info("err : ", err)
		return false
	}

	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery ID
	verified := crypto.VerifySignature(sigPublicKey, hash.Bytes(), signatureNoRecoverID)

	if verified {

		sigPublicKeyECDSA, err2 := crypto.SigToPub(hash.Bytes(), signature)

		if err2 != nil {
			return false
		}

		addrECDSA := crypto.PubkeyToAddress(*sigPublicKeyECDSA)

		return strings.EqualFold(addrECDSA.Hex(), verifyAddr)
	}
	return verified
}
