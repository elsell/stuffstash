#!/usr/bin/env node
import { readFileSync, writeFileSync } from 'node:fs';

const apiImage = requiredArg('--api-image');
const webImage = requiredArg('--web-image');
const envPath = process.env.STUFF_STASH_SELFHOST_ENV_PATH || '.env.example';

requireDigest('API image', apiImage);
requireDigest('web image', webImage);

let text = readFileSync(envPath, 'utf8');
text = replaceEnv(text, 'STUFF_STASH_API_IMAGE', apiImage);
text = replaceEnv(text, 'STUFF_STASH_WEB_IMAGE', webImage);
writeFileSync(envPath, text);

function requiredArg(name) {
  const index = process.argv.indexOf(name);
  const value = index >= 0 ? process.argv[index + 1] : '';
  if (!value || value.startsWith('--')) {
    throw new Error(`${name} is required`);
  }
  return value;
}

function requireDigest(label, value) {
  if (!/^ghcr\.io\/[^@]+@sha256:[a-f0-9]{64}$/.test(value)) {
    throw new Error(`${label} must be an immutable ghcr.io digest reference`);
  }
}

function replaceEnv(text, key, value) {
  const pattern = new RegExp(`^${key}=.*$`, 'm');
  if (!pattern.test(text)) {
    throw new Error(`${key} is missing from ${envPath}`);
  }
  return text.replace(pattern, `${key}=${value}`);
}
