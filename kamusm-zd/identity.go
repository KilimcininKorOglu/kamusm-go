package kamusmzd

import (
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"
)

// EsyaReqEx represents the ASN.1 structure for the identity header.
type EsyaReqEx struct {
	UserID                  int
	Salt                    []byte
	IterationCount          int
	IV                      []byte
	EncryptedMessageImprint []byte
}

// BuildIdentity creates the identity header for KamuSM authentication.
// Returns a hex string of the DER-encoded ASN.1 structure interpreted as a big integer.
func BuildIdentity(customerID uint32, password string, messageImprint []byte, iterations int) (string, error) {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("rastgele IV üretilemedi: %w", err)
	}
	salt := make([]byte, 16)
	copy(salt, iv)

	key := deriveKey(password, salt, iterations)

	ciphertext, err := encryptAesCbc(key, iv, messageImprint)
	if err != nil {
		return "", fmt.Errorf("şifreleme hatası: %w", err)
	}

	token := EsyaReqEx{
		UserID:                  int(customerID),
		Salt:                    salt,
		IterationCount:          iterations,
		IV:                      iv,
		EncryptedMessageImprint: ciphertext,
	}

	der, err := asn1.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("ASN.1 DER kodlama hatası: %w", err)
	}

	n := new(big.Int).SetBytes(der)
	return fmt.Sprintf("%x", n), nil
}
