import { ActionSheetIOS, Alert, Platform } from 'react-native';

export function showPhotoSourceChooser({
  onCamera,
  onLibrary
}: {
  readonly onCamera: () => void;
  readonly onLibrary: () => void;
}) {
  if (Platform.OS === 'ios') {
    ActionSheetIOS.showActionSheetWithOptions(
      {
        options: ['Take Photo', 'Choose from Library', 'Cancel'],
        cancelButtonIndex: 2
      },
      (buttonIndex) => {
        if (buttonIndex === 0) {
          onCamera();
        }
        if (buttonIndex === 1) {
          onLibrary();
        }
      }
    );
    return;
  }

  Alert.alert('Add photos', undefined, [
    { text: 'Take Photo', onPress: onCamera },
    { text: 'Choose from Library', onPress: onLibrary },
    { text: 'Cancel', style: 'cancel' }
  ]);
}
