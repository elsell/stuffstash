---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.2**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.2_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.2/stuffstash_v0.43.2_linux_amd64.tar.gz
printf '%s  %s\n' 'da9a6f1721fcd9d32196b8c6dfacef5d992295d3032a52ebc3ee8a10ce42e444' 'stuffstash_v0.43.2_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.2_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.2_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.2/stuffstash_v0.43.2_linux_arm64.tar.gz
printf '%s  %s\n' 'bfabadc202e529d59815ec8cab1f87a922c33a28cb276579b99610b15b591e74' 'stuffstash_v0.43.2_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.2_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.2_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.2/stuffstash_v0.43.2_darwin_amd64.tar.gz
printf '%s  %s\n' 'c0997d8d2a24fc4b4e5fd14b4464453e2b771d7c9f996300ebb489389ca56927' 'stuffstash_v0.43.2_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.2_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.2_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.2/stuffstash_v0.43.2_darwin_arm64.tar.gz
printf '%s  %s\n' '2c19d215a0a707585ed15787c94645aecf77f51346a0afb022191eaba638c01a' 'stuffstash_v0.43.2_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.2_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.2_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.2/stuffstash_v0.43.2_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.2_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `6432c54b281b6bef03467e85c56f77dfa04d6bac9e4f463bc2c97b3b487a72c0`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.2). Source commit: `2a9e35de59a6a991c140441ad1c4a4910d009bfc`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
