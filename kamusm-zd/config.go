package kamusmzd

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

// ConfigData holds the saved configuration.
type ConfigData struct {
	Sunucu    string `json:"sunucu"`
	MusteriNo uint32 `json:"musteriNo"`
	Parola    string `json:"parola"`
	Hash      string `json:"hash"`
	Iterasyon int    `json:"iterasyon"`
}

// ConfigPath returns the path to the configuration file in the user's home directory.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("kullanıcı dizini bulunamadı: %w", err)
	}
	return filepath.Join(home, configFileName), nil
}

func machineKey(salt []byte) ([]byte, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("hostname alınamadı: %w", err)
	}

	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("kullanıcı bilgisi alınamadı: %w", err)
	}

	source := hostname + u.Username + configSaltPhrase
	return deriveKey(source, salt, configKDIter), nil
}

// SaveConfig encrypts and saves the configuration to ~/.kamusm-go.conf.
func SaveConfig(cfg ConfigData) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("JSON kodlama hatası: %w", err)
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("rastgele salt üretilemedi: %w", err)
	}

	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("rastgele IV üretilemedi: %w", err)
	}

	key, err := machineKey(salt)
	if err != nil {
		return err
	}

	ciphertext, err := encryptAesCbc(key, iv, jsonData)
	if err != nil {
		return fmt.Errorf("şifreleme hatası: %w", err)
	}

	fileData := make([]byte, 0, 32+len(ciphertext))
	fileData = append(fileData, salt...)
	fileData = append(fileData, iv...)
	fileData = append(fileData, ciphertext...)

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, fileData, 0600); err != nil {
		return fmt.Errorf("yapılandırma dosyası yazılamadı: %w", err)
	}

	return nil
}

// LoadConfig reads and decrypts the configuration from ~/.kamusm-go.conf.
func LoadConfig() (*ConfigData, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("yapılandırma dosyası okunamadı: %w", err)
	}

	if len(data) < 33 {
		return nil, fmt.Errorf("yapılandırma dosyası bozuk")
	}

	salt := data[0:16]
	iv := data[16:32]
	ciphertext := data[32:]

	key, err := machineKey(salt)
	if err != nil {
		return nil, err
	}

	jsonData, err := decryptAesCbc(key, iv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("yapılandırma çözülemedi: %w", err)
	}

	var cfg ConfigData
	if err := json.Unmarshal(jsonData, &cfg); err != nil {
		return nil, fmt.Errorf("yapılandırma ayrıştırılamadı: %w", err)
	}

	return &cfg, nil
}

// MaskPassword returns a masked version of the password for display.
func MaskPassword(p string) string {
	if len(p) <= 3 {
		return "***"
	}
	return p[:3] + "****"
}
