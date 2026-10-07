---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.4**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.4_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.4/stuffstash_v0.43.4_linux_amd64.tar.gz
printf '%s  %s\n' '4cea3f2217fedd636913073d23de2b3b70ba629091122f0c4756553718f7a667' 'stuffstash_v0.43.4_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.4_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.4_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.4/stuffstash_v0.43.4_linux_arm64.tar.gz
printf '%s  %s\n' 'da84929752b9958ce6fde3df1827b0f1f625754d9dc60cb91d2f12c9a1740164' 'stuffstash_v0.43.4_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.4_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.4_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.4/stuffstash_v0.43.4_darwin_amd64.tar.gz
printf '%s  %s\n' 'ed60020508741817a6b687986d52dd971fab5a544627313045f72a0308ada997' 'stuffstash_v0.43.4_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.4_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.4_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.4/stuffstash_v0.43.4_darwin_arm64.tar.gz
printf '%s  %s\n' 'f74dadfc6bb1e7d8308a4f60f1f4297996923a5830d80767440403c39cab4297' 'stuffstash_v0.43.4_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.4_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.4_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.4/stuffstash_v0.43.4_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.4_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `071b23a1c5b77bff36d4379b278b0d76b20b0890e5f2e4b9a551bed14f729bbb`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.4). Source commit: `679fc9cac21d0c10745dc6013429511625e20c42`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
