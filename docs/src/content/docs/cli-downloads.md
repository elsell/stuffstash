---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.41.2**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.41.2_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.2/stuffstash_v0.41.2_linux_amd64.tar.gz
printf '%s  %s\n' 'd6be2de1bd8468922e7253203b08ca3afa66535ff4d5a70c380745043d17307a' 'stuffstash_v0.41.2_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.2_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.41.2_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.2/stuffstash_v0.41.2_linux_arm64.tar.gz
printf '%s  %s\n' 'a9202c4c51b4bbc0e5e9b9cb236405fd109b9586fa1a53e207cf2e3f63aab500' 'stuffstash_v0.41.2_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.2_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.41.2_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.2/stuffstash_v0.41.2_darwin_amd64.tar.gz
printf '%s  %s\n' '252300006828eba9610dca5f6001a2bc5b1f0184b071bbc9773d477ed0ffd54d' 'stuffstash_v0.41.2_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.2_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.41.2_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.41.2/stuffstash_v0.41.2_darwin_arm64.tar.gz
printf '%s  %s\n' '82a3f59872213b6f4cbcebfa992b1f21c3ed76e2abaec46a722d1674571b89f1' 'stuffstash_v0.41.2_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.41.2_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.41.2_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.41.2/stuffstash_v0.41.2_windows_amd64.zip
(Get-FileHash stuffstash_v0.41.2_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `54c8b1df44fe6f22377cb3a6d3c3d65a08138e874224dc2847bb3ab99702a1b3`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.41.2). Source commit: `6f257c07336933146857aebd79d5b184638a2b27`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
