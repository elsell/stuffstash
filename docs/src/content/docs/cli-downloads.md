---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.41.1**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.41.1_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.1/stuffstash_v0.41.1_linux_amd64.tar.gz
printf '%s  %s\n' 'a1931bc953ddb21f62fce7e8153b3a7291f33a1562e8007c178b0dde868176db' 'stuffstash_v0.41.1_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.1_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.41.1_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.1/stuffstash_v0.41.1_linux_arm64.tar.gz
printf '%s  %s\n' '72ad5050559a716cc62a17d23899a8c8539e104feb7c6b8c93e2463b11338ea5' 'stuffstash_v0.41.1_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.1_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.41.1_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.1/stuffstash_v0.41.1_darwin_amd64.tar.gz
printf '%s  %s\n' 'bf6477999999bf58d12eca99d4141f42373390eb5bc79fe521c14ab9ee1a8656' 'stuffstash_v0.41.1_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.1_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.41.1_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.1/stuffstash_v0.41.1_darwin_arm64.tar.gz
printf '%s  %s\n' '9913935c1876b2f567667b297a068bebeb83fecf1700ab8a08650c11c3405676' 'stuffstash_v0.41.1_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.1_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.41.1_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.41.1/stuffstash_v0.41.1_windows_amd64.zip
(Get-FileHash stuffstash_v0.41.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `ab62ec23634020b9c44d5c50f88e1481fba9c9bdbc1de190ff989ea51e6ac88f`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.41.1). Source commit: `6d09cbce8680fb238f29c7a3d71ed10015c19769`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
