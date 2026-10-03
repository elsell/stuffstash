---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.40.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.40.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.40.0/stuffstash_v0.40.0_linux_amd64.tar.gz
printf '%s  %s\n' 'e1db4089a382a05fca0fde0ae3873540d6cd10c249b8a05536ee2c1495752bb6' 'stuffstash_v0.40.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.40.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.40.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.40.0/stuffstash_v0.40.0_linux_arm64.tar.gz
printf '%s  %s\n' 'a0e53f74c55d2b9d464922422e13c9ccad23e4fa0bf13119e41a278d10a4966e' 'stuffstash_v0.40.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.40.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.40.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.40.0/stuffstash_v0.40.0_darwin_amd64.tar.gz
printf '%s  %s\n' '38e99d51de9cf373ba263a68a4b60889dd99160f8fcd7610695af7e6efbc54f9' 'stuffstash_v0.40.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.40.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.40.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.40.0/stuffstash_v0.40.0_darwin_arm64.tar.gz
printf '%s  %s\n' '563374ede991b1e615820a6d4e3f7fcb7dec489571087deac5e018ac425f5609' 'stuffstash_v0.40.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.40.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.40.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.40.0/stuffstash_v0.40.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.40.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `e7d22d27c292ef773a8e66558f5a90261c3d517257fce7bb077f64685c241942`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.40.0). Source commit: `4afbbc7fad2fbeb1c52bb640572da2dc4c099b64`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
