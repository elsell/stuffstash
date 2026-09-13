import { Stack } from 'expo-router';
import { useAppearancePalette } from '../../../ui/theme/AppearanceContext';
export default function BrowseLayout(){
 const palette=useAppearancePalette();
 return <Stack screenOptions={{title:'Browse',headerStyle:{backgroundColor:palette.surface},headerTintColor:palette.action,headerTitleStyle:{color:palette.text},contentStyle:{backgroundColor:palette.background}}} />;
}
