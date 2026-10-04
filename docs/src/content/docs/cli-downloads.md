---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.42.2**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.42.2_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.2/stuffstash_v0.42.2_linux_amd64.tar.gz
printf '%s  %s\n' '05c75f5c0c615a4355da3d94a014db662010397eff4b3cb57b433384d34b7332' 'stuffstash_v0.42.2_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.2_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.42.2_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.2/stuffstash_v0.42.2_linux_arm64.tar.gz
printf '%s  %s\n' 'e0d1cfaec3ef97cd75786f3c2453bc544da06614cedaf23c9eaba87305252a9d' 'stuffstash_v0.42.2_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.2_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.42.2_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.2/stuffstash_v0.42.2_darwin_amd64.tar.gz
printf '%s  %s\n' 'ef5fa40d307490463ec4dd60866e592bc2238cbcfc6c240ae76b25f68dc958a2' 'stuffstash_v0.42.2_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.2_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.42.2_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.2/stuffstash_v0.42.2_darwin_arm64.tar.gz
printf '%s  %s\n' 'db22973c8ff93c038643229f53d9b43b2e9c4972cd2e9eda10b292ca5ad793df' 'stuffstash_v0.42.2_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.2_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.42.2_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.42.2/stuffstash_v0.42.2_windows_amd64.zip
(Get-FileHash stuffstash_v0.42.2_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `f813aaddae050fdce4b9e272cd13403bf51c94d36f89c1d9a5042160673e9897`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.42.2). Source commit: `b05f5bc7d6d03811d71e0ac2e787e5de9b4ac7a1`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
