---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.44.1**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.44.1_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.1/stuffstash_v0.44.1_linux_amd64.tar.gz
printf '%s  %s\n' '1128f631893ebbdebde158ea804f084ceea468f48100dee68bd3f523a0f0dedf' 'stuffstash_v0.44.1_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.1_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.44.1_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.1/stuffstash_v0.44.1_linux_arm64.tar.gz
printf '%s  %s\n' 'a620723d8613af17a9260e6178a6001492d4c0f5f4d5af4c066f4c22d8e75508' 'stuffstash_v0.44.1_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.1_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.44.1_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.1/stuffstash_v0.44.1_darwin_amd64.tar.gz
printf '%s  %s\n' '8774d9f3d3e33e19634b722bf32793591f513e3245f9b32f68ea06d9becb4e53' 'stuffstash_v0.44.1_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.1_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.44.1_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.1/stuffstash_v0.44.1_darwin_arm64.tar.gz
printf '%s  %s\n' 'd905d5cc6ede1ad51cfccdc2751d8e042e64dd175958be24b8caadd3c81a57e5' 'stuffstash_v0.44.1_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.1_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.44.1_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.44.1/stuffstash_v0.44.1_windows_amd64.zip
(Get-FileHash stuffstash_v0.44.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `a0a570c0a5e79557a0bbd65986416585032d0bc7bfde4cc32c3c40a3f3ec9e66`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.44.1). Source commit: `99a995fb9b59895a96c8d9db4dd176f992dda51c`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
