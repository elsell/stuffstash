---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.42.4**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.42.4_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.4/stuffstash_v0.42.4_linux_amd64.tar.gz
printf '%s  %s\n' 'c133673ef12a97b9f6a8e5ff765cfc6101b744d190893e95d5208b99f974ad75' 'stuffstash_v0.42.4_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.4_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.42.4_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.4/stuffstash_v0.42.4_linux_arm64.tar.gz
printf '%s  %s\n' '28fc12d0376d07e574257eb97b9a31a6f7fe4797778d2d3113174348fa0f3b6e' 'stuffstash_v0.42.4_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.4_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.42.4_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.4/stuffstash_v0.42.4_darwin_amd64.tar.gz
printf '%s  %s\n' '9dfd145ffdc19e78d3fda542163fd4ae66221f618d450c771e409a1de3e0bdad' 'stuffstash_v0.42.4_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.4_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.42.4_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.4/stuffstash_v0.42.4_darwin_arm64.tar.gz
printf '%s  %s\n' 'c6aacd3f428a276986d2899c23ed7659bcdca53f6d0e478104c93d9578e8d878' 'stuffstash_v0.42.4_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.4_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.42.4_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.42.4/stuffstash_v0.42.4_windows_amd64.zip
(Get-FileHash stuffstash_v0.42.4_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `d005dd1e08d9cbc998e6b1c2a763a59777d4daff208f2112426f4bfcd85bf8b2`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.42.4). Source commit: `a8183b048ddbc09d72b0e629de7b783fcff58727`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
