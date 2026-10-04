---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.42.1**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.42.1_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.1/stuffstash_v0.42.1_linux_amd64.tar.gz
printf '%s  %s\n' '4f7b95bc9b5351d9146af3906c3324178bd808ff27b616f73ed696bdd1dca608' 'stuffstash_v0.42.1_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.1_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.42.1_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.1/stuffstash_v0.42.1_linux_arm64.tar.gz
printf '%s  %s\n' 'bab12c4199eb7f6cbfd78c96aa6a2dfdcd56e168add5cbffdcbe7b649c017227' 'stuffstash_v0.42.1_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.1_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.42.1_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.1/stuffstash_v0.42.1_darwin_amd64.tar.gz
printf '%s  %s\n' '3ed19ce3f90d9ee6996a65551761a8ca87e8a573500d204c5821a397031d55cd' 'stuffstash_v0.42.1_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.1_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.42.1_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.1/stuffstash_v0.42.1_darwin_arm64.tar.gz
printf '%s  %s\n' '37e0d25bd52757c31b81117030039334a1b7c6e95fe7c568d6bbbaa88ed796ca' 'stuffstash_v0.42.1_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.1_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.42.1_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.42.1/stuffstash_v0.42.1_windows_amd64.zip
(Get-FileHash stuffstash_v0.42.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `1a7fa17257967e189c3f7aa649c200782c6e1209923b7ea401e97d0686154e5c`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.42.1). Source commit: `ff03ab2c0924bdf18b4922b171cb220af94eac10`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
