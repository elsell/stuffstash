import { describe, expect, it } from 'vitest';
import { ExpoVoiceAudioPlayerCore, ExpoVoiceAudioRecorderCore, normalizeDbfsLevel } from './ExpoVoiceAudioCore';

describe('ExpoVoiceAudioRecorder', () => {
  it('records through Expo Audio and returns a base64 mp4 chunk', async () => {
    const recorder = new FakeRecorder('file:///recording.m4a');
    const audio = new FakeAudio(recorder);
    const fileSystem = new FakeFileSystem({ 'file:///recording.m4a': 'YXVkaW8=' });
    const voiceRecorder = new ExpoVoiceAudioRecorderCore(audio, fileSystem);

    await voiceRecorder.start();
    const recorded = await voiceRecorder.stop();

    expect(audio.modes).toEqual([
      { allowsRecording: true, playsInSilentMode: true },
      { allowsRecording: false, playsInSilentMode: true }
    ]);
    expect(recorder.prepared).toBe(true);
    expect(recorder.recording).toBe(false);
    expect(recorded).toEqual({
      mimeType: 'audio/mp4',
      sampleRate: 44100,
      channels: 1,
      chunksBase64: ['YXVkaW8=']
    });
    expect(fileSystem.deleted).toEqual(['file:///recording.m4a']);
  });

  it('splits recorded audio into protocol-sized chunks and still deletes the recorder file', async () => {
    const recorder = new FakeRecorder('file:///large-recording.m4a');
    const fileSystem = new FakeFileSystem({ 'file:///large-recording.m4a': 'A'.repeat(700_000) });
    const voiceRecorder = new ExpoVoiceAudioRecorderCore(new FakeAudio(recorder), fileSystem);

    await voiceRecorder.start();
    const recorded = await voiceRecorder.stop();

    expect(recorded.chunksBase64).toHaveLength(3);
    expect(recorded.chunksBase64[0]?.length).toBeLessThanOrEqual(349_524);
    expect(fileSystem.deleted).toEqual(['file:///large-recording.m4a']);
  });

  it('rejects recording when microphone permission is denied', async () => {
    const audio = new FakeAudio(new FakeRecorder('file:///recording.m4a'));
    audio.granted = false;
    const voiceRecorder = new ExpoVoiceAudioRecorderCore(audio, new FakeFileSystem({}));

    await expect(voiceRecorder.start()).rejects.toThrow('Microphone permission is required');
  });

  it('cancels recording without reading or returning audio', async () => {
    const recorder = new FakeRecorder('file:///recording.m4a');
    const audio = new FakeAudio(recorder);
    const fileSystem = new FakeFileSystem({ 'file:///recording.m4a': 'YXVkaW8=' });
    const voiceRecorder = new ExpoVoiceAudioRecorderCore(audio, fileSystem);

    await voiceRecorder.start();
    await voiceRecorder.cancel();

    expect(recorder.recording).toBe(false);
    expect(audio.modes).toEqual([
      { allowsRecording: true, playsInSilentMode: true },
      { allowsRecording: false, playsInSilentMode: true }
    ]);
    expect(fileSystem.reads).toEqual([]);
    expect(fileSystem.deleted).toEqual(['file:///recording.m4a']);
  });

  it('reports normalized native metering while recording', async () => {
    const recorder = new FakeRecorder('file:///recording.m4a');
    const voiceRecorder = new ExpoVoiceAudioRecorderCore(new FakeAudio(recorder), new FakeFileSystem({}));

    expect(voiceRecorder.recordingLevel()).toBe(0);

    await voiceRecorder.start();
    recorder.metering = -30;

    expect(voiceRecorder.recordingLevel()).toBeCloseTo(0.5);
  });

  it('normalizes dBFS metering to a bounded UI level', () => {
    expect(normalizeDbfsLevel(undefined)).toBe(0);
    expect(normalizeDbfsLevel(-90)).toBe(0);
    expect(normalizeDbfsLevel(-30)).toBeCloseTo(0.5);
    expect(normalizeDbfsLevel(0)).toBe(1);
    expect(normalizeDbfsLevel(12)).toBe(1);
  });
});

describe('ExpoVoiceAudioPlayer', () => {
  it('cleans stale voice cache files before playback and after stop', async () => {
    const audio = new FakeAudio(new FakeRecorder('file:///recording.m4a'));
    const fileSystem = new FakeFileSystem({
      'file:///cache/stuffstash-voice-old.mp3': 'old',
      'file:///cache/unrelated.mp3': 'keep'
    });
    const player = new ExpoVoiceAudioPlayerCore(audio, fileSystem);

    await player.playChunk('c3BlZWNo', 'audio/mpeg');
    await player.stop();

    expect(fileSystem.deleted).toContain('file:///cache/stuffstash-voice-old.mp3');
    expect(fileSystem.deleted).not.toContain('file:///cache/unrelated.mp3');
    expect(fileSystem.files['file:///cache/unrelated.mp3']).toBe('keep');
  });

  it('writes tts chunks to cache files and plays them with Expo Audio', async () => {
    const audio = new FakeAudio(new FakeRecorder('file:///recording.m4a'));
    const fileSystem = new FakeFileSystem({});
    const player = new ExpoVoiceAudioPlayerCore(audio, fileSystem);

    await player.playChunk('c3BlZWNo', 'audio/mpeg');

    expect(fileSystem.writes).toHaveLength(1);
    expect(fileSystem.writes[0]).toMatchObject({
      contents: 'c3BlZWNo',
      encoding: 'base64'
    });
    expect(fileSystem.writes[0]?.uri).toMatch(/stuffstash-voice-.+\.mp3$/);
    expect(audio.players[0]?.finished).toBe(true);
    expect(audio.players[0]?.removed).toBe(true);
    expect(fileSystem.deleted).toEqual([fileSystem.writes[0]?.uri]);

    await player.stop();

    expect(fileSystem.deleted).toEqual([fileSystem.writes[0]?.uri]);
  });

  it('deletes playback files even when disposing the player fails', async () => {
    const audio = new FakeAudio(new FakeRecorder('file:///recording.m4a'));
    audio.failRemove = true;
    const fileSystem = new FakeFileSystem({});
    const player = new ExpoVoiceAudioPlayerCore(audio, fileSystem);

    await expect(player.playChunk('c3BlZWNo', 'audio/mpeg')).rejects.toThrow('remove failed');

    expect(fileSystem.deleted).toEqual([fileSystem.writes[0]?.uri]);
  });

  it('continues stale cleanup when one temp file delete fails', async () => {
    const audio = new FakeAudio(new FakeRecorder('file:///recording.m4a'));
    const fileSystem = new FakeFileSystem({
      'file:///cache/stuffstash-voice-one.mp3': 'old',
      'file:///cache/stuffstash-voice-two.mp3': 'old',
      'file:///cache/unrelated.mp3': 'keep'
    });
    fileSystem.failDeleteUris.add('file:///cache/stuffstash-voice-one.mp3');
    const player = new ExpoVoiceAudioPlayerCore(audio, fileSystem);

    await player.stop();

    expect(fileSystem.deleted).toContain('file:///cache/stuffstash-voice-two.mp3');
    expect(fileSystem.files['file:///cache/unrelated.mp3']).toBe('keep');
  });
});

class FakeRecorder {
  prepared = false;
  recording = false;
  metering: number | undefined;

  constructor(readonly uri: string | null) {}

  async prepareToRecordAsync(): Promise<void> {
    this.prepared = true;
  }

  record(): void {
    this.recording = true;
  }

  async stop(): Promise<void> {
    this.recording = false;
  }

  getStatus(): { readonly metering?: number } {
    return { metering: this.metering };
  }
}

class FakePlayer {
  readonly currentStatus = { duration: 1 };
  playing = false;
  finished = false;
  removed = false;
  private listener: ((status: { readonly didJustFinish?: boolean }) => void) | null = null;

  constructor(readonly uri: string, private readonly shouldFailRemove = false) {}

  addListener(_event: 'playbackStatusUpdate', listener: (status: { readonly didJustFinish?: boolean }) => void): { remove(): void } {
    this.listener = listener;
    return {
      remove: () => {
        this.listener = null;
      }
    };
  }

  play(): void {
    this.playing = true;
    queueMicrotask(() => {
      this.finished = true;
      this.playing = false;
      this.listener?.({ didJustFinish: true });
    });
  }

  pause(): void {
    this.playing = false;
  }

  remove(): void {
    if (this.shouldFailRemove) {
      throw new Error('remove failed');
    }
    this.removed = true;
  }
}

class FakeAudio {
  granted = true;
  failRemove = false;
  readonly modes: Array<Record<string, boolean>> = [];
  readonly players: FakePlayer[] = [];

  constructor(private readonly recorder: FakeRecorder) {}

  async requestRecordingPermissionsAsync(): Promise<{ readonly granted: boolean }> {
    return { granted: this.granted };
  }

  async setAudioModeAsync(mode: Record<string, boolean>): Promise<void> {
    this.modes.push(mode);
  }

  createRecorder(): FakeRecorder {
    return this.recorder;
  }

  createAudioPlayer(uri: string): FakePlayer {
    const player = new FakePlayer(uri, this.failRemove);
    this.players.push(player);
    return player;
  }
}

class FakeFileSystem {
  readonly cacheDirectory = 'file:///cache/';
  readonly writes: Array<{ readonly uri: string; readonly contents: string; readonly encoding: string }> = [];
  readonly reads: string[] = [];
  readonly deleted: string[] = [];
  readonly failDeleteUris = new Set<string>();

  constructor(readonly files: Record<string, string>) {}

  async readAsStringAsync(uri: string): Promise<string> {
    this.reads.push(uri);
    return this.files[uri] ?? '';
  }

  async readDirectoryAsync(uri: string): Promise<string[]> {
    return Object.keys(this.files)
      .filter((fileUri) => fileUri.startsWith(uri))
      .map((fileUri) => fileUri.slice(uri.length));
  }

  async writeAsStringAsync(uri: string, contents: string, options: { readonly encoding: string }): Promise<void> {
    this.files[uri] = contents;
    this.writes.push({ uri, contents, encoding: options.encoding });
  }

  async deleteAsync(uri: string): Promise<void> {
    if (this.failDeleteUris.has(uri)) {
      throw new Error(`delete failed: ${uri}`);
    }
    delete this.files[uri];
    this.deleted.push(uri);
  }
}
