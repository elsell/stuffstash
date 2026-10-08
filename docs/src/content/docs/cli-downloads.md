---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.44.2**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.44.2_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.2/stuffstash_v0.44.2_linux_amd64.tar.gz
printf '%s  %s\n' 'b608003b0cc0f847e1bab9c4716cc7c73976e87b36b7033d7d6d07dd3f390025' 'stuffstash_v0.44.2_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.2_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.44.2_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.2/stuffstash_v0.44.2_linux_arm64.tar.gz
printf '%s  %s\n' '11104edc3cf72d5a2944becd028c69d68abdd688f560ad75298732103efb9909' 'stuffstash_v0.44.2_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.2_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.44.2_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.2/stuffstash_v0.44.2_darwin_amd64.tar.gz
printf '%s  %s\n' 'dbbaf9a2a52a859f625d250021781abfffa7106c42d3d81111703a46e0e76c4c' 'stuffstash_v0.44.2_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.2_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.44.2_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.44.2/stuffstash_v0.44.2_darwin_arm64.tar.gz
printf '%s  %s\n' '0cce02780f76d1c2d7c19bab7fdaa59e6ee070f75bb12eaccc9f47468bc9da66' 'stuffstash_v0.44.2_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.44.2_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.44.2_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.44.2/stuffstash_v0.44.2_windows_amd64.zip
(Get-FileHash stuffstash_v0.44.2_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `45f2b93e4624bf3cf313165649d397065c2f396df5bd61d8b1dfaa11b542d1ee`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.44.2). Source commit: `28cf497f203a1bf9c3859f0c183b19fb0427e767`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
