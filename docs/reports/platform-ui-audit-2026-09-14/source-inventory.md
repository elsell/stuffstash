# Source coverage inventory

Revision `8392eeb4cedbd7e327c344f4e6b5a78a2fff4851`.

This enumerates 62 route/layout files and 313 UI implementation/style files
under the declared roots. Enumeration is not a per-file runtime pass. Family-level
assessments and inspected evidence are in `surfaces.md` and the main report.
Platform fallback files are included; test harnesses are excluded.

## Route entrypoints

| Entrypoint | Client | Imported surface / dispatcher |
| --- | --- | --- |
| [(tabs)/(home)/_layout.tsx](../../../apps/mobile/src/app/(tabs)/(home)/_layout.tsx) | Mobile | AppearanceContext, NativeTabHeader |
| [(tabs)/(home)/index.tsx](../../../apps/mobile/src/app/(tabs)/(home)/index.tsx) | Mobile | RefreshExpirationHome, MobileServerStateProvider, ExpirationHomeEntry, AppServicesContext, NotificationHomeEntry, HomeScreen |
| [(tabs)/_layout.tsx](../../../apps/mobile/src/app/(tabs)/_layout.tsx) | Mobile | VoiceBottomAccessory |
| [(tabs)/search/_layout.tsx](../../../apps/mobile/src/app/(tabs)/search/_layout.tsx) | Mobile | AppearanceContext, NativeTabHeader |
| [(tabs)/search/index.tsx](../../../apps/mobile/src/app/(tabs)/search/index.tsx) | Mobile | AppServicesContext, BrowseRouteParams, SearchScreen |
| [_layout.tsx](../../../apps/mobile/src/app/_layout.tsx) | Mobile | PushNotificationNavigation, VoiceConversationReturn, AppServicesContext, InventoryInvitationLinkContext, AppearanceContext, AssetNativeSheetOptions, AppKeyboardAccessory, AppKeyboardProvider |
| [add.tsx](../../../apps/mobile/src/app/add.tsx) | Mobile | AppServicesContext, AddAssetInitialParent, AddAssetScreen |
| [assets/[assetId]/checkouts.tsx](../../../apps/mobile/src/app/assets/[assetId]/checkouts.tsx) | Mobile | AppServicesContext, AssetCheckoutHistoryScreen |
| [assets/[assetId]/edit.tsx](../../../apps/mobile/src/app/assets/[assetId]/edit.tsx) | Mobile | AppServicesContext, AssetNativeActionSheetScreens |
| [assets/[assetId]/history/[activityId].tsx](../../../apps/mobile/src/app/assets/[assetId]/history/[activityId].tsx) | Mobile | AppServicesContext, AssetHistoryDetailRouteScreen |
| [assets/[assetId]/history/index.tsx](../../../apps/mobile/src/app/assets/[assetId]/history/index.tsx) | Mobile | AppServicesContext, AssetHistoryRouteScreen |
| [assets/[assetId]/index.tsx](../../../apps/mobile/src/app/assets/[assetId]/index.tsx) | Mobile | AppServicesContext, AssetDetailRouteScreen |
| [assets/[assetId]/move-here.tsx](../../../apps/mobile/src/app/assets/[assetId]/move-here.tsx) | Mobile | AppServicesContext, AssetNativeActionSheetScreens |
| [assets/[assetId]/move.tsx](../../../apps/mobile/src/app/assets/[assetId]/move.tsx) | Mobile | AppServicesContext, AssetNativeActionSheetScreens |
| [assets/index.tsx](../../../apps/mobile/src/app/assets/index.tsx) | Mobile | AppServicesContext, InventoryAssetsRouteScreen |
| [browse-filters.tsx](../../../apps/mobile/src/app/browse-filters.tsx) | Mobile | AppServicesContext, MobileServerStateProvider, useBrowseFilterNavigation, BrowseFiltersScreen, BrowseRouteParams, BrowseFilterRouteState, SettingsList, ExpirationRouteState, BrowseExpirationFilter |
| [expiration-filters.tsx](../../../apps/mobile/src/app/expiration-filters.tsx) | Mobile | AppServicesContext, ExpirationFiltersScreen, ExpirationRouteState, MobileServerStateProvider, SettingsList |
| [expiration.tsx](../../../apps/mobile/src/app/expiration.tsx) | Mobile | AppServicesContext, MobileServerStateProvider, expirationRefreshDelay, isAccessFailure, ExpirationWorkspaceScreen, ExpirationRouteState, AssetDetailNavigation |
| [invitations/accept.tsx](../../../apps/mobile/src/app/invitations/accept.tsx) | Mobile | AppServicesContext, InventoryInvitationLinkContext, InventoryInvitationScreen |
| [locations/[locationId]/assets/[assetId]/index.tsx](../../../apps/mobile/src/app/locations/[locationId]/assets/[assetId]/index.tsx) | Mobile | AppServicesContext, AssetDetailRouteScreen |
| [locations/[locationId]/index.tsx](../../../apps/mobile/src/app/locations/[locationId]/index.tsx) | Mobile | AppServicesContext, LocationAssetsRouteScreen |
| [notifications.tsx](../../../apps/mobile/src/app/notifications.tsx) | Mobile | AppServicesContext, MobileServerStateProvider, SettingsScreenState, SettingsList, NotificationInboxScreen, AssetDetailNavigation |
| [settings/about.tsx](../../../apps/mobile/src/app/settings/about.tsx) | Mobile | AppServicesContext, SettingsDetailScreens |
| [settings/account.tsx](../../../apps/mobile/src/app/settings/account.tsx) | Mobile | AppServicesContext, SettingsDetailScreens |
| [settings/appearance.tsx](../../../apps/mobile/src/app/settings/appearance.tsx) | Mobile | SettingsDetailScreens |
| [settings/connection.tsx](../../../apps/mobile/src/app/settings/connection.tsx) | Mobile | AppServicesContext, SettingsDetailScreens |
| [settings/diagnostics.tsx](../../../apps/mobile/src/app/settings/diagnostics.tsx) | Mobile | AppServicesContext, SettingsDetailScreens |
| [settings/household/asset-types/[resourceId].tsx](../../../apps/mobile/src/app/settings/household/asset-types/[resourceId].tsx) | Mobile | CustomizationRoutes |
| [settings/household/asset-types/index.tsx](../../../apps/mobile/src/app/settings/household/asset-types/index.tsx) | Mobile | CustomizationRoutes |
| [settings/household/asset-types/new.tsx](../../../apps/mobile/src/app/settings/household/asset-types/new.tsx) | Mobile | CustomizationRoutes |
| [settings/household/fields/[resourceId].tsx](../../../apps/mobile/src/app/settings/household/fields/[resourceId].tsx) | Mobile | CustomizationRoutes |
| [settings/household/fields/index.tsx](../../../apps/mobile/src/app/settings/household/fields/index.tsx) | Mobile | CustomizationRoutes |
| [settings/household/fields/new.tsx](../../../apps/mobile/src/app/settings/household/fields/new.tsx) | Mobile | CustomizationRoutes |
| [settings/household/index.tsx](../../../apps/mobile/src/app/settings/household/index.tsx) | Mobile | AppServicesContext, ScopedSettingsScreens |
| [settings/index.tsx](../../../apps/mobile/src/app/settings/index.tsx) | Mobile | AppServicesContext, SettingsScreen, SettingsScreenPresentation |
| [settings/inventory/asset-types/[resourceId].tsx](../../../apps/mobile/src/app/settings/inventory/asset-types/[resourceId].tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/asset-types/index.tsx](../../../apps/mobile/src/app/settings/inventory/asset-types/index.tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/asset-types/new.tsx](../../../apps/mobile/src/app/settings/inventory/asset-types/new.tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/fields/[resourceId].tsx](../../../apps/mobile/src/app/settings/inventory/fields/[resourceId].tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/fields/index.tsx](../../../apps/mobile/src/app/settings/inventory/fields/index.tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/fields/new.tsx](../../../apps/mobile/src/app/settings/inventory/fields/new.tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/index.tsx](../../../apps/mobile/src/app/settings/inventory/index.tsx) | Mobile | AppServicesContext, ScopedSettingsScreens |
| [settings/inventory/notification-editor.tsx](../../../apps/mobile/src/app/settings/inventory/notification-editor.tsx) | Mobile | SettingsList, NotificationSettingsDestination |
| [settings/inventory/notifications.tsx](../../../apps/mobile/src/app/settings/inventory/notifications.tsx) | Mobile | NotificationSettingsDestination, AppServicesContext, MobileServerStateProvider, SettingsScreenState, NotificationSettingsScreen, SettingsList |
| [settings/inventory/tags/[resourceId].tsx](../../../apps/mobile/src/app/settings/inventory/tags/[resourceId].tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/tags/index.tsx](../../../apps/mobile/src/app/settings/inventory/tags/index.tsx) | Mobile | CustomizationRoutes |
| [settings/inventory/tags/new.tsx](../../../apps/mobile/src/app/settings/inventory/tags/new.tsx) | Mobile | CustomizationRoutes |
| [settings/sharing.tsx](../../../apps/mobile/src/app/settings/sharing.tsx) | Mobile | AppServicesContext, InventorySharingGuard, InventorySharingScreen |
| [settings/voice/[capability].tsx](../../../apps/mobile/src/app/settings/voice/[capability].tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [settings/voice/index.tsx](../../../apps/mobile/src/app/settings/voice/index.tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [settings/voice/profiles/[providerProfileId]/credential.tsx](../../../apps/mobile/src/app/settings/voice/profiles/[providerProfileId]/credential.tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [settings/voice/profiles/[providerProfileId]/index.tsx](../../../apps/mobile/src/app/settings/voice/profiles/[providerProfileId]/index.tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [settings/voice/profiles/[providerProfileId]/prompt.tsx](../../../apps/mobile/src/app/settings/voice/profiles/[providerProfileId]/prompt.tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [settings/voice/profiles/add.tsx](../../../apps/mobile/src/app/settings/voice/profiles/add.tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [settings/voice/profiles/index.tsx](../../../apps/mobile/src/app/settings/voice/profiles/index.tsx) | Mobile | AppServicesContext, VoiceAdminGuard, VoiceSettingsScreens |
| [tenant-switcher.tsx](../../../apps/mobile/src/app/tenant-switcher.tsx) | Mobile | AppServicesContext, TenantSwitcherSheetScreen |
| [voice.tsx](../../../apps/mobile/src/app/voice.tsx) | Mobile | VoiceSessionSheetScreen |
| [+layout.svelte](../../../apps/web/src/routes/+layout.svelte) | Web | index.js |
| [+page.svelte](../../../apps/web/src/routes/+page.svelte) | Web | AuthSignInScreen.svelte, InventoryWorkspaceApp.svelte, index.js |
| [[...workspace]/+page.svelte](../../../apps/web/src/routes/[...workspace]/+page.svelte) | Web | Route delegation / shared shell |
| [callback/+page.svelte](../../../apps/web/src/routes/callback/+page.svelte) | Web | index.js, AuthSurface.svelte |
| [invitations/accept/+page.svelte](../../../apps/web/src/routes/invitations/accept/+page.svelte) | Web | InvitationAcceptSurface.svelte |

## Shared UI sources

| Source | Role |
| --- | --- |
| [apps/mobile/src/ui/components/AppKeyboardAccessory.ios.tsx](../../../apps/mobile/src/ui/components/AppKeyboardAccessory.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AppKeyboardAccessory.tsx](../../../apps/mobile/src/ui/components/AppKeyboardAccessory.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AppKeyboardProvider.ios.tsx](../../../apps/mobile/src/ui/components/AppKeyboardProvider.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AppKeyboardProvider.tsx](../../../apps/mobile/src/ui/components/AppKeyboardProvider.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AppSwitchField.tsx](../../../apps/mobile/src/ui/components/AppSwitchField.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AppTextInput.tsx](../../../apps/mobile/src/ui/components/AppTextInput.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetCard.tsx](../../../apps/mobile/src/ui/components/AssetCard.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetContainedWorkspace.tsx](../../../apps/mobile/src/ui/components/AssetContainedWorkspace.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetDetailIdentitySection.tsx](../../../apps/mobile/src/ui/components/AssetDetailIdentitySection.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetDetailPhotoGallery.tsx](../../../apps/mobile/src/ui/components/AssetDetailPhotoGallery.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetDetailView.tsx](../../../apps/mobile/src/ui/components/AssetDetailView.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetExpirationEditor.tsx](../../../apps/mobile/src/ui/components/AssetExpirationEditor.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetExpirationStatus.tsx](../../../apps/mobile/src/ui/components/AssetExpirationStatus.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/AssetTagChips.tsx](../../../apps/mobile/src/ui/components/AssetTagChips.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/BrandMark.tsx](../../../apps/mobile/src/ui/components/BrandMark.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/CustomizationEditorFields.tsx](../../../apps/mobile/src/ui/components/CustomizationEditorFields.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/CustomizationLifecycleSection.tsx](../../../apps/mobile/src/ui/components/CustomizationLifecycleSection.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/ExpirationField.tsx](../../../apps/mobile/src/ui/components/ExpirationField.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/ExpirationReminderEditor.tsx](../../../apps/mobile/src/ui/components/ExpirationReminderEditor.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/FullScreenPhotoViewer.tsx](../../../apps/mobile/src/ui/components/FullScreenPhotoViewer.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/FullSpectrumTagColorPicker.tsx](../../../apps/mobile/src/ui/components/FullSpectrumTagColorPicker.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/IdentityIcon.tsx](../../../apps/mobile/src/ui/components/IdentityIcon.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeActionMenu.android.tsx](../../../apps/mobile/src/ui/components/NativeActionMenu.android.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeActionMenu.ios.tsx](../../../apps/mobile/src/ui/components/NativeActionMenu.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeActionMenu.tsx](../../../apps/mobile/src/ui/components/NativeActionMenu.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeChoicePicker.ios.tsx](../../../apps/mobile/src/ui/components/NativeChoicePicker.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeChoicePicker.tsx](../../../apps/mobile/src/ui/components/NativeChoicePicker.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeHeaderActions.android.tsx](../../../apps/mobile/src/ui/components/NativeHeaderActions.android.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeHeaderActions.ios.tsx](../../../apps/mobile/src/ui/components/NativeHeaderActions.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeHeaderActions.tsx](../../../apps/mobile/src/ui/components/NativeHeaderActions.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeNavigationSearch.tsx](../../../apps/mobile/src/ui/components/NativeNavigationSearch.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeRefinementButton.android.tsx](../../../apps/mobile/src/ui/components/NativeRefinementButton.android.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeRefinementButton.ios.tsx](../../../apps/mobile/src/ui/components/NativeRefinementButton.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeRefinementButton.tsx](../../../apps/mobile/src/ui/components/NativeRefinementButton.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeSegmentedControl.tsx](../../../apps/mobile/src/ui/components/NativeSegmentedControl.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeSheetActions.android.tsx](../../../apps/mobile/src/ui/components/NativeSheetActions.android.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeSheetActions.ios.tsx](../../../apps/mobile/src/ui/components/NativeSheetActions.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeSheetActions.tsx](../../../apps/mobile/src/ui/components/NativeSheetActions.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeTagColorPicker.ios.tsx](../../../apps/mobile/src/ui/components/NativeTagColorPicker.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NativeTagColorPicker.tsx](../../../apps/mobile/src/ui/components/NativeTagColorPicker.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/NotificationBell.tsx](../../../apps/mobile/src/ui/components/NotificationBell.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/ReminderTimingEditor.tsx](../../../apps/mobile/src/ui/components/ReminderTimingEditor.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/SelectionRow.tsx](../../../apps/mobile/src/ui/components/SelectionRow.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/SettingsSegmentedControl.tsx](../../../apps/mobile/src/ui/components/SettingsSegmentedControl.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/TagColorPicker.tsx](../../../apps/mobile/src/ui/components/TagColorPicker.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/TimeZonePicker.tsx](../../../apps/mobile/src/ui/components/TimeZonePicker.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/components/VoiceLevelMeter.tsx](../../../apps/mobile/src/ui/components/VoiceLevelMeter.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationDateRange.tsx](../../../apps/mobile/src/ui/expiration/ExpirationDateRange.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationFilterHeader.ios.tsx](../../../apps/mobile/src/ui/expiration/ExpirationFilterHeader.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationFilterHeader.tsx](../../../apps/mobile/src/ui/expiration/ExpirationFilterHeader.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationFiltersScreen.tsx](../../../apps/mobile/src/ui/expiration/ExpirationFiltersScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationHomeEntry.tsx](../../../apps/mobile/src/ui/expiration/ExpirationHomeEntry.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationHomeSection.tsx](../../../apps/mobile/src/ui/expiration/ExpirationHomeSection.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/expiration/ExpirationWorkspaceScreen.tsx](../../../apps/mobile/src/ui/expiration/ExpirationWorkspaceScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/feedback/AppFeedback.tsx](../../../apps/mobile/src/ui/feedback/AppFeedback.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/AppServicesContext.tsx](../../../apps/mobile/src/ui/navigation/AppServicesContext.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/CustomizationRoutes.tsx](../../../apps/mobile/src/ui/navigation/CustomizationRoutes.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/InventoryInvitationLinkContext.tsx](../../../apps/mobile/src/ui/navigation/InventoryInvitationLinkContext.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/InventorySharingGuard.tsx](../../../apps/mobile/src/ui/navigation/InventorySharingGuard.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/MobileServerStateProvider.tsx](../../../apps/mobile/src/ui/navigation/MobileServerStateProvider.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/NotificationHomeEntry.tsx](../../../apps/mobile/src/ui/navigation/NotificationHomeEntry.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/PushNotificationNavigation.tsx](../../../apps/mobile/src/ui/navigation/PushNotificationNavigation.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/PushRegistrationLifecycle.tsx](../../../apps/mobile/src/ui/navigation/PushRegistrationLifecycle.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/VoiceAdminGuard.tsx](../../../apps/mobile/src/ui/navigation/VoiceAdminGuard.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/VoiceBottomAccessory.tsx](../../../apps/mobile/src/ui/navigation/VoiceBottomAccessory.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/VoiceConversationReturn.tsx](../../../apps/mobile/src/ui/navigation/VoiceConversationReturn.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/navigation/VoiceInteractionStateContext.tsx](../../../apps/mobile/src/ui/navigation/VoiceInteractionStateContext.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AddAssetScreen.tsx](../../../apps/mobile/src/ui/screens/AddAssetScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetCheckoutHistoryScreen.tsx](../../../apps/mobile/src/ui/screens/AssetCheckoutHistoryScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetCheckoutHistorySheet.tsx](../../../apps/mobile/src/ui/screens/AssetCheckoutHistorySheet.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetDetailRouteErrorState.tsx](../../../apps/mobile/src/ui/screens/AssetDetailRouteErrorState.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetDetailRouteScreen.tsx](../../../apps/mobile/src/ui/screens/AssetDetailRouteScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetDetailSheets.tsx](../../../apps/mobile/src/ui/screens/AssetDetailSheets.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetHeaderOverflow.ios.tsx](../../../apps/mobile/src/ui/screens/AssetHeaderOverflow.ios.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetHeaderOverflow.tsx](../../../apps/mobile/src/ui/screens/AssetHeaderOverflow.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetHistoryDetailRouteScreen.tsx](../../../apps/mobile/src/ui/screens/AssetHistoryDetailRouteScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetHistoryRouteScreen.tsx](../../../apps/mobile/src/ui/screens/AssetHistoryRouteScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetNativeActionSheetScreens.tsx](../../../apps/mobile/src/ui/screens/AssetNativeActionSheetScreens.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetOverflowMenu.tsx](../../../apps/mobile/src/ui/screens/AssetOverflowMenu.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/AssetPhotoViewerSheet.tsx](../../../apps/mobile/src/ui/screens/AssetPhotoViewerSheet.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/BrowseAddHeader.tsx](../../../apps/mobile/src/ui/screens/BrowseAddHeader.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/BrowseFiltersScreen.tsx](../../../apps/mobile/src/ui/screens/BrowseFiltersScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/BrowseHeader.tsx](../../../apps/mobile/src/ui/screens/BrowseHeader.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/BrowsePlaceRow.tsx](../../../apps/mobile/src/ui/screens/BrowsePlaceRow.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/BrowseResultStates.tsx](../../../apps/mobile/src/ui/screens/BrowseResultStates.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/BrowseSurfaceControl.tsx](../../../apps/mobile/src/ui/screens/BrowseSurfaceControl.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/CustomizationCollectionScreen.tsx](../../../apps/mobile/src/ui/screens/CustomizationCollectionScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/CustomizationEditorScreen.tsx](../../../apps/mobile/src/ui/screens/CustomizationEditorScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/HomeNavigationHeader.tsx](../../../apps/mobile/src/ui/screens/HomeNavigationHeader.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/HomeScreen.tsx](../../../apps/mobile/src/ui/screens/HomeScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/InventoryAssetsRouteScreen.tsx](../../../apps/mobile/src/ui/screens/InventoryAssetsRouteScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/InventoryInvitationScreen.tsx](../../../apps/mobile/src/ui/screens/InventoryInvitationScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/InventoryMapScreen.tsx](../../../apps/mobile/src/ui/screens/InventoryMapScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/InventorySharingScreen.tsx](../../../apps/mobile/src/ui/screens/InventorySharingScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/LocationAssetsRouteScreen.tsx](../../../apps/mobile/src/ui/screens/LocationAssetsRouteScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/LocationsScreen.tsx](../../../apps/mobile/src/ui/screens/LocationsScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/NotificationInboxScreen.tsx](../../../apps/mobile/src/ui/screens/NotificationInboxScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/NotificationSettingsScreen.tsx](../../../apps/mobile/src/ui/screens/NotificationSettingsScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/OnboardingScreen.tsx](../../../apps/mobile/src/ui/screens/OnboardingScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/ProviderProfileEditorScreens.tsx](../../../apps/mobile/src/ui/screens/ProviderProfileEditorScreens.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/ProviderProfileScreens.tsx](../../../apps/mobile/src/ui/screens/ProviderProfileScreens.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/ProviderSettingsSupport.tsx](../../../apps/mobile/src/ui/screens/ProviderSettingsSupport.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/ScopedSettingsScreens.tsx](../../../apps/mobile/src/ui/screens/ScopedSettingsScreens.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/SearchScreen.tsx](../../../apps/mobile/src/ui/screens/SearchScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/SettingsDetailScreens.tsx](../../../apps/mobile/src/ui/screens/SettingsDetailScreens.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/SettingsList.tsx](../../../apps/mobile/src/ui/screens/SettingsList.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/SettingsRefreshNotice.tsx](../../../apps/mobile/src/ui/screens/SettingsRefreshNotice.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/SettingsScreen.tsx](../../../apps/mobile/src/ui/screens/SettingsScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/TenantSwitcherSheetScreen.tsx](../../../apps/mobile/src/ui/screens/TenantSwitcherSheetScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoiceConversationComposer.tsx](../../../apps/mobile/src/ui/screens/VoiceConversationComposer.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoiceConversationExchange.tsx](../../../apps/mobile/src/ui/screens/VoiceConversationExchange.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoicePlanHistorySummary.tsx](../../../apps/mobile/src/ui/screens/VoicePlanHistorySummary.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoicePlanPhotoDrafts.tsx](../../../apps/mobile/src/ui/screens/VoicePlanPhotoDrafts.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoicePlanProgress.tsx](../../../apps/mobile/src/ui/screens/VoicePlanProgress.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoiceResponseEntityText.tsx](../../../apps/mobile/src/ui/screens/VoiceResponseEntityText.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoiceSessionSheetScreen.tsx](../../../apps/mobile/src/ui/screens/VoiceSessionSheetScreen.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/screens/VoiceSettingsScreens.tsx](../../../apps/mobile/src/ui/screens/VoiceSettingsScreens.tsx) | Mobile shared implementation |
| [apps/mobile/src/ui/theme/AppearanceContext.tsx](../../../apps/mobile/src/ui/theme/AppearanceContext.tsx) | Mobile shared implementation |
| [apps/web/src/lib/components/auth/AuthBrand.svelte](../../../apps/web/src/lib/components/auth/AuthBrand.svelte) | Web component |
| [apps/web/src/lib/components/auth/AuthSignInScreen.svelte](../../../apps/web/src/lib/components/auth/AuthSignInScreen.svelte) | Web component |
| [apps/web/src/lib/components/auth/AuthSurface.svelte](../../../apps/web/src/lib/components/auth/AuthSurface.svelte) | Web component |
| [apps/web/src/lib/components/invitations/InvitationAcceptSurface.svelte](../../../apps/web/src/lib/components/invitations/InvitationAcceptSurface.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert/alert-action.svelte](../../../apps/web/src/lib/components/ui/alert/alert-action.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert/alert-description.svelte](../../../apps/web/src/lib/components/ui/alert/alert-description.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert/alert-title.svelte](../../../apps/web/src/lib/components/ui/alert/alert-title.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert/alert.svelte](../../../apps/web/src/lib/components/ui/alert/alert.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-action.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-action.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-cancel.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-cancel.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-content.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-description.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-description.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-footer.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-footer.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-header.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-header.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-media.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-media.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-overlay.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-overlay.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-portal.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-portal.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-title.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-title.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog-trigger.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/alert-dialog/alert-dialog.svelte](../../../apps/web/src/lib/components/ui/alert-dialog/alert-dialog.svelte) | Web component |
| [apps/web/src/lib/components/ui/badge/badge.svelte](../../../apps/web/src/lib/components/ui/badge/badge.svelte) | Web component |
| [apps/web/src/lib/components/ui/button/busy-button-content.svelte](../../../apps/web/src/lib/components/ui/button/busy-button-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/button/button.svelte](../../../apps/web/src/lib/components/ui/button/button.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card-action.svelte](../../../apps/web/src/lib/components/ui/card/card-action.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card-content.svelte](../../../apps/web/src/lib/components/ui/card/card-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card-description.svelte](../../../apps/web/src/lib/components/ui/card/card-description.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card-footer.svelte](../../../apps/web/src/lib/components/ui/card/card-footer.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card-header.svelte](../../../apps/web/src/lib/components/ui/card/card-header.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card-title.svelte](../../../apps/web/src/lib/components/ui/card/card-title.svelte) | Web component |
| [apps/web/src/lib/components/ui/card/card.svelte](../../../apps/web/src/lib/components/ui/card/card.svelte) | Web component |
| [apps/web/src/lib/components/ui/checkbox/checkbox.svelte](../../../apps/web/src/lib/components/ui/checkbox/checkbox.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-close.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-close.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-content.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-description.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-description.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-footer.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-footer.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-header.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-header.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-overlay.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-overlay.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-portal.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-portal.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-title.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-title.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog-trigger.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/dialog/dialog.svelte](../../../apps/web/src/lib/components/ui/dialog/dialog.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-checkbox-group.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-checkbox-group.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-checkbox-item.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-checkbox-item.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-content.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-group-heading.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-group-heading.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-group.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-group.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-item.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-item.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-label.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-label.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-portal.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-portal.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-radio-group.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-radio-group.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-radio-item.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-radio-item.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-separator.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-separator.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-shortcut.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-shortcut.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-sub-content.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-sub-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-sub-trigger.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-sub-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-sub.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-sub.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-trigger.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu.svelte](../../../apps/web/src/lib/components/ui/dropdown-menu/dropdown-menu.svelte) | Web component |
| [apps/web/src/lib/components/ui/input/input.svelte](../../../apps/web/src/lib/components/ui/input/input.svelte) | Web component |
| [apps/web/src/lib/components/ui/label/label.svelte](../../../apps/web/src/lib/components/ui/label/label.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-close.svelte](../../../apps/web/src/lib/components/ui/popover/popover-close.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-content.svelte](../../../apps/web/src/lib/components/ui/popover/popover-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-description.svelte](../../../apps/web/src/lib/components/ui/popover/popover-description.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-header.svelte](../../../apps/web/src/lib/components/ui/popover/popover-header.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-portal.svelte](../../../apps/web/src/lib/components/ui/popover/popover-portal.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-title.svelte](../../../apps/web/src/lib/components/ui/popover/popover-title.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover-trigger.svelte](../../../apps/web/src/lib/components/ui/popover/popover-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/popover/popover.svelte](../../../apps/web/src/lib/components/ui/popover/popover.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-content.svelte](../../../apps/web/src/lib/components/ui/select/select-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-group-heading.svelte](../../../apps/web/src/lib/components/ui/select/select-group-heading.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-group.svelte](../../../apps/web/src/lib/components/ui/select/select-group.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-item.svelte](../../../apps/web/src/lib/components/ui/select/select-item.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-label.svelte](../../../apps/web/src/lib/components/ui/select/select-label.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-portal.svelte](../../../apps/web/src/lib/components/ui/select/select-portal.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-scroll-down-button.svelte](../../../apps/web/src/lib/components/ui/select/select-scroll-down-button.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-scroll-up-button.svelte](../../../apps/web/src/lib/components/ui/select/select-scroll-up-button.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-separator.svelte](../../../apps/web/src/lib/components/ui/select/select-separator.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select-trigger.svelte](../../../apps/web/src/lib/components/ui/select/select-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/select/select.svelte](../../../apps/web/src/lib/components/ui/select/select.svelte) | Web component |
| [apps/web/src/lib/components/ui/separator/separator.svelte](../../../apps/web/src/lib/components/ui/separator/separator.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-close.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-close.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-content.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-description.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-description.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-footer.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-footer.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-header.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-header.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-overlay.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-overlay.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-portal.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-portal.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-title.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-title.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet-trigger.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/sheet/sheet.svelte](../../../apps/web/src/lib/components/ui/sheet/sheet.svelte) | Web component |
| [apps/web/src/lib/components/ui/sonner/sonner.svelte](../../../apps/web/src/lib/components/ui/sonner/sonner.svelte) | Web component |
| [apps/web/src/lib/components/ui/step-progress/step-progress.svelte](../../../apps/web/src/lib/components/ui/step-progress/step-progress.svelte) | Web component |
| [apps/web/src/lib/components/ui/table/table-body.svelte](../../../apps/web/src/lib/components/ui/table/table-body.svelte) | Web component |
| [apps/web/src/lib/components/ui/table/table-cell.svelte](../../../apps/web/src/lib/components/ui/table/table-cell.svelte) | Web component |
| [apps/web/src/lib/components/ui/table/table-head.svelte](../../../apps/web/src/lib/components/ui/table/table-head.svelte) | Web component |
| [apps/web/src/lib/components/ui/table/table-header.svelte](../../../apps/web/src/lib/components/ui/table/table-header.svelte) | Web component |
| [apps/web/src/lib/components/ui/table/table-row.svelte](../../../apps/web/src/lib/components/ui/table/table-row.svelte) | Web component |
| [apps/web/src/lib/components/ui/table/table.svelte](../../../apps/web/src/lib/components/ui/table/table.svelte) | Web component |
| [apps/web/src/lib/components/ui/tabs/tabs-content.svelte](../../../apps/web/src/lib/components/ui/tabs/tabs-content.svelte) | Web component |
| [apps/web/src/lib/components/ui/tabs/tabs-list.svelte](../../../apps/web/src/lib/components/ui/tabs/tabs-list.svelte) | Web component |
| [apps/web/src/lib/components/ui/tabs/tabs-trigger.svelte](../../../apps/web/src/lib/components/ui/tabs/tabs-trigger.svelte) | Web component |
| [apps/web/src/lib/components/ui/tabs/tabs.svelte](../../../apps/web/src/lib/components/ui/tabs/tabs.svelte) | Web component |
| [apps/web/src/lib/components/ui/textarea/textarea.svelte](../../../apps/web/src/lib/components/ui/textarea/textarea.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AccountMenu.svelte](../../../apps/web/src/lib/components/workspace/AccountMenu.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AddAssetCustomFieldsSection.svelte](../../../apps/web/src/lib/components/workspace/AddAssetCustomFieldsSection.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AddAssetPhotosSection.svelte](../../../apps/web/src/lib/components/workspace/AddAssetPhotosSection.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AddAssetTray.svelte](../../../apps/web/src/lib/components/workspace/AddAssetTray.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetDetail.svelte](../../../apps/web/src/lib/components/workspace/AssetDetail.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetDetailActionPanel.svelte](../../../apps/web/src/lib/components/workspace/AssetDetailActionPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetDetailHero.svelte](../../../apps/web/src/lib/components/workspace/AssetDetailHero.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetExpirationLabel.svelte](../../../apps/web/src/lib/components/workspace/AssetExpirationLabel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetFilesSection.svelte](../../../apps/web/src/lib/components/workspace/AssetFilesSection.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetLocationTrail.svelte](../../../apps/web/src/lib/components/workspace/AssetLocationTrail.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetTagChips.svelte](../../../apps/web/src/lib/components/workspace/AssetTagChips.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetTagSelector.svelte](../../../apps/web/src/lib/components/workspace/AssetTagSelector.svelte) | Web component |
| [apps/web/src/lib/components/workspace/AssetThumb.svelte](../../../apps/web/src/lib/components/workspace/AssetThumb.svelte) | Web component |
| [apps/web/src/lib/components/workspace/BinaryOption.svelte](../../../apps/web/src/lib/components/workspace/BinaryOption.svelte) | Web component |
| [apps/web/src/lib/components/workspace/BrowsePanel.svelte](../../../apps/web/src/lib/components/workspace/BrowsePanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/CheckoutBadge.svelte](../../../apps/web/src/lib/components/workspace/CheckoutBadge.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ChoiceGrid.svelte](../../../apps/web/src/lib/components/workspace/ChoiceGrid.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ContainedAssetWorkspace.svelte](../../../apps/web/src/lib/components/workspace/ContainedAssetWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/CustomFieldControls.svelte](../../../apps/web/src/lib/components/workspace/CustomFieldControls.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ExpirationField.svelte](../../../apps/web/src/lib/components/workspace/ExpirationField.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ExpirationRefresh.svelte](../../../apps/web/src/lib/components/workspace/ExpirationRefresh.svelte) | Web component |
| [apps/web/src/lib/components/workspace/HomeWorkspace.svelte](../../../apps/web/src/lib/components/workspace/HomeWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportCountGrid.svelte](../../../apps/web/src/lib/components/workspace/ImportCountGrid.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportFlowStepper.svelte](../../../apps/web/src/lib/components/workspace/ImportFlowStepper.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportJobConfirmationPanel.svelte](../../../apps/web/src/lib/components/workspace/ImportJobConfirmationPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportJobDetailPanel.svelte](../../../apps/web/src/lib/components/workspace/ImportJobDetailPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportJobHistory.svelte](../../../apps/web/src/lib/components/workspace/ImportJobHistory.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportJobRunHandoff.svelte](../../../apps/web/src/lib/components/workspace/ImportJobRunHandoff.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportMessagesList.svelte](../../../apps/web/src/lib/components/workspace/ImportMessagesList.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportPreviewPanel.svelte](../../../apps/web/src/lib/components/workspace/ImportPreviewPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportPreviewSamples.svelte](../../../apps/web/src/lib/components/workspace/ImportPreviewSamples.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportSourceChoiceStep.svelte](../../../apps/web/src/lib/components/workspace/ImportSourceChoiceStep.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ImportSourceSetup.svelte](../../../apps/web/src/lib/components/workspace/ImportSourceSetup.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryAccessInvitationActionPanel.svelte](../../../apps/web/src/lib/components/workspace/InventoryAccessInvitationActionPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryAccessManager.svelte](../../../apps/web/src/lib/components/workspace/InventoryAccessManager.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryAuditPanel.svelte](../../../apps/web/src/lib/components/workspace/InventoryAuditPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryCustomizationArchivePanel.svelte](../../../apps/web/src/lib/components/workspace/InventoryCustomizationArchivePanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryCustomizationManager.svelte](../../../apps/web/src/lib/components/workspace/InventoryCustomizationManager.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryImportWorkspace.svelte](../../../apps/web/src/lib/components/workspace/InventoryImportWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventorySettings.svelte](../../../apps/web/src/lib/components/workspace/InventorySettings.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryWorkspaceApp.svelte](../../../apps/web/src/lib/components/workspace/InventoryWorkspaceApp.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryWorkspaceChrome.svelte](../../../apps/web/src/lib/components/workspace/InventoryWorkspaceChrome.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryWorkspaceOverlays.svelte](../../../apps/web/src/lib/components/workspace/InventoryWorkspaceOverlays.svelte) | Web component |
| [apps/web/src/lib/components/workspace/InventoryWorkspaceRouteContent.svelte](../../../apps/web/src/lib/components/workspace/InventoryWorkspaceRouteContent.svelte) | Web component |
| [apps/web/src/lib/components/workspace/KindIcon.svelte](../../../apps/web/src/lib/components/workspace/KindIcon.svelte) | Web component |
| [apps/web/src/lib/components/workspace/LocationView.svelte](../../../apps/web/src/lib/components/workspace/LocationView.svelte) | Web component |
| [apps/web/src/lib/components/workspace/MobileNav.svelte](../../../apps/web/src/lib/components/workspace/MobileNav.svelte) | Web component |
| [apps/web/src/lib/components/workspace/NotificationBell.svelte](../../../apps/web/src/lib/components/workspace/NotificationBell.svelte) | Web component |
| [apps/web/src/lib/components/workspace/NotificationInbox.svelte](../../../apps/web/src/lib/components/workspace/NotificationInbox.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ParentTargetButton.svelte](../../../apps/web/src/lib/components/workspace/ParentTargetButton.svelte) | Web component |
| [apps/web/src/lib/components/workspace/ParentTargetPicker.svelte](../../../apps/web/src/lib/components/workspace/ParentTargetPicker.svelte) | Web component |
| [apps/web/src/lib/components/workspace/SearchSuggestions.svelte](../../../apps/web/src/lib/components/workspace/SearchSuggestions.svelte) | Web component |
| [apps/web/src/lib/components/workspace/SegmentedControl.svelte](../../../apps/web/src/lib/components/workspace/SegmentedControl.svelte) | Web component |
| [apps/web/src/lib/components/workspace/SideNav.svelte](../../../apps/web/src/lib/components/workspace/SideNav.svelte) | Web component |
| [apps/web/src/lib/components/workspace/TopHeader.svelte](../../../apps/web/src/lib/components/workspace/TopHeader.svelte) | Web component |
| [apps/web/src/lib/components/workspace/WorkspaceAddMenu.svelte](../../../apps/web/src/lib/components/workspace/WorkspaceAddMenu.svelte) | Web component |
| [apps/web/src/lib/components/workspace/WorkspaceContextSwitcher.svelte](../../../apps/web/src/lib/components/workspace/WorkspaceContextSwitcher.svelte) | Web component |
| [apps/web/src/lib/components/workspace/WorkspaceSetupPanel.svelte](../../../apps/web/src/lib/components/workspace/WorkspaceSetupPanel.svelte) | Web component |
| [apps/web/src/lib/components/workspace/action-surface/WorkspaceConfirmationDialog.svelte](../../../apps/web/src/lib/components/workspace/action-surface/WorkspaceConfirmationDialog.svelte) | Web component |
| [apps/web/src/lib/components/workspace/action-surface/WorkspaceTaskSheet.svelte](../../../apps/web/src/lib/components/workspace/action-surface/WorkspaceTaskSheet.svelte) | Web component |
| [apps/web/src/lib/components/workspace/expiration/ExpirationBrowseEntry.svelte](../../../apps/web/src/lib/components/workspace/expiration/ExpirationBrowseEntry.svelte) | Web component |
| [apps/web/src/lib/components/workspace/expiration/ExpirationChoice.svelte](../../../apps/web/src/lib/components/workspace/expiration/ExpirationChoice.svelte) | Web component |
| [apps/web/src/lib/components/workspace/expiration/ExpirationFilters.svelte](../../../apps/web/src/lib/components/workspace/expiration/ExpirationFilters.svelte) | Web component |
| [apps/web/src/lib/components/workspace/expiration/ExpirationHome.svelte](../../../apps/web/src/lib/components/workspace/expiration/ExpirationHome.svelte) | Web component |
| [apps/web/src/lib/components/workspace/expiration/ExpirationRows.svelte](../../../apps/web/src/lib/components/workspace/expiration/ExpirationRows.svelte) | Web component |
| [apps/web/src/lib/components/workspace/expiration/ExpirationWorkspace.svelte](../../../apps/web/src/lib/components/workspace/expiration/ExpirationWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/AssetTypeSettingsManager.svelte](../../../apps/web/src/lib/components/workspace/settings/AssetTypeSettingsManager.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/ExpirationReminderEditor.svelte](../../../apps/web/src/lib/components/workspace/settings/ExpirationReminderEditor.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/FieldSettingsManager.svelte](../../../apps/web/src/lib/components/workspace/settings/FieldSettingsManager.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/NotificationSettings.svelte](../../../apps/web/src/lib/components/workspace/settings/NotificationSettings.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/SettingsCollectionState.svelte](../../../apps/web/src/lib/components/workspace/settings/SettingsCollectionState.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/SettingsDestinationList.svelte](../../../apps/web/src/lib/components/workspace/settings/SettingsDestinationList.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/SettingsWorkspace.svelte](../../../apps/web/src/lib/components/workspace/settings/SettingsWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/TagSettingsManager.svelte](../../../apps/web/src/lib/components/workspace/settings/TagSettingsManager.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/CaseEditor.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/CaseEditor.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/CaseExpectations.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/CaseExpectations.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/CaseFixtures.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/CaseFixtures.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/CaseSummary.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/CaseSummary.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/CaseWorkspace.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/CaseWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/ConversationWorkspace.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/ConversationWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/RunActivation.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/RunActivation.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/RunComparison.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/RunComparison.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/RunDetails.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/RunDetails.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/RunResult.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/RunResult.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/RunSetup.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/RunSetup.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/RunWorkspace.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/RunWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/ValidationMessage.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/ValidationMessage.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/WorkflowEditor.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/WorkflowEditor.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/WorkflowSelect.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/WorkflowSelect.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/conversations/WorkflowWorkspace.svelte](../../../apps/web/src/lib/components/workspace/settings/conversations/WorkflowWorkspace.svelte) | Web component |
| [apps/web/src/lib/components/workspace/settings/settings-management.css](../../../apps/web/src/lib/components/workspace/settings/settings-management.css) | Web component |
| [docs/src/styles/brand.css](../../../docs/src/styles/brand.css) | Documentation theme |

## Documentation pages

Page inventory only; documentation shell and theme were source-reviewed.

- [architecture](../../../docs/src/content/docs/architecture.md)
- [concepts](../../../docs/src/content/docs/concepts.md)
- [configuration](../../../docs/src/content/docs/configuration.md)
- [dex-users](../../../docs/src/content/docs/dex-users.md)
- [expiration](../../../docs/src/content/docs/expiration.md)
- [first-inventory](../../../docs/src/content/docs/first-inventory.md)
- [index](../../../docs/src/content/docs/index.mdx)
- [local-development](../../../docs/src/content/docs/local-development.md)
- [product](../../../docs/src/content/docs/product.md)
- [security](../../../docs/src/content/docs/security.md)
- [self-host-operations](../../../docs/src/content/docs/self-host-operations.md)
- [self-hosting](../../../docs/src/content/docs/self-hosting.md)
- [specs-and-process](../../../docs/src/content/docs/specs-and-process.md)
- [testflight](../../../docs/src/content/docs/testflight.md)
