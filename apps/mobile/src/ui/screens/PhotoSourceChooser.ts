import { t } from '../../presentation/localization';
import { ActionSheetIOS, Alert, Platform } from 'react-native';

export function showPhotoSourceChooser({
  isCurrent,
  onCamera,
  onLibrary
}: {
  readonly isCurrent: () => boolean;
  readonly onCamera: () => void;
  readonly onLibrary: () => void;
}) {
  if (!isCurrent()) return;
  const choose = (action: () => void) => { if (isCurrent()) action(); };
  if (Platform.OS === 'ios') {
    ActionSheetIOS.showActionSheetWithOptions(
      {
        options: ['Take Photo', 'Choose from Library', 'Cancel'],
        cancelButtonIndex: 2
      },
      (buttonIndex) => {
        if (buttonIndex === 0) {
          choose(onCamera);
        }
        if (buttonIndex === 1) {
          choose(onLibrary);
        }
      }
    );
    return;
  }

  Alert.alert(t('mobile.PhotoSourceChooser.addPhotos'), undefined, [
    { text: t('mobile.PhotoSourceChooser.takePhoto'), onPress: () => choose(onCamera) },
    { text: t('mobile.PhotoSourceChooser.chooseFromLibrary'), onPress: () => choose(onLibrary) },
    { text: t('mobile.PhotoSourceChooser.cancel'), style: 'cancel' }
  ]);
}
