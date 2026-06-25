import type { RecordedVoiceAudio, VoiceAudioPlayer, VoiceAudioRecorder } from '../../application/voice/RealtimeVoiceSession';

export type NativePermissionResponse = {
  readonly granted: boolean;
};

export type NativeAudioRecorder = {
  readonly uri: string | null;
  prepareToRecordAsync(): Promise<void>;
  record(): void;
  stop(): Promise<void>;
};

export type NativeAudioPlayer = {
  play(): void;
  pause(): void;
  remove(): void;
};

export type ExpoVoiceAudioNative = {
  requestRecordingPermissionsAsync(): Promise<NativePermissionResponse>;
  setAudioModeAsync(mode: {
    readonly allowsRecording?: boolean;
    readonly playsInSilentMode?: boolean;
    readonly shouldPlayInBackground?: boolean;
  }): Promise<void>;
  createRecorder(): NativeAudioRecorder;
  createAudioPlayer(uri: string): NativeAudioPlayer;
};

export type ExpoVoiceFileSystem = {
  readonly cacheDirectory: string | null;
  readAsStringAsync(uri: string, options: { readonly encoding: 'base64' }): Promise<string>;
  writeAsStringAsync(uri: string, contents: string, options: { readonly encoding: 'base64' }): Promise<void>;
  deleteAsync?(uri: string, options?: { readonly idempotent?: boolean }): Promise<void>;
};

export class ExpoVoiceAudioRecorderCore implements VoiceAudioRecorder {
  private recorder: NativeAudioRecorder | null = null;

  constructor(
    private readonly audio: ExpoVoiceAudioNative,
    private readonly fileSystem: ExpoVoiceFileSystem
  ) {}

  async start(): Promise<void> {
    const permission = await this.audio.requestRecordingPermissionsAsync();
    if (!permission.granted) {
      throw new Error('Microphone permission is required for voice control.');
    }

    await this.audio.setAudioModeAsync({
      allowsRecording: true,
      playsInSilentMode: true
    });

    const recorder = this.audio.createRecorder();
    await recorder.prepareToRecordAsync();
    recorder.record();
    this.recorder = recorder;
  }

  async stop(): Promise<RecordedVoiceAudio> {
    const recorder = this.recorder;
    if (recorder === null) {
      throw new Error('Voice recording has not started.');
    }

    this.recorder = null;
    await recorder.stop();
    await this.audio.setAudioModeAsync({
      allowsRecording: false,
      playsInSilentMode: true
    });

    if (!recorder.uri) {
      throw new Error('Voice recording did not produce an audio file.');
    }

    const audioBase64 = await this.fileSystem.readAsStringAsync(recorder.uri, { encoding: 'base64' });
    if (audioBase64.length === 0) {
      throw new Error('Voice recording produced an empty audio file.');
    }

    return {
      mimeType: 'audio/mp4',
      sampleRate: 44100,
      channels: 1,
      chunksBase64: [audioBase64]
    };
  }
}

export class ExpoVoiceAudioPlayerCore implements VoiceAudioPlayer {
  private readonly players: NativeAudioPlayer[] = [];
  private readonly tempUris: string[] = [];

  constructor(
    private readonly audio: ExpoVoiceAudioNative,
    private readonly fileSystem: ExpoVoiceFileSystem
  ) {}

  async playChunk(audioBase64: string, mimeType: string): Promise<void> {
    const cacheDirectory = this.fileSystem.cacheDirectory;
    if (!cacheDirectory) {
      throw new Error('Audio cache directory is unavailable.');
    }
    const uri = `${cacheDirectory}stuffstash-voice-${Date.now().toString(36)}-${this.tempUris.length + 1}${audioExtension(mimeType)}`;
    await this.fileSystem.writeAsStringAsync(uri, audioBase64, { encoding: 'base64' });
    const player = this.audio.createAudioPlayer(uri);
    this.players.push(player);
    this.tempUris.push(uri);
    player.play();
  }

  async stop(): Promise<void> {
    for (const player of this.players.splice(0)) {
      player.pause();
      player.remove();
    }

    const uris = this.tempUris.splice(0);
    await Promise.all(uris.map((uri) => this.fileSystem.deleteAsync?.(uri, { idempotent: true }) ?? Promise.resolve()));
  }
}

function audioExtension(mimeType: string): string {
  return mimeType === 'audio/mpeg' ? '.mp3' : '.audio';
}
