---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.0/stuffstash_v0.43.0_linux_amd64.tar.gz
printf '%s  %s\n' '1f52940ae24547dca68c4ab5d3e6e2fa0db3f0f19bc5d20c968875913107ae3c' 'stuffstash_v0.43.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.0/stuffstash_v0.43.0_linux_arm64.tar.gz
printf '%s  %s\n' '485862f969af795c62c3e4e5a2eaba1c4c28c53c5b43c99e14fd343226f53e5c' 'stuffstash_v0.43.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.0/stuffstash_v0.43.0_darwin_amd64.tar.gz
printf '%s  %s\n' '645ff94636d036775effc0b6eeaead45aece28eccad570229bf87e94564de995' 'stuffstash_v0.43.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.0/stuffstash_v0.43.0_darwin_arm64.tar.gz
printf '%s  %s\n' '7a7d70a5944972de383b266ee7ede0dceb18d6c480d5332658af955e98a7a22c' 'stuffstash_v0.43.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.0/stuffstash_v0.43.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `d1423e7357c35e42536e75586947473a0a97b9480b7e9a81679ae8376a26b872`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.0). Source commit: `7211bf40ca1791440f92991e3e09c892f998a2e7`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
