import { Mic, SendHorizontal } from 'lucide-react-native';
import type { VoiceAccessorySymbolProps } from './VoiceAccessorySymbol.types';

export function VoiceAccessorySymbol({ name, color, size }: VoiceAccessorySymbolProps) {
  const Symbol = name === 'microphone' ? Mic : SendHorizontal;
  return <Symbol color={color} size={size} strokeWidth={name === 'microphone' ? 2.5 : 2.6} />;
}
