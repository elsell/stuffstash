---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.3**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.3_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.3/stuffstash_v0.43.3_linux_amd64.tar.gz
printf '%s  %s\n' '4866e8881c4b5770879ac42f05bd648fbe88b0c7f899e5c5a6fce4ee4e46e8ef' 'stuffstash_v0.43.3_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.3_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.3_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.3/stuffstash_v0.43.3_linux_arm64.tar.gz
printf '%s  %s\n' 'c252b84b023b1f82f07347d717813c29f3ecdc9c3a080807629d0849b78f8d1c' 'stuffstash_v0.43.3_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.3_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.3_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.3/stuffstash_v0.43.3_darwin_amd64.tar.gz
printf '%s  %s\n' '5fe7a7a99023312250660df56e283ea484fc6d3c88fa606ba220028dc601890d' 'stuffstash_v0.43.3_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.3_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.3_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.3/stuffstash_v0.43.3_darwin_arm64.tar.gz
printf '%s  %s\n' 'c380f3f1ee6a3b3de3e2f9df7951d6afe1e8d68da49a6d6cf9ff786cfad815d2' 'stuffstash_v0.43.3_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.3_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.3_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.3/stuffstash_v0.43.3_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.3_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `391dcdfbb5101954b45f5e7721f6c5303d82178ab041854031bc40ebe08bb93b`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.3). Source commit: `75a4cc111332a3de1f4b2ad7554fc2920e39f572`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
