---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.38.1**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.38.1_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.38.1/stuffstash_v0.38.1_linux_amd64.tar.gz
printf '%s  %s\n' '50ceeac7bd076d471dc5ad74eaa2e1829ec6d9fcaeeb7712b474f977e85e5328' 'stuffstash_v0.38.1_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.38.1_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.38.1_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.38.1/stuffstash_v0.38.1_linux_arm64.tar.gz
printf '%s  %s\n' 'dfc51538b5681f32fdc17062c55a5a295e0d9b6b019312f87244adce597ce541' 'stuffstash_v0.38.1_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.38.1_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.38.1_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.38.1/stuffstash_v0.38.1_darwin_amd64.tar.gz
printf '%s  %s\n' '3bb4f0a7161d17bab9d54fc19aa29bd3039fe507bf9c4a1fd356662d7183c67c' 'stuffstash_v0.38.1_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.38.1_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.38.1_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.38.1/stuffstash_v0.38.1_darwin_arm64.tar.gz
printf '%s  %s\n' '11afb7a4a8f41fe5dc8390206dc8dd043fa5e234c9abf1db50f56b5f7ee53f7b' 'stuffstash_v0.38.1_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.38.1_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.38.1_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.38.1/stuffstash_v0.38.1_windows_amd64.zip
(Get-FileHash stuffstash_v0.38.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `798147c23564fe5a1dd283f11a32b3a2a7f52b97b6172f7d0fdb1799838c1a12`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.38.1). Source commit: `0e52e16a268b8721e64f6298cd6934734b075d24`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
