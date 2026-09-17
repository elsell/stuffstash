# Push permission and device setup — all 24 source axes

S116 at6a4f9221 plus M205. Reviewed NotificationSettingsScreen, SettingsList,
PushSession, PushSetup and ExpoPushDevice. In-app reminders, saved push preference
and this device's permission are separate states; setup success is not a permanent
claim that OS permission remains granted.

[Apple's notification permission guidance](https://developer.apple.com/documentation/UserNotifications/asking-permission-to-use-notifications)
supports asking in context so people understand the purpose. This screen requests
permission after an explicit enable action and offers recovery after denial.

| Axis | Conclusion and remaining acceptance |
| --- | --- |
| Task | Enable inventory alerts and register this installation. Saved preference uses a switch; per-device setup is a command. |
| Navigation | Ordinary reminder settings route, with an explicit external Settings action after denial/success. Native app-to-Settings return remains unverified. |
| Selection | Switch reflects saved preference, not an inferred current OS permission. Denial leaves preference unchanged. |
| Modality | OS owns its permission prompt. No additional confirmation before this requested setup. |
| Layout | Grouped scrolling settings, automatic content insets, no overlay action footer. Native safe-area/scroll acceptance remains open. |
| Adaptation | Shared vertical rows; no fixed-width permission dialog. Long content and tablet/enlarged-text checks remain pending. |
| Typography | Shared settings labels/footers; recovery copy remains inline. Native wrapping/legibility needs acceptance. |
| Appearance | Native switch with semantic settings text. System owns permission and Settings appearance; light/dark parity unverified. |
| Localization | English instructions; no dates/numbers in setup. Long translations and RTL remain unverified. |
| Imagery | Text actions; no remote images or meaningful icon-only controls. |
| Targets | Shared full-row action targets and native switch. Actual device hit regions remain unverified. |
| Gestures | Explicit switch/action rows; no gesture-only recovery. Pull refresh is gesture-owned, not background-loading-owned. |
| Keyboard | No push-entry keyboard task. OS transitions with an existing editor keyboard remain native acceptance work. |
| Accessibility | Switch is labeled. M205 changes the successful action's accessible name to Open device settings, matching what it now does. Native speech/focus remains pending. |
| Motion | No custom animation. System transition and reduced-motion behavior remain unverified. |
| Content | Explains independent in-app reminders and server delivery support; setup feedback is transient. No promise of physical delivery from registration alone. |
| Search | No discovery task within this short permission section. |
| Loading | Shared pending guard disables competing settings changes; Settings launch has its own duplicate-suppression token. |
| Recovery | Denial offers Settings. Launch rejection gives manual instructions and retry; transient failure preserves preferences. M199 removes editable cache after access denial. |
| Editing | Disable only saves pushEnabled=false, preserving reminder policies. Enable requests permission/registers before saving. No optimistic false success. |
| Privacy | Permission is requested explicitly; token/journal/repository stay behind ports. No tokens in UI feedback. Parent authorization remains required. |
| Notifications | Platform permission is authoritative; background invalidates transient outcome. Physical permission changes, token rotation and APNs/FCM delivery are not certified by mounted tests. |
| Media | No camera/library/audio/file request. |
| Lifecycle | Setup cancellation guards subsequent writes; feedback is visit/background-owned. Settings launch cannot publish old failure into a new visit. Physical OS lifecycle remains open. |

M205 extends device-Settings recovery to both enabled/denied paths across current,
returned and background contexts. Three enabled-path RED cases preceded the label
fix. All13 related tests across two files, TypeScript and structural checks pass
on paul; setup-count and preference-write assertions guard unintended repetition.
Code critic found no confirmed blocker. These checks use a native-launch fake;
they do not prove actual device Settings presentation or push delivery.
