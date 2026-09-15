import { useRef } from 'react';
import { Platform, type TextInputProps } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';

type Props = Omit<TextInputProps, 'value' | 'defaultValue'> & { readonly value: string };

export function AddAssetNameField(props: Props) {
  return Platform.OS === 'ios' ? <NativeNameField {...props} /> : <AppTextInput {...props} />;
}

function NativeNameField({ value, ...props }: Props) {
  const initialValue = useRef(value).current;
  return <AppTextInput {...props} defaultValue={initialValue} />;
}
