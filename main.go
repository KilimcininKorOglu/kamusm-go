package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	kamusmzd "github.com/KilimcininKorOglu/kamusm-go/kamusm-zd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "kimlik":
		runIdentity(os.Args[2:])
	case "gonder":
		runSend(os.Args[2:])
	case "bakiye":
		runCredits(os.Args[2:])
	case "dogrula":
		runVerify(os.Args[2:])
	case "ayar-kaydet":
		runSaveConfig(os.Args[2:])
	case "ayar-goster":
		runShowConfig()
	case "versiyon", "--version", "-v":
		fmt.Printf("kamusm-go %s\n", kamusmzd.Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Bilinmeyen komut: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "kamusm-go - Go ile yazılmış KamuSM zaman damgası istemcisi")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Kullanım:")
	fmt.Fprintln(os.Stderr, "  kamusm-go <komut> [seçenekler]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Komutlar:")
	fmt.Fprintln(os.Stderr, "  kimlik      \"identity\" başlığı oluştur")
	fmt.Fprintln(os.Stderr, "  gonder      Zaman damgası isteği gönder")
	fmt.Fprintln(os.Stderr, "  bakiye      Bakiyeyi kontrol et")
	fmt.Fprintln(os.Stderr, "  dogrula     Zaman damgası dosyasını doğrula")
	fmt.Fprintln(os.Stderr, "  ayar-kaydet Bağlantı bilgilerini şifreli kaydet")
	fmt.Fprintln(os.Stderr, "  ayar-goster Kayıtlı ayarları göster")
}

func runIdentity(args []string) {
	fs := flag.NewFlagSet("identity", flag.ExitOnError)
	musteriNo := fs.Uint("musteri-no", 0, "Müşteri numarası")
	parola := fs.String("parola", "", "Müşteri şifresi")
	ozetHex := fs.String("ozet-hex", "", "Önceden hesaplanmış özet (hex)")
	zaman := fs.Uint64("zaman", 0, "Unix zaman damgası (milisaniye)")
	iterasyon := fs.Int("iterasyon", 100, "PBKDF2 iterasyon sayısı")
	jsonOut := fs.Bool("json", false, "Çıktıyı JSON formatında ver")
	fs.Parse(args)

	applyConfigDefaults(nil, musteriNo, parola, nil, iterasyon)

	if *musteriNo == 0 {
		fatal("--musteri-no parametresi gereklidir")
	}
	if *parola == "" {
		fatal("--parola parametresi gereklidir")
	}
	if *iterasyon < 1 {
		fatal("--iterasyon değeri en az 1 olmalıdır")
	}

	var digest []byte

	if *ozetHex != "" {
		var err error
		digest, err = hex.DecodeString(*ozetHex)
		if err != nil {
			fatal("Geçersiz hex özet: %v", err)
		}
	} else if *zaman != 0 {
		s := fmt.Sprintf("%d%d", *musteriNo, *zaman)
		h := sha1.Sum([]byte(s))
		digest = h[:]
	} else {
		fatal("--ozet-hex veya --zaman parametrelerinden biri sağlanmalıdır")
	}

	identity, err := kamusmzd.BuildIdentity(uint32(*musteriNo), *parola, digest, *iterasyon)
	if err != nil {
		fatal("Identity oluşturulamadı: %v", err)
	}

	if *jsonOut {
		printJSON(map[string]any{"identity": identity})
	} else {
		fmt.Println(identity)
	}
}

func runSend(args []string) {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	sunucu := fs.String("sunucu", "", "Sunucu URL'si")
	musteriNo := fs.Uint("musteri-no", 0, "Müşteri numarası")
	parola := fs.String("parola", "", "Müşteri şifresi")
	dosya := fs.String("dosya", "", "Zaman damgası alınacak dosya yolu")
	ozetHex := fs.String("ozet-hex", "", "Önceden hesaplanmış özet (hex)")
	hashAlg := fs.String("hash", "sha256", "Hash algoritması (sha1 veya sha256)")
	iterasyon := fs.Int("iterasyon", 100, "PBKDF2 iterasyon sayısı")
	jsonOut := fs.Bool("json", false, "Çıktıyı JSON formatında ver")
	dogrula := fs.Bool("dogrula", false, "Kaydedilen dosyayı KamuSM sertifikalarıyla doğrula")
	fs.Parse(args)

	applyConfigDefaults(sunucu, musteriNo, parola, hashAlg, iterasyon)

	if *sunucu == "" {
		fatal("--sunucu parametresi gereklidir")
	}
	if *musteriNo == 0 {
		fatal("--musteri-no parametresi gereklidir")
	}
	if *parola == "" {
		fatal("--parola parametresi gereklidir")
	}
	if *iterasyon < 1 {
		fatal("--iterasyon değeri en az 1 olmalıdır")
	}

	var digest []byte
	var outputFilename string

	if *dosya != "" {
		var err error
		digest, err = kamusmzd.ComputeFileDigest(*dosya, *hashAlg)
		if err != nil {
			fatal("Dosya hash'i hesaplanamadı: %v", err)
		}
		stem := fileNameWithoutExt(*dosya)
		dir := filepath.Dir(*dosya)
		outputFilename = filepath.Join(dir, stem+"_zd.der")
	} else if *ozetHex != "" {
		var err error
		digest, err = hex.DecodeString(*ozetHex)
		if err != nil {
			fatal("Geçersiz hex özet: %v", err)
		}
		ts := time.Now().Unix()
		outputFilename = fmt.Sprintf("zd_%d.der", ts)
	} else {
		fatal("--dosya veya --ozet-hex parametrelerinden biri sağlanmalıdır")
	}

	der, err := kamusmzd.BuildTsaRequest(digest, *hashAlg)
	if err != nil {
		fatal("TSA isteği oluşturulamadı: %v", err)
	}

	identity, err := kamusmzd.BuildIdentity(uint32(*musteriNo), *parola, digest, *iterasyon)
	if err != nil {
		fatal("Identity oluşturulamadı: %v", err)
	}

	status, body, err := kamusmzd.SendTimestampRequest(*sunucu, identity, der)
	if err != nil {
		fatal("İstek gönderilemedi: %v", err)
	}

	if kamusmzd.IsValidTimestampResponse(body) {
		var saved bool
		if pkcs7Data := kamusmzd.ExtractPkcs7(body); pkcs7Data != nil {
			if err := os.WriteFile(outputFilename, pkcs7Data, 0644); err != nil {
				fatal("Yanıt yazılamadı: %v", err)
			}
			saved = true
		} else {
			if err := os.WriteFile(outputFilename, body, 0644); err != nil {
				fatal("Yanıt yazılamadı: %v", err)
			}
			saved = true
		}

		if *jsonOut {
			result := map[string]any{"durum": status, "basarili": true, "dosya": outputFilename}
			if *dogrula {
				savedData, _ := os.ReadFile(outputFilename)
				vr, err := kamusmzd.VerifyTimestamp(savedData)
				if err != nil {
					result["dogrulama"] = map[string]any{"gecerli": false, "hata": err.Error()}
				} else {
					result["dogrulama"] = vr
				}
			}
			printJSON(result)
		} else {
			fmt.Printf("Yanıt durumu: %d\n", status)
			if saved {
				fmt.Printf("Çıkarılan PKCS#7 SignedData %s dosyasına kaydedildi\n", outputFilename)
			}
			if *dogrula {
				savedData, _ := os.ReadFile(outputFilename)
				vr, err := kamusmzd.VerifyTimestamp(savedData)
				if err != nil {
					fmt.Printf("Doğrulama hatası: %v\n", err)
				} else if vr.Valid {
					fmt.Println("Doğrulama başarılı")
					if vr.Signer != "" {
						fmt.Printf("  İmzalayan: %s\n", vr.Signer)
					}
				} else {
					fmt.Printf("Doğrulama başarısız: %s\n", vr.Error)
				}
			}
		}
	} else {
		texts := kamusmzd.ExtractTextFromAsn1(body)

		if *jsonOut {
			result := map[string]any{"durum": status, "basarili": false}
			if len(texts) > 0 {
				result["hatalar"] = texts
			} else {
				text := string(body)
				if isPrintableString(text) {
					result["hatalar"] = []string{strings.TrimSpace(text)}
				} else {
					if err := os.WriteFile(outputFilename, body, 0644); err != nil {
						fatal("Yanıt yazılamadı: %v", err)
					}
					result["dosya"] = outputFilename
				}
			}
			printJSON(result)
		} else {
			fmt.Printf("Yanıt durumu: %d\n", status)
			fmt.Printf("Hata yanıtı alındı (HTTP %d)\n", status)
			if len(texts) > 0 {
				fmt.Println("Hata mesajları:")
				for _, text := range texts {
					fmt.Printf("  %s\n", text)
				}
			} else {
				text := string(body)
				if isPrintableString(text) {
					fmt.Printf("Yanıt gövdesi (metin):\n%s\n", text)
				} else {
					if err := os.WriteFile(outputFilename, body, 0644); err != nil {
						fatal("Yanıt yazılamadı: %v", err)
					}
					fmt.Printf("Binary hata yanıtı %s dosyasına kaydedildi\n", outputFilename)
				}
			}
		}
	}
}

func runCredits(args []string) {
	fs := flag.NewFlagSet("credits", flag.ExitOnError)
	sunucu := fs.String("sunucu", "", "Sunucu URL'si")
	musteriNo := fs.Uint("musteri-no", 0, "Müşteri numarası")
	parola := fs.String("parola", "", "Müşteri şifresi")
	iterasyon := fs.Int("iterasyon", 100, "PBKDF2 iterasyon sayısı")
	zaman := fs.Uint64("zaman", 0, "Override zaman damgası (milisaniye)")
	jsonOut := fs.Bool("json", false, "Çıktıyı JSON formatında ver")
	fs.Parse(args)

	applyConfigDefaults(sunucu, musteriNo, parola, nil, iterasyon)

	if *sunucu == "" {
		fatal("--sunucu parametresi gereklidir")
	}
	if *musteriNo == 0 {
		fatal("--musteri-no parametresi gereklidir")
	}
	if *parola == "" {
		fatal("--parola parametresi gereklidir")
	}
	if *iterasyon < 1 {
		fatal("--iterasyon değeri en az 1 olmalıdır")
	}

	ts := *zaman
	if ts == 0 {
		ts = uint64(time.Now().UnixMilli())
	}

	s := fmt.Sprintf("%d%d", *musteriNo, ts)
	h := sha1.Sum([]byte(s))
	digest := h[:]

	identity, err := kamusmzd.BuildIdentity(uint32(*musteriNo), *parola, digest, *iterasyon)
	if err != nil {
		fatal("Identity oluşturulamadı: %v", err)
	}

	status, contentType, body, err := kamusmzd.SendCreditRequest(*sunucu, identity, uint32(*musteriNo), ts)
	if err != nil {
		fatal("Bakiye kontrolü isteği gönderilemedi: %v", err)
	}

	if *jsonOut {
		result := map[string]any{"durum": status}
		if strings.HasPrefix(contentType, "application/timestamp-reply") {
			if credits, ok := kamusmzd.ParseCreditsFromBody(body); ok {
				result["bakiye"] = credits
			} else {
				result["hata"] = strings.TrimSpace(string(body))
			}
		} else {
			result["hata"] = strings.TrimSpace(string(body))
		}
		printJSON(result)
	} else {
		fmt.Printf("Yanıt durumu: %d\n", status)

		if strings.HasPrefix(contentType, "application/timestamp-reply") {
			if credits, ok := kamusmzd.ParseCreditsFromBody(body); ok {
				fmt.Printf("Kalan zaman damgası bakiyesi: %d\n", credits)
			} else {
				text := string(body)
				if isPrintableString(text) {
					fmt.Printf("Yanıt gövdesi (metin):\n%s\n", text)
				} else {
					if err := os.WriteFile("timestamp_resp.der", body, 0644); err != nil {
						fatal("Yanıt yazılamadı: %v", err)
					}
					fmt.Println("Binary yanıt; timestamp_resp.der dosyasına kaydedildi")
				}
			}
		} else {
			fmt.Printf("Content-Type: %s\n", contentType)
			text := string(body)
			if isPrintableString(text) {
				fmt.Printf("Yanıt gövdesi (metin):\n%s\n", text)
			} else {
				if err := os.WriteFile("timestamp_resp.der", body, 0644); err != nil {
					fatal("Yanıt yazılamadı: %v", err)
				}
				fmt.Println("Binary yanıt; timestamp_resp.der dosyasına kaydedildi")
			}
		}
	}
}

func runVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	dosya := fs.String("dosya", "", "Doğrulanacak .der dosyası")
	jsonOut := fs.Bool("json", false, "Çıktıyı JSON formatında ver")
	fs.Parse(args)

	if *dosya == "" {
		fatal("--dosya parametresi gereklidir")
	}

	data, err := os.ReadFile(*dosya)
	if err != nil {
		fatal("Dosya okunamadı: %v", err)
	}

	result, err := kamusmzd.VerifyTimestamp(data)
	if err != nil {
		if *jsonOut {
			printJSON(map[string]any{"gecerli": false, "hata": err.Error()})
		} else {
			fatal("Doğrulama hatası: %v", err)
		}
		return
	}

	if *jsonOut {
		printJSON(result)
	} else {
		if result.Valid {
			fmt.Println("Doğrulama başarılı")
			if result.Signer != "" {
				fmt.Printf("  İmzalayan: %s\n", result.Signer)
			}
		} else {
			fmt.Printf("Doğrulama başarısız: %s\n", result.Error)
		}
	}
}

func runSaveConfig(args []string) {
	fs := flag.NewFlagSet("ayar-kaydet", flag.ExitOnError)
	sunucu := fs.String("sunucu", "", "Sunucu URL'si")
	musteriNo := fs.Uint("musteri-no", 0, "Müşteri numarası")
	parola := fs.String("parola", "", "Müşteri şifresi")
	hashAlg := fs.String("hash", "sha256", "Hash algoritması")
	iterasyon := fs.Int("iterasyon", 100, "PBKDF2 iterasyon sayısı")
	fs.Parse(args)

	if *sunucu == "" {
		fatal("--sunucu parametresi gereklidir")
	}
	if *musteriNo == 0 {
		fatal("--musteri-no parametresi gereklidir")
	}
	if *parola == "" {
		fatal("--parola parametresi gereklidir")
	}
	if *iterasyon < 1 {
		fatal("--iterasyon değeri en az 1 olmalıdır")
	}

	cfg := kamusmzd.ConfigData{
		Sunucu:    *sunucu,
		MusteriNo: uint32(*musteriNo),
		Parola:    *parola,
		Hash:      *hashAlg,
		Iterasyon: *iterasyon,
	}

	if err := kamusmzd.SaveConfig(cfg); err != nil {
		fatal("Ayarlar kaydedilemedi: %v", err)
	}

	path, _ := kamusmzd.ConfigPath()
	fmt.Printf("Ayarlar şifreli olarak kaydedildi: %s\n", path)
}

func runShowConfig() {
	cfg, err := kamusmzd.LoadConfig()
	if err != nil {
		fatal("Ayarlar okunamadı: %v", err)
	}

	fmt.Println("Kayıtlı ayarlar:")
	fmt.Printf("  Sunucu:     %s\n", cfg.Sunucu)
	fmt.Printf("  Müşteri No: %d\n", cfg.MusteriNo)
	fmt.Printf("  Parola:     %s\n", kamusmzd.MaskPassword(cfg.Parola))
	fmt.Printf("  Hash:       %s\n", cfg.Hash)
	fmt.Printf("  İterasyon:  %d\n", cfg.Iterasyon)
}

func applyConfigDefaults(sunucu *string, musteriNo *uint, parola *string, hashAlg *string, iterasyon *int) {
	cfg, err := kamusmzd.LoadConfig()
	if err != nil {
		return
	}
	if sunucu != nil && *sunucu == "" && cfg.Sunucu != "" {
		*sunucu = cfg.Sunucu
	}
	if *musteriNo == 0 && cfg.MusteriNo != 0 {
		*musteriNo = uint(cfg.MusteriNo)
	}
	if *parola == "" && cfg.Parola != "" {
		*parola = cfg.Parola
	}
	if hashAlg != nil && *hashAlg == "sha256" && cfg.Hash != "" {
		*hashAlg = cfg.Hash
	}
	if iterasyon != nil && *iterasyon == 100 && cfg.Iterasyon != 0 {
		*iterasyon = cfg.Iterasyon
	}
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func fileNameWithoutExt(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func isPrintableString(s string) bool {
	for _, b := range []byte(s) {
		if b < 0x20 || b > 0x7E {
			if b != '\n' && b != '\r' && b != '\t' {
				return false
			}
		}
	}
	return true
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Hata: "+format+"\n", args...)
	os.Exit(1)
}
