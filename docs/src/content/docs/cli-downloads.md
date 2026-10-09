---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.45.1**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.45.1_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.1/stuffstash_v0.45.1_linux_amd64.tar.gz
printf '%s  %s\n' '514693442de25c9dd5d43b209e9be190c223f0e8d3acd9a088fb869c0312e43e' 'stuffstash_v0.45.1_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.1_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.45.1_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.1/stuffstash_v0.45.1_linux_arm64.tar.gz
printf '%s  %s\n' 'e5e8fac448e70bf3a81ee41d3aeed3ebc1ba125dd124f169c4526f4f5a3e7654' 'stuffstash_v0.45.1_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.1_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.45.1_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.1/stuffstash_v0.45.1_darwin_amd64.tar.gz
printf '%s  %s\n' '737266b677ac13be767e5239ee4f0ef85f570ae2e5c482743dfe92e8860a344e' 'stuffstash_v0.45.1_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.1_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.45.1_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.1/stuffstash_v0.45.1_darwin_arm64.tar.gz
printf '%s  %s\n' '95d0482be2d47b732b5e4a804dbe305a24575cea9e1b575ec4105edb7bdb9977' 'stuffstash_v0.45.1_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.1_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.45.1_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.45.1/stuffstash_v0.45.1_windows_amd64.zip
(Get-FileHash stuffstash_v0.45.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `60c74fff67133a7376138932ba774c967ce625e69d090842ba9b53cca0d46785`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.45.1). Source commit: `c0bc978d46a3c6e8af5fd467693d611089762fe2`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
