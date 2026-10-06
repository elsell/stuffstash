---
title: Download the CLI
description: Versioned Stuff Stash command-line downloads.
---

Version **v0.43.1**. Download the archive for your computer and verify its checksum before extracting it.

USB printer support initially targets Linux with a Brother QL-800 and 29 × 90 mm labels. Other builds support ordinary CLI commands.

## Linux amd64

```sh
curl --fail --location --output stuffstash_v0.43.1_linux_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.1/stuffstash_v0.43.1_linux_amd64.tar.gz
printf '%s  %s\n' '889ed80d8ff2698e8383ee70d697d9c101a52133d4d36c18d154ef6d71e96b61' 'stuffstash_v0.43.1_linux_amd64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.1_linux_amd64.tar.gz
./stuffstash version
```

## Linux arm64

```sh
curl --fail --location --output stuffstash_v0.43.1_linux_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.1/stuffstash_v0.43.1_linux_arm64.tar.gz
printf '%s  %s\n' '93c49f412116cc049ecde680f979f4190cebf419068332c58b10b3139186ae29' 'stuffstash_v0.43.1_linux_arm64.tar.gz' | sha256sum --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.1_linux_arm64.tar.gz
./stuffstash version
```

## macOS amd64

```sh
curl --fail --location --output stuffstash_v0.43.1_darwin_amd64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.1/stuffstash_v0.43.1_darwin_amd64.tar.gz
printf '%s  %s\n' 'ac066620000f6f2ed7662cb0944c8375c0a33cc3cefa14795be4483101ab8e5b' 'stuffstash_v0.43.1_darwin_amd64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.1_darwin_amd64.tar.gz
./stuffstash version
```

## macOS arm64

```sh
curl --fail --location --output stuffstash_v0.43.1_darwin_arm64.tar.gz https://github.com/elsell/stuffstash/releases/download/v0.43.1/stuffstash_v0.43.1_darwin_arm64.tar.gz
printf '%s  %s\n' '5930df8f94a533e63d92f2de29174c377ea0588ea0ddc29a74cea46bf636c70d' 'stuffstash_v0.43.1_darwin_arm64.tar.gz' | shasum -a 256 --check
```

After verification succeeds:

```sh
tar -xzf stuffstash_v0.43.1_darwin_arm64.tar.gz
./stuffstash version
```

## Windows amd64

```powershell
curl.exe --fail --location --output stuffstash_v0.43.1_windows_amd64.zip https://github.com/elsell/stuffstash/releases/download/v0.43.1/stuffstash_v0.43.1_windows_amd64.zip
(Get-FileHash stuffstash_v0.43.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
```

Expected SHA-256: `33d48a373c0ac862143311079841a1891f8562a765cbcd1f67e39a3e7654fd52`. Extract only after it matches.

[Release notes and all assets](https://github.com/elsell/stuffstash/releases/tag/v0.43.1). Source commit: `5051dc22a2e74f6367f25be1077f083ed74e6d20`.

Continue with [sign-in and inventory commands](../cli/#sign-in).
