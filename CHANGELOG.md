# Changelog

## [1.0.0] - 2026-04-06

### Added
- RFC 3161 timestamp client with identity/send/credits subcommands
- PBKDF2-HMAC-SHA256 key derivation and AES-256-CBC encryption
- RFC 3161 TimeStampReq builder with SHA1/SHA256 support
- PKCS#7 SignedData extraction from server responses
- PKI verification with embedded KamuSM root CA certificates (v4-v7)
- Encrypted config file with machine-key encryption (ayar-kaydet/ayar-goster)
- JSON output flag (--json) on all subcommands
- Version command (versiyon/--version/-v)
- GitHub Actions CI and goreleaser release workflows

### Changed
- Subcommands renamed to Turkish (kimlik, gonder, bakiye, dogrula)
- CLI parameters renamed to Turkish (--musteri-no, --parola, --sunucu, etc.)
- User-Agent set to KilimcininKorOglu/kamusm-go/<version>

### Fixed
- Add KamuSM root CA v7 for current production timestamp verification
