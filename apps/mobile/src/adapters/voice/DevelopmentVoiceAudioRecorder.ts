import type { RecordedVoiceAudio, VoiceAudioRecorder } from '../../application/voice/RealtimeVoiceSession';

export class DevelopmentVoiceAudioRecorder implements VoiceAudioRecorder {
  private recording = false;

  async start(): Promise<void> {
    this.recording = true;
  }

  async stop(): Promise<RecordedVoiceAudio> {
    if (!this.recording) {
      throw new Error('Voice recording has not started.');
    }
    this.recording = false;
    return {
      mimeType: 'audio/mp4',
      sampleRate: 44100,
      channels: 1,
      chunksBase64: ['ZGV2ZWxvcG1lbnQtdm9pY2UtYXVkaW8=']
    };
  }
}
