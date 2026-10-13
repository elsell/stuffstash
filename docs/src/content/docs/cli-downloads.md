---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.45.4**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.45.4_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.4/stuffstash_v0.45.4_linux_amd64.tar.gz
printf '%s  %s\n' 'c0b368e5ffd4d3cc31f47d10dab8a4a0014c12154b45855dcd281eb558c5ae73' 'stuffstash_v0.45.4_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.4_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.45.4_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.4/stuffstash_v0.45.4_linux_arm64.tar.gz
printf '%s  %s\n' '1ee128fd5b8ea7de420438f51a60dbb3ff871033cec863c355ed604fd012890c' 'stuffstash_v0.45.4_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.4_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.45.4_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.4/stuffstash_v0.45.4_darwin_amd64.tar.gz
printf '%s  %s\n' '637fcfe2981d2bf16636ba57cc5254add5c830721750efd4ddabed85b0b1f039' 'stuffstash_v0.45.4_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.4_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.45.4_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.4/stuffstash_v0.45.4_darwin_arm64.tar.gz
printf '%s  %s\n' '78354ecc4f03cee9c009a48a13694dcfe955fd9f8472297ef4603ff7d267183c' 'stuffstash_v0.45.4_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.4_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.45.4_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.45.4/stuffstash_v0.45.4_windows_amd64.zip
(Get-FileHash stuffstash_v0.45.4_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `63b1acf240a9cf4539871a298f410bf8ad088dac874d7c64fb1923554a7ff00e`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.45.4). Source commit: `b96807b947ac6c6e02b6983f16901a16cf8f355d`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
