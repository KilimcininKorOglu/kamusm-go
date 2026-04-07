package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

// pkcs7Pad applies PKCS#7 padding with 16-byte block size.
func pkcs7Pad(data []byte) []byte {
	blockSize := aes.BlockSize
	padLen := blockSize - (len(data) % blockSize)
	padded := make([]byte, len(data)+padLen)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}
	return padded
}

// deriveKey derives a 32-byte AES-256 key using PBKDF2-HMAC-SHA256.
func deriveKey(password string, salt []byte, iterations int) []byte {
	return pbkdf2.Key([]byte(password), salt, iterations, 32, sha256.New)
}

// encryptAesCbc encrypts plaintext using AES-256 in CBC mode with PKCS#7 padding.
func encryptAesCbc(key, iv, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("geçersiz anahtar uzunluğu: %d (32 olmalı)", len(key))
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("geçersiz IV uzunluğu: %d (%d olmalı)", len(iv), aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(plaintext)
	ciphertext := make([]byte, len(padded))

	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

// pkcs7Unpad removes PKCS#7 padding from the data.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("boş veri")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > aes.BlockSize || padLen > len(data) {
		return nil, fmt.Errorf("geçersiz PKCS#7 padding")
	}
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, fmt.Errorf("geçersiz PKCS#7 padding")
		}
	}
	return data[:len(data)-padLen], nil
}

// decryptAesCbc decrypts ciphertext using AES-256 in CBC mode and removes PKCS#7 padding.
func decryptAesCbc(key, iv, ciphertext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("geçersiz anahtar uzunluğu: %d (32 olmalı)", len(key))
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("geçersiz IV uzunluğu: %d (%d olmalı)", len(iv), aes.BlockSize)
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("geçersiz ciphertext uzunluğu")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext)
}
