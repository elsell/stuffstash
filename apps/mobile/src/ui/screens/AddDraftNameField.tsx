import { useRef } from 'react';
import { Platform } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
import type { AddDraftNameFieldProps as Props } from './AddDraftNameField.types';

export function AddDraftNameField(props: Props) {
  return Platform.OS === 'ios' ? <NativeNameField {...props} /> : <AppTextInput {...props} />;
}

function NativeNameField({ value, ...props }: Props) {
  const initialValue = useRef(value).current;
  return <AppTextInput {...props} defaultValue={initialValue} />;
}
