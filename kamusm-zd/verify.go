package kamusmzd

import (
	"crypto/x509"
	"fmt"
	"time"

	"go.mozilla.org/pkcs7"
)

// VerifyResult holds the result of a timestamp verification.
type VerifyResult struct {
	Valid  bool      `json:"gecerli"`
	Signer string   `json:"imzalayan,omitempty"`
	Date   time.Time `json:"tarih,omitempty"`
	Error  string   `json:"hata,omitempty"`
}

// VerifyTimestamp verifies the PKCS#7 SignedData in the given DER data
// against KamuSM root CA certificates.
func VerifyTimestamp(derData []byte) (*VerifyResult, error) {
	p7, err := pkcs7.Parse(derData)
	if err != nil {
		return nil, fmt.Errorf("PKCS#7 ayrıştırma hatası: %w", err)
	}

	if err := p7.Verify(); err != nil {
		return &VerifyResult{
			Valid: false,
			Error: fmt.Sprintf("imza doğrulama başarısız: %v", err),
		}, nil
	}

	roots := KamusmRootCAs()
	var signer *x509.Certificate

	for _, cert := range p7.Certificates {
		if !cert.IsCA {
			signer = cert
			break
		}
	}

	if signer == nil && len(p7.Certificates) > 0 {
		signer = p7.Certificates[0]
	}

	result := &VerifyResult{Valid: true}

	if signer != nil {
		result.Signer = signer.Subject.CommonName

		if !signer.NotBefore.IsZero() {
			result.Date = time.Now()
		}

		intermediates := x509.NewCertPool()
		for _, cert := range p7.Certificates {
			if cert != signer {
				intermediates.AddCert(cert)
			}
		}

		opts := x509.VerifyOptions{
			Roots:         roots,
			Intermediates: intermediates,
			CurrentTime:   signer.NotBefore.Add(time.Hour),
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		}

		if _, err := signer.Verify(opts); err != nil {
			result.Valid = false
			result.Error = fmt.Sprintf("sertifika zinciri doğrulanamadı: %v", err)
		}
	}

	return result, nil
}
