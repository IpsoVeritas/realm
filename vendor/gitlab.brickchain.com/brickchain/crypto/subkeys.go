package crypto

import (
	"encoding/base64"
	"encoding/json"
	"time"

	jose "gopkg.in/square/go-jose.v1"

	hash "crypto"

	"fmt"

	"gitlab.brickchain.com/brickchain/document"
)

func VerifyCertificateChain(chain string, keyLevel int) (*document.CertificateChain, error) {
	certChainJWS, err := UnmarshalSignature([]byte(chain))
	if err != nil {
		return nil, fmt.Errorf("Invalid certificate chain JWS")
	}
	if len(certChainJWS.Signatures) < 1 {
		return nil, fmt.Errorf("Invalid header of JWS, does not contain signing key")
	}
	if certChainJWS.Signatures[0].Header.JsonWebKey == nil {
		return nil, fmt.Errorf("Invalid header of JWS, does not contain signing key")
	}

	signingRootTPbytes, _ := certChainJWS.Signatures[0].Header.JsonWebKey.Thumbprint(hash.SHA256)
	signingRootTP := base64.URLEncoding.EncodeToString(signingRootTPbytes)

	payload, err := certChainJWS.Verify(certChainJWS.Signatures[0].Header.JsonWebKey)
	if err != nil {
		return nil, fmt.Errorf("Invalid signature of certificate chain")
	}

	certificateChain := document.CertificateChain{}
	err = json.Unmarshal(payload, &certificateChain)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal certificate chain")
	}

	if certificateChain.KeyLevel > keyLevel {
		return nil, fmt.Errorf("Key level %d is higher than allowed level of %d", certificateChain.KeyLevel, keyLevel)
	}

	rootTPbytes, _ := certificateChain.Root.Thumbprint(hash.SHA256)
	rootTP := base64.URLEncoding.EncodeToString(rootTPbytes)
	if rootTP != signingRootTP {
		return nil, fmt.Errorf("Chain was not signed by root key specified in chain")
	}

	if certificateChain.TTL != 0 && certificateChain.Timestamp.Add(time.Second*time.Duration(certificateChain.TTL)).Before(time.Now().UTC()) {
		return nil, fmt.Errorf("Certificate chain has expired")
	}

	return &certificateChain, nil
}

func CreateCertificateChain(rootKey, subKey *jose.JsonWebKey, keyLevel int, documentTypes []string, ttl int) (string, error) {
	rootPK, err := NewPublicKey(rootKey)
	if err != nil {
		return "", err
	}

	chain := &document.CertificateChain{
		Timestamp:     time.Now().UTC(),
		TTL:           ttl,
		Root:          rootPK,
		SubKey:        subKey,
		DocumentTypes: documentTypes,
		KeyLevel:      keyLevel,
	}

	signer, err := NewSigner(rootKey)
	if err != nil {
		return "", err
	}

	chainBytes, err := json.Marshal(chain)
	if err != nil {
		return "", err
	}

	sig, err := signer.Sign(chainBytes)
	if err != nil {
		return "", err
	}

	sigString, err := sig.CompactSerialize()
	if err != nil {
		return "", err
	}

	return sigString, nil
}
