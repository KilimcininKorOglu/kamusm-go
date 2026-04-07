# kamusm-go

Kamu SM zaman damgası sunucuları ile iletişim kuran komut satırı aracı. TÜBİTAK BİLGEM tarafından işletilen [Kamu SM](https://kamusm.bilgem.tubitak.gov.tr/) altyapısı üzerinden RFC 3161 uyumlu zaman damgası almak, bakiye sorgulamak, kimlik doğrulama başlığı üretmek ve zaman damgası doğrulamak için kullanılır.

Go standart kütüphanesi üzerine inşa edilmiştir. Harici bağımlılıklar: `golang.org/x/crypto` (PBKDF2) ve `go.mozilla.org/pkcs7` (doğrulama).

## Kurulum

```bash
go install github.com/KilimcininKorOglu/kamusm-go@latest
```

Ya da kaynak koddan:

```bash
git clone https://github.com/KilimcininKorOglu/kamusm-go.git
cd kamusm-go
make build
```

Binary `bin/kamusm-go` altında oluşur. Tüm build işlemleri Makefile üzerinden yapılır:

```bash
make build       # bin/kamusm-go derle
make vet         # Statik analiz
make test        # build + vet
make install     # GOPATH/bin'e kur
make clean       # Build çıktılarını temizle
make run ARGS="" # Derle ve çalıştır
make version     # Mevcut versiyon
make help        # Tüm hedefler
```

## Hızlı Başlangıç

İlk kullanımda bağlantı bilgilerini kaydedin:

```bash
kamusm-go ayar-kaydet \
    --sunucu http://zd.kamusm.gov.tr \
    --musteri-no 123456 \
    --parola "sifre"
```

Sonraki çağrılarda parametre gerekmez:

```bash
kamusm-go gonder --dosya belge.pdf --dogrula
kamusm-go bakiye
```

Başarılı sonuçta `belge_zd.der` dosyası oluşur ve KamuSM kök sertifikalarıyla doğrulanır.

## Komutlar

### gonder

Dosya veya özet değeri için zaman damgası ister.

```bash
# Dosyadan
kamusm-go gonder --dosya DOSYA [--hash sha256] [--dogrula]

# Hex özetten
kamusm-go gonder --ozet-hex HEX [--hash sha256]

# Tüm parametreler açık (config kullanmadan)
kamusm-go gonder --sunucu URL --musteri-no ID --parola PAROLA --dosya DOSYA
```

`--dosya` verildiğinde çıktı girdi dosyasının yanına `{ad}_zd.der` olarak yazılır. `--ozet-hex` verildiğinde çalışma dizinine `zd_{timestamp}.der` olarak yazılır.

`--dogrula` aktifse kaydedilen dosya otomatik olarak KamuSM kök sertifikalarıyla doğrulanır.

### bakiye

Hesaptaki kalan zaman damgası bakiyesini sorgular.

```bash
kamusm-go bakiye
```

Çıktı:

```
Yanıt durumu: 200
Kalan zaman damgası bakiyesi: 847
```

### dogrula

Daha önce alınmış bir zaman damgası dosyasını KamuSM kök sertifikalarına (v4-v7) karşı doğrular.

```bash
kamusm-go dogrula --dosya belge_zd.der
```

Çıktı:

```
Doğrulama başarılı
  İmzalayan: Kamu SM Zaman Damgasi Sunucusu S2
```

PKCS#7 imzası ve sertifika zinciri kontrol edilir. Gömülü kök sertifikalar: Sürüm 4, 5, 6 ve 7.

### kimlik

Sunucuya gönderilecek `identity` HTTP başlığını üretir. Hata ayıklama ve entegrasyon geliştirme amaçlıdır.

```bash
# Hex özetten
kamusm-go kimlik --musteri-no ID --parola PAROLA --ozet-hex HEX

# Zaman damgasından (bakiye sorgusu formatında)
kamusm-go kimlik --musteri-no ID --parola PAROLA --zaman 1712400000000
```

### ayar-kaydet

Bağlantı bilgilerini makine anahtarıyla şifreleyerek `~/.kamusm-go.conf` dosyasına kaydeder. Kaydedilen bilgiler diğer komutlarda parametre verilmediğinde otomatik olarak kullanılır.

```bash
kamusm-go ayar-kaydet \
    --sunucu http://zd.kamusm.gov.tr \
    --musteri-no 123456 \
    --parola "sifre" \
    [--hash sha256] \
    [--iterasyon 100]
```

Şifreleme makineye özeldir (hostname + kullanıcı adı ile türetilen anahtar). Dosya başka bir makinede çözülemez.

### ayar-goster

Kayıtlı ayarları görüntüler (parola maskeli).

```bash
kamusm-go ayar-goster
```

Çıktı:

```
Kayıtlı ayarlar:
  Sunucu:     http://zd.kamusm.gov.tr
  Müşteri No: 123456
  Parola:     sif****
  Hash:       sha256
  İterasyon:  100
```

### versiyon

```bash
kamusm-go versiyon
```

## JSON Çıktı

Tüm komutlar `--json` parametresiyle yapılandırılmış JSON çıktısı verir:

```bash
kamusm-go bakiye --json
```

```json
{
  "durum": 200,
  "bakiye": 847
}
```

```bash
kamusm-go dogrula --dosya belge_zd.der --json
```

```json
{
  "gecerli": true,
  "imzalayan": "Kamu SM Zaman Damgasi Sunucusu S2",
  "tarih": "2026-04-06T20:11:36Z"
}
```

## Parametre Referansı

Tüm komutlarda ortak (config'den de okunabilir):

| Parametre      | Zorunlu | Varsayılan | Açıklama                 |
|----------------|---------|------------|--------------------------|
| `--musteri-no` | Evet*   | -          | Kamu SM müşteri numarası |
| `--parola`     | Evet*   | -          | Müşteri parolası         |
| `--iterasyon`  | Hayır   | 100        | PBKDF2 iterasyon sayısı  |
| `--json`       | Hayır   | false      | JSON formatında çıktı    |

*Config dosyası varsa zorunlu değil.

`gonder` ve `bakiye` komutlarında ek:

| Parametre  | Zorunlu | Açıklama           |
|------------|---------|---------------------|
| `--sunucu` | Evet*   | Sunucu adresi (URL) |

`gonder` komutuna özel:

| Parametre    | Açıklama                                  |
|--------------|-------------------------------------------|
| `--dosya`    | Damgalanacak dosya yolu                   |
| `--ozet-hex` | Önceden hesaplanmış özet (hex)            |
| `--hash`     | `sha1` veya `sha256` (varsayılan: sha256) |
| `--dogrula`  | Kaydedilen dosyayı otomatik doğrula       |

`dogrula` komutuna özel:

| Parametre | Zorunlu | Açıklama                   |
|-----------|---------|----------------------------|
| `--dosya` | Evet    | Doğrulanacak .der dosyası  |

## Yapılandırma Dosyası

`ayar-kaydet` komutuyla kaydedilen bilgiler `~/.kamusm-go.conf` dosyasında şifreli tutulur. Diğer komutlar çalıştırıldığında eksik parametreler bu dosyadan okunur. CLI parametreleri her zaman config değerlerinin üstüne yazar.

Dosya formatı: AES-256-CBC ile şifrelenmiş binary. Anahtar, makine kimliğinden (hostname + kullanıcı adı) PBKDF2 ile türetilir.

## Nasıl Çalışır

### Kimlik Doğrulama

Kamu SM sunucuları standart RFC 3161 protokolünü kullanır ancak ek olarak özel bir `identity` HTTP başlığı gerektirir. Bu başlık şu şekilde üretilir:

1. 16 byte kriptografik rastgele değer üretilir (hem salt hem IV olarak kullanılır)
2. Kullanıcı parolasından PBKDF2-HMAC-SHA256 ile 32 byte AES anahtarı türetilir
3. İstek özetiyle (messageImprint) AES-256-CBC şifreleme yapılır
4. Sonuç ASN.1 DER formatında kodlanır ve hex string'e dönüştürülür

Her istekte yeni salt/IV üretildiği için aynı parametrelerle bile farklı başlık oluşur.

### Zaman Damgası İsteği

İstemci, dosyanın SHA-256 (veya SHA-1) özetini hesaplar, RFC 3161 TimeStampReq yapısını DER formatında kodlar ve HTTP POST ile sunucuya gönderir. Yanıtta PKCS#7 SignedData yapısı aranır; bulunursa `.der` dosyasına kaydedilir.

### PKI Doğrulama

Gömülü KamuSM kök sertifikaları (Sürüm 4, 5, 6, 7) kullanılarak PKCS#7 imzası ve sertifika zinciri doğrulanır. Bu sayede harici araç gerekmeden zaman damgasının bütünlüğü teyit edilir.

### Hata Tespiti

Sunucu hata durumlarında bile HTTP 200 döner. İstemci yanıt gövdesinde PKCS#7 SignedData OID'sini (`1.2.840.113549.1.7.2`) arar. OID bulunamazsa ASN.1 yapısındaki metin alanları taranarak hata mesajı çıkarılır.

## Sık Karşılaşılan Hatalar

| Mesaj                                    | Sebep                                  |
|------------------------------------------|----------------------------------------|
| `Account X could not be authenticated`   | Hatalı parola veya süresi dolmuş hesap |
| `User X is not known`                    | Geçersiz müşteri numarası              |
| Bağlantı hatası                          | Ağ erişimi veya güvenlik duvarı sorunu |

## Gereksinimler

- Go 1.22 veya üzeri
- Make
- Geçerli Kamu SM hesabı
- Sunucuya ağ erişimi (varsayılan port 80)
