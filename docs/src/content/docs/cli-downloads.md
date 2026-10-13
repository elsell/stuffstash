---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.45.3**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.45.3_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.3/stuffstash_v0.45.3_linux_amd64.tar.gz
printf '%s  %s\n' '894ed1157331bcdf503c0cfb45728eea9c51ce795d8c3f025e112baffef299a5' 'stuffstash_v0.45.3_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.3_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.45.3_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.3/stuffstash_v0.45.3_linux_arm64.tar.gz
printf '%s  %s\n' '7e6c3d0e92bfa16c6bea9451b7deb3cd66e0b613a5a6735fb58b4b223029b80b' 'stuffstash_v0.45.3_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.3_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.45.3_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.3/stuffstash_v0.45.3_darwin_amd64.tar.gz
printf '%s  %s\n' '43255c8d1066b512a5901f7b35b140adf8b251d617038b409f5a6d6106533851' 'stuffstash_v0.45.3_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.3_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.45.3_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.3/stuffstash_v0.45.3_darwin_arm64.tar.gz
printf '%s  %s\n' '55ff7e2471c3f3284e6af3a957d891c22de2d366bc8784a6d81fe035ffa033b1' 'stuffstash_v0.45.3_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.3_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.45.3_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.45.3/stuffstash_v0.45.3_windows_amd64.zip
(Get-FileHash stuffstash_v0.45.3_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `b9b373575ad4b1c8d0db70fd50fd37e3e5eced95d8cc1c76c29fb1812bdae17b`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.45.3). Source commit: `9e61cac5dcf5f31db613d22982e30e89f6cd5082`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
