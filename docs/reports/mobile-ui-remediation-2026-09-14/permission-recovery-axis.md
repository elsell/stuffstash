# Device permission recovery

S132, source 91c22aa3 plus M190. Reviewed photo selection, voice recorder startup,
notification device setup and their existing task/error consumers. Permission is
requested by explicit camera/record/setup actions, not a global startup prompt.
Library selection uses the system picker without broad library permission.

M190 adds recovery guidance for camera/microphone denial: allow access for Stuff
Stash in device settings and retry; library selection/typed conversation remain
alternatives. It does not automatically open Settings or resume capture. A direct
Settings shortcut for these two workflows remains a usability recommendation;
this change provides guidance through their existing error surfaces.

M191 is still open: NotificationSettingsScreen discards both Linking.openSettings
promises without handling failure. The [React Native API](https://reactnative.dev/docs/linking#opensettings)
returns a promise; source review establishes missing rejection handling, not an
observed OS failure. Add safe current-visit fallback/retry and controlled tests.
Apple's [privacy guidance](https://developer.apple.com/design/human-interface-guidelines/privacy)
returned a JavaScript shell during this pass; no unavailable text is quoted.

| Axis | Source result and remaining acceptance |
| --- | --- |
| Task | Consent enables the requested feature; denial must leave a usable alternative. M190 adds missing recovery guidance. |
| Navigation | Notification setup has explicit Settings links. Camera/microphone currently provide manual Settings guidance; no automatic external navigation. M191 tracks failed Settings opening. |
| Selection | OS owns permission choice. App does not imitate the system permission prompt. |
| Modality | OS consent is system-modal; existing app error surfaces explain denial. No additional pre-permission modal added. |
| Layout | Existing Add/detail/voice feedback and notification settings own layout. Longer M190 guidance needs normal-size native captures. |
| Adaptation | Photo/audio APIs shared across mobile platforms; push device asserts iOS/Android support. Native phone/iPad/Android verification remains open. |
| Typography | Error text follows consumer styles; long guidance and large text not certified. |
| Appearance | Native system prompt plus semantic app feedback; contrast/disabled states require runtime evidence. |
| Localization | Guidance is English and avoids a platform-specific Settings hierarchy. Translation/RTL unverified. |
| Imagery | N/A: recovery itself has no custom image content. Permission prompt app identity is OS configuration. |
| Targets | Notification Settings rows use existing actions. Camera/mic manual guidance has no separate tap target; direct shortcut is a recommendation. |
| Gestures | Permission/retry alternatives have explicit commands; no gesture-only requirement. |
| Keyboard | Typed conversation remains the mic-denied alternative. Actual keyboard restoration and capture-source return need native acceptance. |
| Accessibility | Existing error presentation supplies guidance; system permission reading order and app error announcements require assistive checks. |
| Motion | OS transitions, no authored permission animation. Reduced-motion behavior unverified. |
| Content | Camera/mic reasons now include correction and alternatives. Notification outcome describes setup attempt, not live permission truth. |
| Search | N/A: permission recovery is not discovery/search. |
| Loading | Notification setup is guarded as pending; voice startup supports cancellation before capture. Camera system prompt lifetime still needs physical testing. |
| Recovery | Denial does not launch camera/start audio. Later explicit attempt rechecks permission. M191 still lacks Settings-opening failure recovery. |
| Editing | Photo and conversation tasks retain their existing drafts/state; no new clearing on denial. Adapter tests do not prove every mounted draft consumer. |
| Privacy | Library avoids broad authorization; camera/audio gate capture on granted. OS restrictions, revocation and limited access need physical evidence. |
| Notifications | Permission and inventory push preference are distinct. Background reconciliation reads status; explicit setup requests consent. |
| Media | Camera and recording begin only after granted result. Tests cover denial then grant/retry; no physical media claim. |
| Lifecycle | Voice cancellation checks after permission; notification feedback retires on background/focus change. Actual Settings return must reattempt/reconcile rather than assume grant. |

All 109 related photo/audio/session checks plus TypeScript and structural checks
pass on paul (`/tmp/permission-recovery-green.log`). The two strengthened denial
checks failed before M190 because guidance was absent; permission gating itself
was already correct. Physical consent, Settings return, permanent restriction,
revocation and long-message native fit remain open. Code review found no blocker
in the M190 guidance change.
