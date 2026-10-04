---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.42.0**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.42.0_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.0/stuffstash_v0.42.0_linux_amd64.tar.gz
printf '%s  %s\n' 'd7bdc1d72e9b4df66467b2328be4f1cd3428ccd40a6678b51af95739b419e0ec' 'stuffstash_v0.42.0_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.0_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.42.0_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.0/stuffstash_v0.42.0_linux_arm64.tar.gz
printf '%s  %s\n' '670ea7d8712553b99071c4940d16291c45ff14c52a128135f04a143027f4b9cb' 'stuffstash_v0.42.0_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.0_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.42.0_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.0/stuffstash_v0.42.0_darwin_amd64.tar.gz
printf '%s  %s\n' '7f65fe8071c7333fd05e41ef9111374eb266ef2d25db37b7a95c6b9d24a18298' 'stuffstash_v0.42.0_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.0_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.42.0_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.42.0/stuffstash_v0.42.0_darwin_arm64.tar.gz
printf '%s  %s\n' 'fe7aba3c0f5d015019bf0f486701589b861a399ac042674974e9a855d8531c55' 'stuffstash_v0.42.0_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.42.0_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.42.0_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.42.0/stuffstash_v0.42.0_windows_amd64.zip
(Get-FileHash stuffstash_v0.42.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `e3851f920eef0a95bc274db5708868fb1072a709929b2bd4e8c25ced18287dc8`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.42.0). Source commit: `cba6a7cee7b7b0f68d440cb28f7263d4f7fa5f27`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
