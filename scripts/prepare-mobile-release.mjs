#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { createRequire } from 'node:module';

const argumentsByName = parseArguments(process.argv.slice(2));
const repositoryRoot = path.resolve(argumentsByName['repository-root'] ?? process.cwd());
const releaseTag = requiredArgument(argumentsByName, 'tag');
const buildNumber = requiredArgument(argumentsByName, 'build-number');

const require = createRequire(import.meta.url);
const { resolveMobileReleaseVersion } = require('../apps/mobile/scripts/release-version.js');
const version = resolveMobileReleaseVersion({ tag: releaseTag, production: true });

if (!/^[1-9][0-9]{0,3}\.[1-9][0-9]?$/.test(buildNumber)) {
  fail('--build-number must contain a 1-4 digit run number and a 1-2 digit attempt.');
}

const projectPath = path.join(repositoryRoot, 'apps/mobile/ios/StuffStash.xcodeproj/project.pbxproj');
const project = fs.readFileSync(projectPath, 'utf8');
const marketingVersionPattern = /MARKETING_VERSION = [^;\n]+;/g;
const occurrences = project.match(marketingVersionPattern)?.length ?? 0;
if (occurrences !== 2) {
  fail(`expected exactly two MARKETING_VERSION settings in ${projectPath}, found ${occurrences}.`);
}
const buildVersionPattern = /CURRENT_PROJECT_VERSION = [^;\n]+;/g;
const buildOccurrences = project.match(buildVersionPattern)?.length ?? 0;
if (buildOccurrences !== 2) {
  fail(`expected exactly two CURRENT_PROJECT_VERSION settings in ${projectPath}, found ${buildOccurrences}.`);
}
const updatedProject = project
  .replace(marketingVersionPattern, `MARKETING_VERSION = ${version};`)
  .replace(buildVersionPattern, `CURRENT_PROJECT_VERSION = ${buildNumber};`);

fs.writeFileSync(
  projectPath,
  updatedProject
);

process.stdout.write(`Prepared Stuff Stash iOS ${version} (${buildNumber}) from ${releaseTag}.\n`);

function parseArguments(values) {
  const result = {};
  for (let index = 0; index < values.length; index += 2) {
    const name = values[index];
    const value = values[index + 1];
    if (!name?.startsWith('--') || value === undefined || value.startsWith('--')) {
      fail('usage: prepare-mobile-release.mjs --tag vX.Y.Z --build-number NUMBER [--repository-root PATH]');
    }
    result[name.slice(2)] = value;
  }
  return result;
}

function requiredArgument(values, name) {
  const value = values[name]?.trim();
  if (!value) fail(`--${name} is required.`);
  return value;
}

function fail(message) {
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
