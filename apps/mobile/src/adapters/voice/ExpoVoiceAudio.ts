import { createAudioPlayer, requestRecordingPermissionsAsync, setAudioModeAsync } from 'expo-audio';
import type { AudioPlayer } from 'expo-audio';
import AudioModule from 'expo-audio/src/AudioModule';
import { RecordingPresets } from 'expo-audio';
import * as FileSystem from 'expo-file-system/legacy';
import type { VoiceAudioPlayer, VoiceAudioRecorder } from '../../application/voice/RealtimeVoiceSession';
import {
  ExpoVoiceAudioPlayerCore,
  ExpoVoiceAudioRecorderCore,
  type ExpoVoiceAudioNative
} from './ExpoVoiceAudioCore';

const nativeAudio: ExpoVoiceAudioNative = {
  requestRecordingPermissionsAsync,
  setAudioModeAsync,
  createRecorder: () => new AudioModule.AudioRecorder(RecordingPresets.HIGH_QUALITY),
  createAudioPlayer: (uri: string) => createAudioPlayer({ uri }) as AudioPlayer
};

export class ExpoVoiceAudioRecorder implements VoiceAudioRecorder {
  private readonly core: ExpoVoiceAudioRecorderCore;

  constructor() {
    this.core = new ExpoVoiceAudioRecorderCore(nativeAudio, FileSystem);
  }

  start() {
    return this.core.start();
  }

  stop() {
    return this.core.stop();
  }
}

export class ExpoVoiceAudioPlayer implements VoiceAudioPlayer {
  private readonly core: ExpoVoiceAudioPlayerCore;

  constructor() {
    this.core = new ExpoVoiceAudioPlayerCore(nativeAudio, FileSystem);
  }

  playChunk(audioBase64: string, mimeType: string) {
    return this.core.playChunk(audioBase64, mimeType);
  }

  stop() {
    return this.core.stop();
  }
}
