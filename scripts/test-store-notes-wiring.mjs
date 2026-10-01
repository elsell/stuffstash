import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
const read = name => readFileSync(new URL('../.github/workflows/' + name, import.meta.url), 'utf8');
test('existing Stuff Stash TestFlight notes delivery remains connected', () => {
 const upload=read('testflight.yml');
 assert.match(upload,/needs: publish/);
 assert.match(upload,/uses: \.\/\.github\/workflows\/testflight-notes\.yml/);
 assert.match(read('testflight-notes.yml'),/MOBILE_BUNDLE_ID: org\.stuffstash\.mobile/);
});
test('Play uses the Android identity and explicit release inputs', () => {
 const play=read('play-store-notes.yml');
 assert.match(play,/app\.stuffstash\.mobile/);
 assert.match(play,/environment: play-store/);
 assert.match(play,/PLAY_VERSION_CODE: \$\{\{ inputs\.version_code \}\}/);
 assert.match(play,/PLAY_TRACK: \$\{\{ inputs\.track \}\}/);
});
