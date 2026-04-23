# Changelog

## [1.2.0] - 2026-04-23

### Added
- Importable `kamusm-zd` library package for use in other Go projects
- Library API documentation and usage examples in README
- Single version source in `kamusm-zd/version.go` shared by CLI and library

### Changed
- Extract core logic from flat `package main` into `kamusm-zd/` package with exported API
- Update README for both end-user and developer audiences

## [1.1.1] - 2026-04-08

### Changed
- Improve README with badges, parameter tables, and error examples

### Fixed
- Reject PBKDF2 iteration count below 1 to prevent zero key stretching

## [1.1.0] - 2026-04-07

### Changed
- Document security limitations in README (config encryption, identity protocol, password visibility)

### Fixed
- Use per-file random salt in config key derivation instead of ignoring it
- Add 30-second HTTP client timeout to prevent indefinite hangs

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
