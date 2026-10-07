---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.5**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.5_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.5/stuffstash_v0.43.5_linux_amd64.tar.gz
printf '%s  %s\n' 'e2eb4e6c2db719bd645e05336a6a67213bb8c4a73164d76e95a82e363d6ccd0c' 'stuffstash_v0.43.5_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.5_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.5_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.5/stuffstash_v0.43.5_linux_arm64.tar.gz
printf '%s  %s\n' '3e11e34ba5ac59fcd2aa377e5bb380373fadae934795895b1467051bb9db62b0' 'stuffstash_v0.43.5_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.5_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.5_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.5/stuffstash_v0.43.5_darwin_amd64.tar.gz
printf '%s  %s\n' 'bc7ccd4516dacd1c8b86b7e581ac4916c26edcbce3235985613cebd3ea009291' 'stuffstash_v0.43.5_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.5_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.5_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.5/stuffstash_v0.43.5_darwin_arm64.tar.gz
printf '%s  %s\n' 'b3f54241ee1adda1e2f3fd8936b223fdc42d374032800089926779dd6ef96801' 'stuffstash_v0.43.5_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.5_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.5_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.5/stuffstash_v0.43.5_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.5_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `54468071cee30dcd45c2c6ae7c6fcb99e460fcdf099a7db752ac616a03ff93b3`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.5). Source commit: `d0e0b990b3886cae3adc80b0dbdf8bfaa2336abf`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
