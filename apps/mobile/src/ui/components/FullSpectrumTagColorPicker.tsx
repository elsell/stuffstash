import { useEffect, useMemo, useState } from 'react';
import { PanResponder, Platform, Pressable, StyleSheet, Text, View } from 'react-native';
import Svg, { Defs, LinearGradient, Rect, Stop } from 'react-native-svg';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { adjustSpectrumValue, androidSpectrumAccessibility, expoUIColorPickerAvailable, fullSpectrumPickerKind, spectrumGestureOwnership } from './FullSpectrumTagColorPickerPresentation';
import { NativeTagColorPicker } from './NativeTagColorPicker';

export function FullSpectrumTagColorPicker({ compact = false, disabled = false, onChange, value }: { readonly compact?: boolean; readonly disabled?: boolean; readonly onChange: (value: string) => void; readonly value: string }) {
  const pickerKind = fullSpectrumPickerKind(Platform.OS, expoUIColorPickerAvailable(Platform.OS));
  if (pickerKind === 'native-ios') {
    return <NativeTagColorPicker disabled={disabled} onChange={onChange} value={value} />;
  }
  return <AndroidSpectrum compact={compact} disabled={disabled} onChange={onChange} value={value} />;
}

function AndroidSpectrum({ compact, disabled, onChange, value }: { readonly compact: boolean; readonly disabled: boolean; readonly onChange: (value: string) => void; readonly value: string }) {
  const palette = useAppearancePalette();
  const initial = rgbToHsv(hexToRgb(value) ?? { red: 47, green: 128, blue: 237 });
  const [hue, setHue] = useState(initial.hue);
  const [saturation, setSaturation] = useState(initial.saturation);
  const [brightness, setBrightness] = useState(initial.brightness);
  const [spectrumSize, setSpectrumSize] = useState({ width: 320, height: compact ? compactSpectrumHeight : spectrumHeight });
  useEffect(() => {
    const rgb = hexToRgb(value);
    if (!rgb) return;
    const next = rgbToHsv(rgb);
    setHue(next.hue); setSaturation(next.saturation); setBrightness(next.brightness);
  }, [value]);
  const hueColor = rgbToHex(hsvToRgb({ hue, saturation: 1, brightness: 1 }));
  const spectrumResponder = useMemo(() => PanResponder.create({
    ...spectrumGestureOwnership,
    onStartShouldSetPanResponder: () => !disabled,
    onMoveShouldSetPanResponder: () => !disabled,
    onPanResponderGrant: (event) => selectSpectrum(event.nativeEvent.locationX, event.nativeEvent.locationY),
    onPanResponderMove: (event) => selectSpectrum(event.nativeEvent.locationX, event.nativeEvent.locationY)
  }), [disabled, hue, spectrumSize]);
  const hueResponder = useMemo(() => PanResponder.create({
    ...spectrumGestureOwnership,
    onStartShouldSetPanResponder: () => !disabled,
    onMoveShouldSetPanResponder: () => !disabled,
    onPanResponderGrant: (event) => selectHue(event.nativeEvent.locationX),
    onPanResponderMove: (event) => selectHue(event.nativeEvent.locationX)
  }), [disabled, saturation, brightness, spectrumSize.width]);

  function selectSpectrum(x: number, y: number) {
    const nextSaturation = clamp(x / spectrumSize.width); const nextBrightness = 1 - clamp(y / spectrumSize.height);
    setSaturation(nextSaturation); setBrightness(nextBrightness);
    onChange(rgbToHex(hsvToRgb({ hue, saturation: nextSaturation, brightness: nextBrightness })));
  }
  function selectHue(x: number) {
    const nextHue = clamp(x / spectrumSize.width) * 360; setHue(nextHue);
    onChange(rgbToHex(hsvToRgb({ hue: nextHue, saturation, brightness })));
  }
  function apply(next: { hue: number; saturation: number; brightness: number }) {
    setHue(next.hue); setSaturation(next.saturation); setBrightness(next.brightness);
    onChange(rgbToHex(hsvToRgb(next)));
  }
  const accessibility = androidSpectrumAccessibility({ hue, saturation, brightness }, disabled);

  return <View style={[styles.android, disabled && styles.disabled]}>
    <View {...spectrumResponder.panHandlers} {...accessibility.spectrum} onAccessibilityAction={(event) => {
      if (disabled) return;
      apply(adjustSpectrumValue({ hue, saturation, brightness }, 'spectrum', event.nativeEvent.actionName));
    }} onLayout={(event) => setSpectrumSize({ width: event.nativeEvent.layout.width, height: event.nativeEvent.layout.height })} style={[styles.spectrum, compact && styles.compactSpectrum, { borderColor: palette.border }]}>
      <Svg height="100%" width="100%"><Defs><LinearGradient id="white" x1="0" y1="0" x2="1" y2="0"><Stop offset="0" stopColor="#FFFFFF" /><Stop offset="1" stopColor="#FFFFFF" stopOpacity="0" /></LinearGradient><LinearGradient id="black" x1="0" y1="0" x2="0" y2="1"><Stop offset="0" stopColor="#000000" stopOpacity="0" /><Stop offset="1" stopColor="#000000" /></LinearGradient></Defs><Rect fill={hueColor} height="100%" width="100%" /><Rect fill="url(#white)" height="100%" width="100%" /><Rect fill="url(#black)" height="100%" width="100%" /></Svg>
      <View pointerEvents="none" style={[styles.marker, { borderColor: palette.surface, left: `${saturation * 100}%`, top: `${(1 - brightness) * 100}%` }]} />
    </View>
    <View {...hueResponder.panHandlers} {...accessibility.hue} onAccessibilityAction={(event) => {
      if (disabled) return;
      apply(adjustSpectrumValue({ hue, saturation, brightness }, 'hue', event.nativeEvent.actionName));
    }} style={[styles.hue, { borderColor: palette.border }]}><Svg height="100%" width="100%"><Defs><LinearGradient id="hue" x1="0" y1="0" x2="1" y2="0">{['#FF0000','#FFFF00','#00FF00','#00FFFF','#0000FF','#FF00FF','#FF0000'].map((color, index) => <Stop key={color + index} offset={index / 6} stopColor={color} />)}</LinearGradient></Defs><Rect fill="url(#hue)" height="100%" width="100%" /></Svg><View pointerEvents="none" style={[styles.hueMarker, { borderColor: palette.surface, left: `${(hue / 360) * 100}%` }]} /></View>
    {!compact ? <><Adjustment label="Hue" value={`${Math.round(hue)} degrees`} disabled={disabled} onDecrease={() => apply({ hue: (hue + 355) % 360, saturation, brightness })} onIncrease={() => apply({ hue: (hue + 5) % 360, saturation, brightness })} /><Adjustment label="Saturation" value={`${Math.round(saturation * 100)} percent`} disabled={disabled} onDecrease={() => apply({ hue, saturation: clamp(saturation - 0.05), brightness })} onIncrease={() => apply({ hue, saturation: clamp(saturation + 0.05), brightness })} /><Adjustment label="Brightness" value={`${Math.round(brightness * 100)} percent`} disabled={disabled} onDecrease={() => apply({ hue, saturation, brightness: clamp(brightness - 0.05) })} onIncrease={() => apply({ hue, saturation, brightness: clamp(brightness + 0.05) })} /></> : null}
  </View>;
}

function Adjustment({ disabled, label, onDecrease, onIncrease, value }: { readonly disabled: boolean; readonly label: string; readonly onDecrease: () => void; readonly onIncrease: () => void; readonly value: string }) {
  const palette = useAppearancePalette();
  return <View accessibilityLabel={`${label}, ${value}`} style={styles.adjustment}><Text style={[styles.adjustmentLabel, { color: palette.text }]}>{label}</Text><Pressable accessibilityLabel={`Decrease ${label.toLocaleLowerCase()}`} accessibilityRole="button" accessibilityState={{ disabled }} disabled={disabled} onPress={onDecrease} style={[styles.adjustButton, { borderColor: palette.border }]}><Text style={[styles.adjustButtonText, { color: palette.action }]}>−</Text></Pressable><Text style={[styles.adjustmentValue, { color: palette.textMuted }]}>{value}</Text><Pressable accessibilityLabel={`Increase ${label.toLocaleLowerCase()}`} accessibilityRole="button" accessibilityState={{ disabled }} disabled={disabled} onPress={onIncrease} style={[styles.adjustButton, { borderColor: palette.border }]}><Text style={[styles.adjustButtonText, { color: palette.action }]}>+</Text></Pressable></View>;
}

const spectrumHeight = 160;
const compactSpectrumHeight = 112;
function clamp(value: number) { return Math.max(0, Math.min(1, value)); }
function validColor(value: string) { return /^#[0-9a-fA-F]{6}$/.test(value); }
function hexToRgb(value: string) { if (!validColor(value)) return undefined; return { red: Number.parseInt(value.slice(1, 3), 16), green: Number.parseInt(value.slice(3, 5), 16), blue: Number.parseInt(value.slice(5, 7), 16) }; }
function rgbToHex(rgb: { red: number; green: number; blue: number }) { return `#${[rgb.red, rgb.green, rgb.blue].map((channel) => Math.round(channel).toString(16).padStart(2, '0')).join('')}`.toUpperCase(); }
function hsvToRgb({ hue, saturation, brightness }: { hue: number; saturation: number; brightness: number }) { const c = brightness * saturation; const x = c * (1 - Math.abs(((hue / 60) % 2) - 1)); const m = brightness - c; const values = hue < 60 ? [c,x,0] : hue < 120 ? [x,c,0] : hue < 180 ? [0,c,x] : hue < 240 ? [0,x,c] : hue < 300 ? [x,0,c] : [c,0,x]; return { red: (values[0] + m) * 255, green: (values[1] + m) * 255, blue: (values[2] + m) * 255 }; }
function rgbToHsv({ red, green, blue }: { red: number; green: number; blue: number }) { const r = red / 255, g = green / 255, b = blue / 255; const max = Math.max(r,g,b), min = Math.min(r,g,b), delta = max - min; const hue = delta === 0 ? 0 : max === r ? 60 * (((g-b)/delta) % 6) : max === g ? 60 * (((b-r)/delta)+2) : 60 * (((r-g)/delta)+4); return { hue: hue < 0 ? hue + 360 : hue, saturation: max === 0 ? 0 : delta/max, brightness: max }; }
const styles = StyleSheet.create({ android: { gap: 10 }, disabled: { opacity: 0.55 }, spectrum: { borderRadius: 10, borderWidth: 1, height: spectrumHeight, overflow: 'hidden', position: 'relative', width: '100%' }, compactSpectrum: { height: 112 }, marker: { borderRadius: 10, borderWidth: 3, height: 20, marginLeft: -10, marginTop: -10, position: 'absolute', width: 20 }, hue: { borderRadius: 10, borderWidth: 1, height: 44, overflow: 'hidden', position: 'relative', width: '100%' }, hueMarker: { borderRadius: 3, borderWidth: 3, height: 44, marginLeft: -4, position: 'absolute', width: 8 }, adjustment: { alignItems: 'center', flexDirection: 'row', gap: 8, minHeight: 44 }, adjustmentLabel: { flex: 1, fontSize: 14, fontWeight: '600' }, adjustmentValue: { fontSize: 13, minWidth: 82, textAlign: 'center' }, adjustButton: { alignItems: 'center', borderRadius: 8, borderWidth: 1, height: 44, justifyContent: 'center', width: 44 }, adjustButtonText: { fontSize: 22, fontWeight: '700' } });
