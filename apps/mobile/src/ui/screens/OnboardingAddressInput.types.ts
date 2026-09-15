export type OnboardingAddressInputProps = {
  readonly initialValue: string;
  readonly onChangeText: (value: string) => void;
  readonly disabled: boolean;
  readonly onSubmit: () => void;
};
