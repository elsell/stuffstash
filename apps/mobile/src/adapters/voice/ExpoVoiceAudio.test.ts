import { describe, expect, it } from 'vitest';
import { ExpoVoiceAudioPlayerCore, ExpoVoiceAudioRecorderCore } from './ExpoVoiceAudioCore';

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
  });

  it('rejects recording when microphone permission is denied', async () => {
    const audio = new FakeAudio(new FakeRecorder('file:///recording.m4a'));
    audio.granted = false;
    const voiceRecorder = new ExpoVoiceAudioRecorderCore(audio, new FakeFileSystem({}));

    await expect(voiceRecorder.start()).rejects.toThrow('Microphone permission is required');
  });
});

describe('ExpoVoiceAudioPlayer', () => {
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
    expect(audio.players[0]?.playing).toBe(true);

    await player.stop();

    expect(audio.players[0]?.removed).toBe(true);
    expect(fileSystem.deleted).toEqual([fileSystem.writes[0]?.uri]);
  });
});

class FakeRecorder {
  prepared = false;
  recording = false;

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
}

class FakePlayer {
  playing = false;
  removed = false;

  constructor(readonly uri: string) {}

  play(): void {
    this.playing = true;
  }

  pause(): void {
    this.playing = false;
  }

  remove(): void {
    this.removed = true;
  }
}

class FakeAudio {
  granted = true;
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
    const player = new FakePlayer(uri);
    this.players.push(player);
    return player;
  }
}

class FakeFileSystem {
  readonly cacheDirectory = 'file:///cache/';
  readonly writes: Array<{ readonly uri: string; readonly contents: string; readonly encoding: string }> = [];
  readonly deleted: string[] = [];

  constructor(private readonly files: Record<string, string>) {}

  async readAsStringAsync(uri: string): Promise<string> {
    return this.files[uri] ?? '';
  }

  async writeAsStringAsync(uri: string, contents: string, options: { readonly encoding: string }): Promise<void> {
    this.files[uri] = contents;
    this.writes.push({ uri, contents, encoding: options.encoding });
  }

  async deleteAsync(uri: string): Promise<void> {
    delete this.files[uri];
    this.deleted.push(uri);
  }
}
