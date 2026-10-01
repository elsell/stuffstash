import { execFileSync } from 'node:child_process';

export function releaseHighlights(messages) {
  return [...new Set(messages.flatMap(message => {
    const lines = message.trim().split(/\r?\n/);
    const subject = /^(?:feat|fix|perf)(?:\([^)]*\))?!?:\s*(.+)$/.exec(lines[0]);
    if (!subject) return [];
    const shared = lines.indexOf('Release notes:');
    const marker = shared > 0 ? shared : lines.indexOf('TestFlight notes:');
    const highlights = [];
    if (marker > 0) {
      for (const line of lines.slice(marker + 1)) {
        if (!line.startsWith('- ') || !line.slice(2).trim()) break;
        highlights.push(line.slice(2).trim());
      }
    }
    return highlights.length ? highlights : [subject[1]];
  }))];
}

export function renderPlayNotes(messages) {
  const highlights = releaseHighlights(messages);
  const body = highlights.length ? highlights.map(line => '- ' + line).join('\n') : 'Maintenance and reliability updates.';
  const points = [...body];
  return points.length <= 500 ? body : points.slice(0, 499).join('') + '…';
}

export function renderTestFlightNotes(tag, buildNumber, messages, repositoryURL = null) {
  const changes = releaseHighlights(messages);
  const heading = `Stuff Stash ${tag.slice(1)} (${buildNumber})\n\n`;
  const footer = repositoryURL ? `\n\nFull changelog: ${repositoryURL}/releases/tag/${tag}` : '';
  const body = changes.length ? changes.map(change => '- ' + change).join('\n') : 'Maintenance and reliability updates. See the full changelog for details.';
  const available = 4000 - heading.length - footer.length;
  return heading + (body.length <= available ? body : body.slice(0, available - 1) + '…') + footer;
}

export function readReleaseMessages(tag, cwd = process.cwd()) {
  const stable = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
  if (!stable.test(tag ?? '')) throw new Error('Invalid release tag');
  const env = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.startsWith('GIT_')));
  const git = (...args) => execFileSync('git', args, { cwd, env, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
  const ref = 'refs/tags/' + tag;
  git('merge-base', '--is-ancestor', ref, 'HEAD');
  const version = tag.slice(1).split('.').map(BigInt);
  const earlier = value => {
    const parts = value.slice(1).split('.').map(BigInt);
    for (let index = 0; index < 3; index++) {
      if (parts[index] !== version[index]) return parts[index] < version[index];
    }
    return false;
  };
  const previous = git('tag', '--merged', ref, '--list', 'v[0-9]*.[0-9]*.[0-9]*', '--sort=-v:refname')
    .split('\n').find(value => stable.test(value) && earlier(value));
  return git('log', '--first-parent', '--format=%B%x00', previous ? 'refs/tags/' + previous + '..' + ref : ref).split('\0');
}
