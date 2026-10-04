---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.41.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.41.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.0/stuffstash_v0.41.0_linux_amd64.tar.gz
printf '%s  %s\n' '426e40829661918f96db4c625db20e2b05fd420859b32b71cc3dd5d3265ae6d1' 'stuffstash_v0.41.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.41.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.0/stuffstash_v0.41.0_linux_arm64.tar.gz
printf '%s  %s\n' '93b38a5579e93d60c1712de77ce6e06d6384341a8f629e32bc4996c14ce2bc64' 'stuffstash_v0.41.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.41.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.0/stuffstash_v0.41.0_darwin_amd64.tar.gz
printf '%s  %s\n' 'e036f6452924953130437f5306e4d076b9fbb376fbfd81e5dc4f4af2d34db885' 'stuffstash_v0.41.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.41.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.0/stuffstash_v0.41.0_darwin_arm64.tar.gz
printf '%s  %s\n' 'c9222ebf39efaf8fd74bd092b2ccdbf95e3fd662693989ec158dd8afe0c9e59f' 'stuffstash_v0.41.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.41.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.41.0/stuffstash_v0.41.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.41.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `ad1c6c9ac8663e0a8e81de4b5d05f52d4ed180819cee931e84632a8ddc58f087`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.41.0). Source commit: `27346f77d59db3714ae6ee45be5201a5182af7bb`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
