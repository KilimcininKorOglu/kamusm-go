package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
	"time"
)

var (
	oidSHA1   = asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 5}
	oidSHA256 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
)

// algorithmIdentifier represents the AlgorithmIdentifier ASN.1 structure.
type algorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

// messageImprint represents the MessageImprint ASN.1 structure.
type messageImprint struct {
	HashAlgorithm algorithmIdentifier
	HashedMessage []byte
}

// timeStampReq represents the RFC 3161 TimeStampReq ASN.1 structure.
type timeStampReq struct {
	Version        int
	MessageImprint messageImprint
	Nonce          int64
}

// computeFileDigest computes the hash of a file using the specified algorithm.
func computeFileDigest(path, alg string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("dosya okunamadı: %w", err)
	}
	defer f.Close()

	var h hash.Hash
	switch strings.ToLower(alg) {
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	default:
		return nil, fmt.Errorf("desteklenmeyen hash algoritması: %s", alg)
	}

	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("dosya okunamadı: %w", err)
	}

	return h.Sum(nil), nil
}

// buildTsaRequest creates an RFC 3161 TimeStampReq DER-encoded structure.
func buildTsaRequest(digest []byte, hashAlg string) ([]byte, error) {
	var oid asn1.ObjectIdentifier
	switch strings.ToLower(hashAlg) {
	case "sha1":
		oid = oidSHA1
	case "sha256":
		oid = oidSHA256
	default:
		return nil, fmt.Errorf("desteklenmeyen hash algoritması: %s", hashAlg)
	}

	nonce := time.Now().UnixMilli() & ((1 << 63) - 1)

	req := timeStampReq{
		Version: 1,
		MessageImprint: messageImprint{
			HashAlgorithm: algorithmIdentifier{
				Algorithm:  oid,
				Parameters: asn1.RawValue{Tag: asn1.TagNull},
			},
			HashedMessage: digest,
		},
		Nonce: nonce,
	}

	der, err := asn1.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ASN.1 DER kodlama hatası: %w", err)
	}

	return der, nil
}
