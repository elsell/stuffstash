import type { RecordedVoiceAudio, VoiceAudioPlayer, VoiceAudioRecorder } from '../../application/voice/RealtimeVoiceSession';

export type NativePermissionResponse = {
  readonly granted: boolean;
};

export type NativeAudioRecorder = {
  readonly uri: string | null;
  prepareToRecordAsync(): Promise<void>;
  record(): void;
  stop(): Promise<void>;
  getStatus?(): { readonly metering?: number };
};

export type NativeAudioPlayer = {
  readonly currentStatus?: { readonly duration?: number };
  addListener?(
    event: 'playbackStatusUpdate',
    listener: (status: { readonly didJustFinish?: boolean }) => void
  ): { remove(): void };
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
  readDirectoryAsync?(uri: string): Promise<string[]>;
  writeAsStringAsync(uri: string, contents: string, options: { readonly encoding: 'base64' }): Promise<void>;
  deleteAsync?(uri: string, options?: { readonly idempotent?: boolean }): Promise<void>;
};

const targetAudioChunkRawBytes = 256 * 1024;
const voiceTempFilePrefix = 'stuffstash-voice-';
const playbackCompletionTimeoutMs = 120_000;

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

    let audioBase64 = '';
    try {
      audioBase64 = await this.fileSystem.readAsStringAsync(recorder.uri, { encoding: 'base64' });
      if (audioBase64.length === 0) {
        throw new Error('Voice recording produced an empty audio file.');
      }
    } finally {
      await this.fileSystem.deleteAsync?.(recorder.uri, { idempotent: true });
    }

    return {
      mimeType: 'audio/mp4',
      sampleRate: 44100,
      channels: 1,
      chunksBase64: chunkBase64Audio(audioBase64)
    };
  }

  async cancel(): Promise<void> {
    const recorder = this.recorder;
    if (recorder === null) {
      return;
    }

    this.recorder = null;
    try {
      await recorder.stop();
    } finally {
      await this.audio.setAudioModeAsync({
        allowsRecording: false,
        playsInSilentMode: true
      });
      if (recorder.uri) {
        await this.fileSystem.deleteAsync?.(recorder.uri, { idempotent: true });
      }
    }
  }

  recordingLevel(): number {
    const recorder = this.recorder;
    if (recorder === null || !recorder.getStatus) {
      return 0;
    }

    try {
      return normalizeDbfsLevel(recorder.getStatus().metering);
    } catch {
      return 0;
    }
  }
}

export function normalizeDbfsLevel(metering: number | undefined): number {
  if (typeof metering !== 'number' || !Number.isFinite(metering)) {
    return 0;
  }
  const clampedDbfs = Math.max(-60, Math.min(0, metering));
  return Math.max(0, Math.min(1, (clampedDbfs + 60) / 60));
}

export class ExpoVoiceAudioPlayerCore implements VoiceAudioPlayer {
  private readonly players: NativeAudioPlayer[] = [];
  private readonly tempUris: string[] = [];

  constructor(
    private readonly audio: ExpoVoiceAudioNative,
    private readonly fileSystem: ExpoVoiceFileSystem
  ) {}

  async playChunk(audioBase64: string, mimeType: string): Promise<void> {
    await cleanupStaleVoiceTempFiles(this.fileSystem);
    const cacheDirectory = this.fileSystem.cacheDirectory;
    if (!cacheDirectory) {
      throw new Error('Audio cache directory is unavailable.');
    }
    const uri = `${cacheDirectory}${voiceTempFilePrefix}${Date.now().toString(36)}-${this.tempUris.length + 1}${audioExtension(mimeType)}`;
    await this.fileSystem.writeAsStringAsync(uri, audioBase64, { encoding: 'base64' });
    this.tempUris.push(uri);
    let player: NativeAudioPlayer | null = null;
    try {
      player = this.audio.createAudioPlayer(uri);
      this.players.push(player);
      await playUntilFinished(player);
    } finally {
      let cleanupError: unknown;
      try {
        if (player !== null) {
          await this.disposePlayer(player);
        }
      } catch (error) {
        cleanupError = error;
      } finally {
        try {
          await this.deleteTempUri(uri);
        } catch (error) {
          cleanupError ??= error;
        }
      }
      if (cleanupError) {
        throw cleanupError;
      }
    }
  }

  async stop(): Promise<void> {
    const uris = this.tempUris.splice(0);
    const players = this.players.splice(0);
    await Promise.allSettled(players.map((player) => this.disposePlayer(player)));
    await Promise.allSettled(uris.map((uri) => this.deleteTempUri(uri)));
    await cleanupStaleVoiceTempFiles(this.fileSystem);
  }

  private async disposePlayer(player: NativeAudioPlayer): Promise<void> {
    removeFromArray(this.players, player);
    player.pause();
    player.remove();
  }

  private async deleteTempUri(uri: string): Promise<void> {
    removeFromArray(this.tempUris, uri);
    await this.fileSystem.deleteAsync?.(uri, { idempotent: true });
  }
}

function playUntilFinished(player: NativeAudioPlayer): Promise<void> {
  return new Promise((resolve) => {
    let resolved = false;
    const finish = () => {
      if (resolved) {
        return;
      }
      resolved = true;
      subscription?.remove();
      clearTimeout(timeout);
      resolve();
    };
    const timeout = setTimeout(finish, playbackTimeoutMs(player));
    const subscription = player.addListener?.('playbackStatusUpdate', (status) => {
      if (status.didJustFinish) {
        finish();
      }
    });
    player.play();
    if (!subscription) {
      finish();
    }
  });
}

function playbackTimeoutMs(player: NativeAudioPlayer): number {
  const durationSeconds = player.currentStatus?.duration;
  if (typeof durationSeconds === 'number' && durationSeconds > 0) {
    return Math.max(playbackCompletionTimeoutMs, Math.ceil(durationSeconds * 1000) + 5000);
  }
  return playbackCompletionTimeoutMs;
}

async function cleanupStaleVoiceTempFiles(fileSystem: ExpoVoiceFileSystem): Promise<void> {
  const cacheDirectory = fileSystem.cacheDirectory;
  if (!cacheDirectory || !fileSystem.readDirectoryAsync || !fileSystem.deleteAsync) {
    return;
  }
  const entries = await fileSystem.readDirectoryAsync(cacheDirectory);
  await Promise.allSettled(
    entries
      .filter((entry) => entry.startsWith(voiceTempFilePrefix))
      .map((entry) => fileSystem.deleteAsync?.(cacheDirectory + entry, { idempotent: true }) ?? Promise.resolve())
  );
}

function chunkBase64Audio(audioBase64: string): readonly string[] {
  const chunkBase64Chars = Math.floor(targetAudioChunkRawBytes / 3) * 4;
  const chunks: string[] = [];
  for (let index = 0; index < audioBase64.length; index += chunkBase64Chars) {
    chunks.push(audioBase64.slice(index, index + chunkBase64Chars));
  }
  return chunks;
}

function audioExtension(mimeType: string): string {
  return mimeType === 'audio/mpeg' ? '.mp3' : '.audio';
}

function removeFromArray<T>(items: T[], item: T): void {
  const index = items.indexOf(item);
  if (index >= 0) {
    items.splice(index, 1);
  }
}
