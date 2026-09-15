# Provider credential and prompt editors — all 24 axes

Source checkpoint: da14195c. Surfaces R051/S125 (credential) and R053/S126
(prompt), including their route wrappers and shared provider query/guard. This is
source inspection; no current native editor capture has been verified.

The task is a deliberate replacement of a secret or hidden prompt, not an immediate
preference toggle. A dedicated editor is justified by secure entry, multiline text,
and an explicit server mutation. Existing hidden values must not be fetched merely
to populate these fields. The native navigation stack and text input are appropriate;
the custom command chrome and editing/recovery semantics need correction.

Apple recommends reducing entry errors and using secure text fields where needed
([Entering data](https://developer.apple.com/design/human-interface-guidelines/entering-data)).
It recommends labels and field-adjacent actionable errors
([Writing](https://developer.apple.com/design/human-interface-guidelines/writing)).
Critical navigation and completion commands belong in familiar distinct toolbar
sections ([Toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars)).
The choice of one native Save command with normal Back, plus a dirty-draft decision,
is this project's application of that guidance, not an Apple requirement that all
forms use the same location or presentation.

## Source and interaction decisions

| Axis | Evidence, decision and remaining acceptance |
| --- | --- |
| Task | Credential replaces one provider secret (or chooses server ADC); prompt replaces hidden guidance. Keep explicit Save. A destination is earned by sensitive/multiline editing, not a flat option choice. |
| Navigation | Routes use native stack Back and onSaved/onCancel both call router.back. M107: no draft-aware removal guard. M105 now rejects departed save navigation. Check Back and navigation return natively. |
| Selection | N/A: these editors do not select among competing values. Credential purpose is established by the selected profile; server ADC intentionally has no secret field. |
| Modality | Pushed destinations, not nested sheets. Keep existing destination until evidence warrants modality change. M107 applies to Back and other removal paths. |
| Layout | Custom action rows sit after input inside ScrollView; global errors use the shared overlay (M103). M106: use native completion controls and field-local recovery. Native header, insets, scrolling and keyboard overlap remain unverified. |
| Adaptation | Form grows with available width and has no fixed screen height. Multiline has a minimum height. Phone, iPad, split window and normal-size long text still require runtime checks. |
| Typography | Fixed nominal fonts and some fixed line heights come from Settings styles; input is 17 points. Duplicated route/form headings consume space. Assess normal-size hierarchy before enlarged-text work. No readability pass from source. |
| Appearance | Semantic palette used, but custom primary/secondary button fills bypass the native command adapter. M106. Light/dark and disabled contrast need native captures. |
| Localization | English literals throughout; prompt/credential labels differ by purpose. Long names and RTL remain pending under the broader localization audit. Do not claim locale support here. |
| Imagery | N/A: no content image or essential icon in these forms. Native navigation symbols belong to route controls. |
| Targets | Custom buttons have minimum 48-point height; this is source intent, not measured hit bounds. M106 replaces unnecessary custom commands; native targets and reach with keyboard pending. |
| Gestures | ScrollView supports interactive iOS keyboard dismissal and on-drag elsewhere. M107: back swipe/removal can drop local drafts. Verify keep-editing/discard on actual navigation. |
| Keyboard | AppTextInput is React Native TextInput. Credential disables autocorrection/capitalization and enables secureTextEntry; prompt is multiline. Submission, hardware keyboard, dismissal, focus and viewport avoidance remain runtime work. |
| Accessibility | Fields have labels; Save has busy/disabled state; headings use header role. Global failure announcements depend on AppFeedback. M106 moves failures beside the field with announcement semantics. VoiceOver/TalkBack traversal and error focus are unverified. |
| Motion | No custom form animation. Navigation uses the platform; notices use the shared feedback component. Reduced-motion behavior remains native acceptance, not a new form-specific animation requirement. |
| Content | One replacement value and profile context, not a collection. Existing hidden values are deliberately absent. Remove repetitive setup-only copy where it distracts from replacing an existing profile value. |
| Search | N/A: no search or refinement task in either editor. |
| Loading | Query loading/error lives outside the keyed form. Save locks field and local actions with a synchronous ref until settlement. Background refresh retains available data with SettingsRefreshNotice. M107 must define navigation while a save is pending. |
| Recovery | Initial query and refresh retry exist. M106: empty required values still offer Save, then report command validation globally; server failures also use the overlay. Keep the draft and explain correction/retry beside it. Server ADC is a valid empty-input exception. |
| Editing | Each keyed form owns local text; failed saves retain it. Successful credential saves clear their own secret even after blur (M105). M107: no unsaved-draft decision on Back/Cancel. No persistent secret draft store should be added. |
| Privacy | VoiceAdminGuard requires tenant configure access; repository/API remain the authorization boundary. Secure input is not stored in query data; hidden values are never returned to populate the editor. Existing scope keys isolate forms. Do not claim security certification from a UI review. |
| Notifications | N/A: neither form schedules or receives OS notifications. In-app notices are covered by recovery/lifecycle. |
| Media | N/A: no camera, file, image or audio acquisition in these editors. Voice configuration is not recording. |
| Lifecycle | M105 guards delayed completion by command/resource and focused visit. Mounted tests cover blur/return and replacement-profile secrets. M107 must handle dirty removal. Cold query, device backgrounding and real native return remain pending. |

## Remediation and acceptance

M106: adopt existing native Save controls, show required-input readiness before
submission, retain the server-ADC exception, and put failure/retry next to the
replacement field. Preserve focused success navigation and M105 ownership.

M107: dirty Back/Cancel requires Keep Editing or Discard; pending Save must not
allow ambiguous departure, and authorized successful exit must not prompt to
discard saved work. Keep secrets only in the current form. Delayed confirmations
must not affect a replacement profile or a new visit.

Run normal-size phone/iPad journeys first: empty, filled, ADC, multiline, failed
save, retry, Back/keep/discard, pending mutation, successful return, new profile,
light/dark and keyboard. Retain native screenshots and command bounds. Existing
mounted evidence is useful but does not replace those journeys.
