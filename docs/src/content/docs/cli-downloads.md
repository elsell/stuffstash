---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.6**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.6_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.6/stuffstash_v0.43.6_linux_amd64.tar.gz
printf '%s  %s\n' '9861ee6a0f101e74857bb9e4197c71b9ce47234b190b86af9f0744206ae6e706' 'stuffstash_v0.43.6_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.6_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.6_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.6/stuffstash_v0.43.6_linux_arm64.tar.gz
printf '%s  %s\n' 'b82d9442ffa05b5cef9aea4354df34a0e027e8431cc87fef8a3620df2f6244b0' 'stuffstash_v0.43.6_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.6_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.6_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.6/stuffstash_v0.43.6_darwin_amd64.tar.gz
printf '%s  %s\n' 'df1f7babf7eb25dc120a02028aebfde1fad21ae702863cbfc486ef154321c607' 'stuffstash_v0.43.6_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.6_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.6_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.6/stuffstash_v0.43.6_darwin_arm64.tar.gz
printf '%s  %s\n' '7e7bbea8e367d00e0319caea743fed83e42f0db2d9905550a82ddb1218791f3f' 'stuffstash_v0.43.6_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.6_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.6_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.6/stuffstash_v0.43.6_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.6_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `1f378210b22f61ec25b0c73cc6a8f7496cbb14e43f56fce708f92c5ceb57eaed`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.6). Source commit: `4eaeb2f12bd75627ff770b3515aa109ec080b732`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
