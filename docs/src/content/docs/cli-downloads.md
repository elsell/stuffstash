---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.45.2**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.45.2_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.2/stuffstash_v0.45.2_linux_amd64.tar.gz
printf '%s  %s\n' '2ae4826c7409f0c9efd53c3604604a09dcdcb299edf68e56b4952e644a1f367b' 'stuffstash_v0.45.2_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.2_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.45.2_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.2/stuffstash_v0.45.2_linux_arm64.tar.gz
printf '%s  %s\n' 'cd7087dffd6ffe589ed6004213dba0a5070a0c20e01f860020af814b225d06ee' 'stuffstash_v0.45.2_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.2_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.45.2_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.2/stuffstash_v0.45.2_darwin_amd64.tar.gz
printf '%s  %s\n' 'bb141fbc2940f42a3c1e58a1be9a9363407ae6fb4c795e3a90583b14ad5beaaf' 'stuffstash_v0.45.2_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.2_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.45.2_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.45.2/stuffstash_v0.45.2_darwin_arm64.tar.gz
printf '%s  %s\n' 'c5040bf59c725da19a52a8c3bbf5ae1d4cb00b36f877ac47c523b2ea967e1c1d' 'stuffstash_v0.45.2_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.45.2_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.45.2_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.45.2/stuffstash_v0.45.2_windows_amd64.zip
(Get-FileHash stuffstash_v0.45.2_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `fb2c1a52e6766841a2b29ec9499de281298f0f01608feea64884d5584ace60fc`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.45.2). Source commit: `cc22877a823e8bbf57fb365e1ad1dd50ea0d10c7`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
