import { pathToFileURL } from 'node:url';
import { readReleaseMessages, renderPlayNotes } from './release-notes.mjs';
import { GooglePlayClient, serviceAccountToken } from './google-play-client.mjs';

function selectRelease(track, target) {
  if (track?.track !== target.track || !Array.isArray(track.releases)) throw new Error('Play track mismatch');
  for (const release of track.releases) {
    if (!release || !Array.isArray(release.versionCodes) ||
        !release.versionCodes.every(code => typeof code === 'string' && /^[1-9]\d*$/.test(code)) ||
        (release.releaseNotes !== undefined && (!Array.isArray(release.releaseNotes) ||
          !release.releaseNotes.every(note => note && typeof note.language === 'string' && typeof note.text === 'string')))) {
      throw new Error('Invalid Play release');
    }
  }
  const releases = track.releases.filter(release => release.versionCodes?.includes(target.versionCode));
  if (releases.length !== 1 || releases[0].versionCodes.length !== 1) throw new Error('Expected one exact Play release');
  const release = releases[0];
  if (!['draft', 'inProgress', 'halted', 'completed'].includes(release.status) ||
      (release.releaseNotes !== undefined && !Array.isArray(release.releaseNotes))) throw new Error('Invalid Play release');
  const locales = (release.releaseNotes ?? []).filter(note => note.language === target.language);
  if (locales.length > 1) throw new Error('Ambiguous Play locale');
  return { release, current: locales[0]?.text };
}

export async function publishPlayNotes(client, target) {
  const { packageName, track, versionCode, language, notes } = target;
  if (!/^[a-zA-Z][\w]*(?:\.[a-zA-Z][\w]*)+$/.test(packageName ?? '') ||
      !/^[a-zA-Z0-9][a-zA-Z0-9:_-]{0,99}$/.test(track ?? '') ||
      !/^[1-9]\d{0,9}$/.test(versionCode ?? '') || Number(versionCode) > 2100000000 ||
      !/^[a-z]{2,3}(?:-[a-zA-Z0-9]{2,8})*$/.test(language ?? '') ||
      typeof notes !== 'string' || !notes.trim() || [...notes].length > 500) throw new Error('Invalid Play notes target');
  const base = '/applications/' + encodeURIComponent(packageName) + '/edits';
  const createEdit = async () => {
    const edit = await client.request(base, { method: 'POST', body: {} });
    if (!/^[a-zA-Z0-9_-]+$/.test(edit?.id ?? '')) throw new Error('Invalid Play edit');
    return base + '/' + edit.id;
  };
  const trackPath = edit => edit + '/tracks/' + encodeURIComponent(track);
  let edit;
  try {
    edit = await createEdit();
    const original = await client.request(trackPath(edit));
    const { release, current } = selectRelease(original, target);
    if (current === notes) return;
    const otherLocales = (release.releaseNotes ?? []).filter(note => note.language !== language);
    release.releaseNotes = [...otherLocales, { language, text: notes }];
    await client.request(trackPath(edit), { method: 'PUT', body: original });
    if (selectRelease(await client.request(trackPath(edit)), target).current !== notes) throw new Error('Play edit notes verification failed');
    await client.request(edit + ':commit', { method: 'POST' });
    edit = undefined;
    edit = await createEdit();
    if (selectRelease(await client.request(trackPath(edit)), target).current !== notes) throw new Error('Play committed notes verification failed');
  } finally {
    // Cleanup cannot mask the original failure; uncommitted edits also expire.
    if (edit) await client.request(edit, { method: 'DELETE' }).catch(() => {});
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    const notes = renderPlayNotes(readReleaseMessages(process.env.STUFF_STASH_MOBILE_RELEASE_TAG));
    if (process.argv.includes('--preview')) process.stdout.write(notes + '\n');
    else {
      const token = await serviceAccountToken(JSON.parse(process.env.GOOGLE_PLAY_SERVICE_ACCOUNT_JSON ?? ''));
      await publishPlayNotes(new GooglePlayClient(token), {
        packageName: process.env.PLAY_PACKAGE_NAME, track: process.env.PLAY_TRACK,
        versionCode: process.env.PLAY_VERSION_CODE, language: process.env.PLAY_LANGUAGE || 'en-US', notes,
      });
      process.stdout.write('Verified Google Play release notes.\n');
    }
  } catch {
    process.stderr.write('Google Play release notes failed; check the release target, permissions and store state.\n');
    process.exitCode = 1;
  }
}
