# Conversation route — 24 source axes

R057 at63d17e79 plus M179. This integrates the typed composer, recording,
processing, response, proposal, progress and destination reviews; it does not
replace their findings or certify their native behavior. Route composition was
inspected in VoiceSessionSheetScreen, VoiceConversationHeader, root navigation
and VoiceNativeSheetOptions. Provider and physical-device acceptance stay separate.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | One retained conversation supports typed/spoken requests, reading results and approving explicit writes. M170/M178/M179 use native command adapters rather than competing custom command styles. |
| Navigation | Native Close/New header; response links dismiss before detail navigation with visit ownership (M171). M174 destination route returns to the proposal. Native sheet-to-stack/return remain pending. |
| Selection | Editable proposal names stay inline; hierarchical destination selection has current/checkmarked values and native search. Photos use native source chooser. See separate S122/R142 reviews. |
| Modality | Native form sheet offers compact/expanded detents. Destructive new-conversation confirmation is scoped (M144); duplicate reset removed (M178). Native swipe/background acceptance remains open. |
| Layout | Header actions belong to navigation; body scroll and bottom composer/approval area divide the viewport. Header replacement changes available height; keyboard and compact detent fit require fresh native evidence. |
| Adaptation | Flexible body, wrapping context text, platform native header/commands. iPad window sizing and Android rendering remain unverified. |
| Typography | Context no longer forced to one line. Response, progress and proposal copy wrap; long names and normal-size dense content still require native inspection. |
| Appearance | Shared semantic palette, native header/commands. M176 distinguishes attachment warning from saved-change success; M179 removes the custom provider recovery button. No measured contrast certification. |
| Localization | English interface and locale-aware expiration helpers; long translations/RTL remain open. No implied full localization support. |
| Imagery | Result cards and photo drafts supplement text. Missing/error thumbnails are covered by their shared consumers; camera/library permission remains physical evidence. |
| Targets | M170/M175/M178/M179 replace custom command controls with native adapters. Proposal edit/location rows and diagnostic disclosure remain semantic content controls. Actual target bounds are not established by source dimensions. |
| Gestures | Tap alternatives for rail browsing and native explicit Close. Diagnostics disclose through a button; voice input also has typed alternative. Keyboard/scroll interaction requires runtime checks. |
| Keyboard | Keyboard avoidance wraps conversation and safe-bottom action area; native search belongs to destination route. Approval includes pending name (M173). Native overlap/focus remains open after header change. |
| Accessibility | Named native actions, progress labels, checked destinations and grouped result text. Structured content source review does not establish VoiceOver order or focus restoration. |
| Motion | Shared reduced-motion preference covers result scrolling; native transitions and spinner behavior need runtime acceptance. Existing M62 evidence remains retained. |
| Content | Bounded in-memory history, transcript, response, proposal and attachment status. Diagnostics require explicit enablement and use safe event summaries; no production instruction exposes raw traces. |
| Search | Conversational requests are submitted, not live filtered. Destination lookup uses explicit native search with debounce and empty/error feedback. |
| Loading | Context load, recording, processing, write execution and photo upload have separate owners. M177 supplies context Retry; recording-start indication remains an open usability risk in S119. |
| Recovery | Initial context retries in place (M177), provider recovery opens Voice setup (M179), photo retry targets failed attachments (M176), and invalid destinations return to conversation. Native recovery transitions remain pending. |
| Editing | Draft names/locations/photos are above the sheet; explicit review controls approve/cancel. M178 leaves one protected reset path. Scope changes retire edits; no offline mutation queue. |
| Privacy | Scoped context, explicit mutation approval, safe diagnostic presentation and opt-in capture. Existing security suites remain authoritative; UI synthetic fixtures do not establish authorization or physical permission behavior. |
| Notifications | No direct conversation notification registration. App-level incoming links and interruptions need lifecycle acceptance; no route-owned notification setting to verify. |
| Media | Native recorder, playback and photo selection remain behind ports. M172 startup/cancel serialization is source-tested; actual permission prompts, interruption and camera access remain open. |
| Lifecycle | Provider retains exchanges across sheet close and clears on scope/reset. Generation/visit guards protect async results (M171/M172/M177). Cold start intentionally loses in-memory conversation; native close/reopen fixture is pending. |

Linked source reviews: [typed entry](conversation-composer-axis.md),
[recording](voice-recording-axis.md), [processing](voice-processing-axis.md),
[response](conversation-response-axis.md), [proposal editing](voice-plan-edit-axis.md),
[progress/recovery](voice-progress-axis.md), [destination](voice-location-axis.md).

M179 is a direct adapter substitution: the failure-specific label and callback are
unchanged. Existing presentation/adapter/navigation tests are used rather than a
new test that only mirrors component markup. Native visual acceptance is pending.
