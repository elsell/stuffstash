---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.45.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.45.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.0/stuffstash_v0.45.0_linux_amd64.tar.gz
printf '%s  %s\n' 'b9d274a54347ba6c2b44996a2d3ac4a7e95504ce5e590ed26c1fe1a48a607cc7' 'stuffstash_v0.45.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.45.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.0/stuffstash_v0.45.0_linux_arm64.tar.gz
printf '%s  %s\n' '24c5eb27c3b26c1e82f2c2239b362d5a63f4059340e2a5ef73eeb5cf7879a89f' 'stuffstash_v0.45.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.45.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.0/stuffstash_v0.45.0_darwin_amd64.tar.gz
printf '%s  %s\n' 'dd30541e987538fe44b396564af9dbd6178b2cf91e51962deba63bb2b061d232' 'stuffstash_v0.45.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.45.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.0/stuffstash_v0.45.0_darwin_arm64.tar.gz
printf '%s  %s\n' '75ed01fecca185d3df3471fd90260758e6fde11e388695e0c5f42cf925e19e2f' 'stuffstash_v0.45.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.45.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.45.0/stuffstash_v0.45.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.45.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `0e82b8f2eb2cc71924e929eda30cbb681915834690d7d779b60df1e97f732592`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.45.0). Source commit: `7dc05f7275fc2fcd5f09adf440f33c245bd6bddd`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
