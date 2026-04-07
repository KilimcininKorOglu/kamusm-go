package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const (
	configFileName   = ".kamusm-go.conf"
	configSaltPhrase = "kamusm-go-config-salt"
	configKDIter     = 100000
)

// configData holds the saved configuration.
type configData struct {
	Sunucu    string `json:"sunucu"`
	MusteriNo uint32 `json:"musteriNo"`
	Parola    string `json:"parola"`
	Hash      string `json:"hash"`
	Iterasyon int    `json:"iterasyon"`
}

// configPath returns the path to the configuration file in the user's home directory.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("kullanıcı dizini bulunamadı: %w", err)
	}
	return filepath.Join(home, configFileName), nil
}

// machineKey derives a 32-byte AES key from hostname + username + salt phrase.
func machineKey() ([]byte, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("hostname alınamadı: %w", err)
	}

	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("kullanıcı bilgisi alınamadı: %w", err)
	}

	source := hostname + u.Username + configSaltPhrase
	// Use PBKDF2 with the source as password and a fixed salt
	fixedSalt := []byte("kamusm-go-fixed-salt-v1")
	return deriveKey(source, fixedSalt, configKDIter), nil
}

// saveConfig encrypts and saves the configuration to ~/.kamusm-go.conf.
// File format: salt(16) + iv(16) + ciphertext
func saveConfig(cfg configData) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("JSON kodlama hatası: %w", err)
	}

	key, err := machineKey()
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("rastgele salt üretilemedi: %w", err)
	}

	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("rastgele IV üretilemedi: %w", err)
	}

	ciphertext, err := encryptAesCbc(key, iv, jsonData)
	if err != nil {
		return fmt.Errorf("şifreleme hatası: %w", err)
	}

	// salt(16) + iv(16) + ciphertext
	fileData := make([]byte, 0, 32+len(ciphertext))
	fileData = append(fileData, salt...)
	fileData = append(fileData, iv...)
	fileData = append(fileData, ciphertext...)

	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, fileData, 0600); err != nil {
		return fmt.Errorf("yapılandırma dosyası yazılamadı: %w", err)
	}

	return nil
}

// loadConfig reads and decrypts the configuration from ~/.kamusm-go.conf.
func loadConfig() (*configData, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("yapılandırma dosyası okunamadı: %w", err)
	}

	if len(data) < 33 { // salt(16) + iv(16) + min 1 byte ciphertext
		return nil, fmt.Errorf("yapılandırma dosyası bozuk")
	}

	iv := data[16:32]
	ciphertext := data[32:]

	key, err := machineKey()
	if err != nil {
		return nil, err
	}

	jsonData, err := decryptAesCbc(key, iv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("yapılandırma çözülemedi: %w", err)
	}

	var cfg configData
	if err := json.Unmarshal(jsonData, &cfg); err != nil {
		return nil, fmt.Errorf("yapılandırma ayrıştırılamadı: %w", err)
	}

	return &cfg, nil
}

// maskPassword returns a masked version of the password for display.
func maskPassword(p string) string {
	if len(p) <= 3 {
		return "***"
	}
	return p[:3] + "****"
}
