package rest

import (
	"github.com/IpsoVeritas/crypto"
	httphandler "github.com/IpsoVeritas/httphandler"
	jose "gopkg.in/square/go-jose.v1"
)

func hasMandateForRealm(mandates []httphandler.AuthenticatedMandate, realmID string) bool {
	for _, m := range mandates {
		if m.Mandate.Realm == realmID {
			return true
		}
	}

	return false
}

func validSig(keyset *jose.JsonWebKeySet, sig *jose.JsonWebSignature) bool {
	for _, key := range keyset.Keys {
		pkey, err := crypto.NewPublicKey(&key)
		if err == nil {
			var count = 0
			count, _, _, err = sig.VerifyMulti(pkey.Key)
			if err == nil {
				if count >= 0 {
					return true
				}
			}
		}
	}
	return false
}
