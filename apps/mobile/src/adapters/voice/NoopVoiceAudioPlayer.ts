import type { VoiceAudioPlayer } from '../../application/voice/RealtimeVoiceSession';

export class NoopVoiceAudioPlayer implements VoiceAudioPlayer {
  async playChunk(): Promise<void> {}

  async stop(): Promise<void> {}
}
