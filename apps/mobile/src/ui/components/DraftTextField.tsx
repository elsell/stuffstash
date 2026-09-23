import { AppTextInput } from './AppTextInput';
import type { DraftTextFieldProps } from './DraftTextField.types';

export function DraftTextField(props: DraftTextFieldProps) {
  return <AppTextInput {...props} />;
}
