# Voice setup and readiness — 24 source axes

R049/R050/S124 at684c54d5 plus M180. Inspected route guards/params, VoiceSetupScreen,
VoiceCapabilityScreen, readiness/stage presentation, ProviderSettingsSupport,
settings query state and native picker/action consumers. Credential/prompt content
has its own review; this pass includes their shared loading/error view only.

Apple's [progress guidance](https://developer.apple.com/design/human-interface-guidelines/progress-indicators)
notes that spinners often need no label when the initiating action supplies context.
This is not a universal unlabeled-spinner violation. M180 is a project choice for
shared full-screen states and fixes demonstrably wrong task headings. Apple's
[labels guidance](https://developer.apple.com/design/human-interface-guidelines/labels)
describes labels as context for available actions. Native menus are appropriate
here for a flat service choice; inspecting a profile or editing credentials is a
separate task that justifies navigation. These fit judgments are project analysis.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Overview describes Listen/Understand/Speak readiness and household-wide impact. Stage screen offers a recommended corrective action and current service. |
| Navigation | Native stack routes behind VoiceAdminGuard. Profile details, credential editing and Add are distinct tasks. M105 protects completion notices after departure; native return remains open. |
| Selection | SettingsPickerRow selects a service in place. Current selection is retained even when absent from ordinary candidates; archived choices excluded unless current. Disabled native events are guarded (M54). |
| Modality | No nested selection screen for flat service choice. Profile/credential editing owns separate navigation/draft protection. Native menu dismissal remains unverified. |
| Layout | Settings sections in a ScrollView; no fixed action overlay. Error view scrolls. Native header insets and menu geometry require acceptance. |
| Adaptation | Shared wrapping rows/native picker layout. Phone/tablet window sizing remains unverified. |
| Typography | Shared unconstrained picker labels (M53), wrapping section descriptions and issue text. Normal-size long provider/model names need native checks. |
| Appearance | Semantic palette and shared native choices/actions; M180 retains the standard labeled progress row. Contrast has not been measured here. |
| Localization | English readiness/action copy; implementation terms are disclosed in advanced profile context. RTL and translated service labels remain open. |
| Imagery | No media thumbnails; navigation chevrons distinguish inspection from service choice. Status is communicated in text. |
| Targets | Shared Settings row targets and native picker/Retry. Actual iOS/iPadOS/Android target bounds remain unverified. |
| Gestures | Explicit native Back and row/menu commands; no gesture-only state change. |
| Keyboard | No keyboard for overview/service selection. Credentials and prompt entry are separate audited editors. External keyboard menu navigation remains pending. |
| Accessibility | Named stage/profile navigation with status/context, disabled working actions, progress text. VoiceOver row/value ordering and announcements remain unverified. |
| Motion | No custom screen animation. Native transitions/progress need reduced-motion runtime acceptance. |
| Content | Current service, readiness, selection source and issue-specific next action. Overview counts advanced profiles; no fabricated ready state for missing/invalid selection. |
| Search | Not applicable to these overview/stage surfaces: no live-search task; service choice is a flat native menu. Large-profile-set usability is not certified. |
| Loading | M180 names setup/stage/profile/editor loading instead of an anonymous spinner. Context bridge names household context. Access verification already has text. |
| Recovery | Retry retains existing scoped query behavior; stale refresh errors show a notice with cached content. M180 error heading names the actual task. Missing stage remains an explicit error with native Back available. |
| Editing | Selection/test/enable lock concurrent commands; profile edits navigate separately. No local unsaved overview draft. Async scope/visit notice ownership remains M105 evidence. |
| Privacy | Every route uses the configure-permission guard. Existing tests cover permission decisions; this presentation change does not modify server authorization. Credential loading means settings metadata, never retrieval of the secret. |
| Notifications | Not applicable: no push registration or notification entry is owned by these screens. |
| Media | No recording or playback in readiness UI; Test invokes the existing provider test command. Physical microphone/speech acceptance belongs to recording. |
| Lifecycle | Scoped queries withdraw access-failed data; controls reject working duplicates. M105 guards stale completion notices. Native background/resume and open-menu scope change remain pending. |

M180 shared consumers: setup, capability, profile list/detail, credential editor,
prompt editor, and household-context bridge. Six mounted RED cases reproduced
missing task text, then verify loading/error identity and Retry dispatch. All110
related provider/settings/guard tests and static checks pass remotely on paul
(`/tmp/provider-readiness-green.log`); a final credential-settings wording check
passes six cases and static checks (`/tmp/provider-state-copy-check.log`). Code
critic found no blocker. Native layout and announcements are not established.
