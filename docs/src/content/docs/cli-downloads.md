---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.42.5**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.42.5_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.5/stuffstash_v0.42.5_linux_amd64.tar.gz
printf '%s  %s\n' '076f6939c34611828d757e0ef72229ad6c71afeaadf9addc4e3053ffb5f2a6f9' 'stuffstash_v0.42.5_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.5_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.42.5_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.5/stuffstash_v0.42.5_linux_arm64.tar.gz
printf '%s  %s\n' 'e157249b2e01f68128c3feef30ae37b21da89a3e6a2a8f1cd50bc475c0a721c2' 'stuffstash_v0.42.5_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.5_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.42.5_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.5/stuffstash_v0.42.5_darwin_amd64.tar.gz
printf '%s  %s\n' '91d2e9fdfd87602c21930f2b2590a73a4a58b15e921022349e3655be9af8321c' 'stuffstash_v0.42.5_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.5_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.42.5_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.5/stuffstash_v0.42.5_darwin_arm64.tar.gz
printf '%s  %s\n' 'e0036973c2b3ead518667afa1857463d5bd1e985aae24c7c2d2d1e0439b12ddd' 'stuffstash_v0.42.5_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.5_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.42.5_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.42.5/stuffstash_v0.42.5_windows_amd64.zip
(Get-FileHash stuffstash_v0.42.5_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `84ef817e9c1ed1e24a2ff4b4466e0e6d7d906289f5f7d09e94e85bd388a1eb7c`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.42.5). Source commit: `8c6e3c7b66fd397c8a6062fdc55f56642c5f117f`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
