---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.42.3**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.42.3_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.3/stuffstash_v0.42.3_linux_amd64.tar.gz
printf '%s  %s\n' 'e8d980df4bfcfcf0ccea94a8307a71b23714f9a966758d3e53f91a0eec588753' 'stuffstash_v0.42.3_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.3_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.42.3_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.3/stuffstash_v0.42.3_linux_arm64.tar.gz
printf '%s  %s\n' 'af891404e6b9d6b8994890277831fea11e6e8cf8a40a245d03d940d4abb2baea' 'stuffstash_v0.42.3_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.3_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.42.3_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.3/stuffstash_v0.42.3_darwin_amd64.tar.gz
printf '%s  %s\n' '1d58adf4a14cf40d18c032a2374f8122f844892574633823e09c46133a301f8e' 'stuffstash_v0.42.3_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.3_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.42.3_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.3/stuffstash_v0.42.3_darwin_arm64.tar.gz
printf '%s  %s\n' '00b4851f2e0dff3239401d3c521c80ea2bd4096e930fdf1b74e7a19c578b463e' 'stuffstash_v0.42.3_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.3_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.42.3_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.42.3/stuffstash_v0.42.3_windows_amd64.zip
(Get-FileHash stuffstash_v0.42.3_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `e4ee37137cef8c915a9cc106a4ebca4167fedd980dff513ef3d5ca142321cba6`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.42.3). Source commit: `7ef56c132c8a6e0297dd4f8f9134f0bf9e1c9354`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
