---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.44.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.44.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.0/stuffstash_v0.44.0_linux_amd64.tar.gz
printf '%s  %s\n' '55356407949f7e719d9698affca19e98d529bd5fa1165a1250e05ae9da9c4bce' 'stuffstash_v0.44.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.44.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.0/stuffstash_v0.44.0_linux_arm64.tar.gz
printf '%s  %s\n' 'dfe507742a6866f03e309f712543934a30e080eea098af9a9617c4b8e4a77266' 'stuffstash_v0.44.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.44.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.0/stuffstash_v0.44.0_darwin_amd64.tar.gz
printf '%s  %s\n' '1170247b4a8153d795aa7c3c6c59ba595a93b6b67883d22af9e4e02becee47e9' 'stuffstash_v0.44.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.44.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.0/stuffstash_v0.44.0_darwin_arm64.tar.gz
printf '%s  %s\n' 'c6acffdd6b5e2868b9f39284895c5be22ead03871359838177b955cef22cb07b' 'stuffstash_v0.44.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.44.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.44.0/stuffstash_v0.44.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.44.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `0468a6d716ccc59977a792ec324c6745f4dc6a9bc0e666fc794e86a8e350bac0`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.44.0). Source commit: `25d387adc55bd5e4448d7c49d2feb2e4ec16045e`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
