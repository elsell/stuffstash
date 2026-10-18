---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.46.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.46.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.46.0/stuffstash_v0.46.0_linux_amd64.tar.gz
printf '%s  %s\n' '70de923b36cd974ccf40e6c39e15dc7b8a5f887447d4625594f82b9c4e3ccb07' 'stuffstash_v0.46.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.46.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.46.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.46.0/stuffstash_v0.46.0_linux_arm64.tar.gz
printf '%s  %s\n' 'adb8734e3e5c69043bf74a14a58ea45474774fa030057a77a7ba007471018270' 'stuffstash_v0.46.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.46.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.46.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.46.0/stuffstash_v0.46.0_darwin_amd64.tar.gz
printf '%s  %s\n' 'a0b35171bee9151d93e2156b2dc0c8fc47ce769f65a3610b7f8c085a5c6894c9' 'stuffstash_v0.46.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.46.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.46.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.46.0/stuffstash_v0.46.0_darwin_arm64.tar.gz
printf '%s  %s\n' '0cac1ec289edef83d54ee73a31704acf08947c3234fd0a36959ebd9a50f743cf' 'stuffstash_v0.46.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.46.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.46.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.46.0/stuffstash_v0.46.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.46.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `918b35059fe415baac43915e23fecca2980d50bd3db457c0e99ef5c32946ebd1`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.46.0). Source commit: `e0c53a6c458b61e83741ceb65bc4900c63d19651`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
