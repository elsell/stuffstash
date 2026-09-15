# Exact-date and month/year entry

S092–S093, inspected atcd6f6556 plus M134 candidate. Shared ExpirationField is
consumed by AssetExpirationEditor in Add and Edit. Native runtime acceptance is
pending; these source checks do not establish keyboard or picker geometry.

| Axis | Exact date and month/year assessment |
| --- | --- |
| Task | Precision reflects the label on the item; a two-option native segment fits. Do not invent a day for month-only labels. |
| Navigation | Inline disclosure keeps editing in the current draft. No new navigation stack is added. |
| Selection | iOS compact date picker; Android system date dialog. Month is a native choice menu; year uses number-pad input. |
| Modality | Android dismissal publishes nothing. iOS edits the parent draft directly; persistence belongs to parent Save. |
| Layout | Inline vertical field with gap; disclosure retains hidden child content. Keyboard/scroll clearance requires native checks. |
| Adaptation | Native controls own picker presentation; shared menu adapts its label. Phone/iPad presentation unverified. |
| Typography | No explicit line truncation in field. Segment widths and long locale month labels unverified; larger-text pass deferred. |
| Appearance | Theme colors and shared native command/menu adapters. Date picker appearance and contrast require rendering. |
| Localization | Month names use Gregorian locale formatting; alternate-calendar users receive a notice. Year placeholder YYYY is explicit. Exact dates store local civil fields. |
| Imagery | No asset imagery in this field; disclosure symbol supplements current value. |
| Targets | Native date/menu/command controls; year minHeight44. Actual bounds and keyboard interaction unverified. |
| Gestures | Date/year selection and explicit Clear have visible controls. Native date dismissal and picker gestures need runtime. |
| Keyboard | Year number pad, no freeform exact-date input. Late year callback while disabled was not rejected (M134); date callback already was. |
| Accessibility | Expiration/date/month/year names and disclosure value; validation has live-region text. Actual traversal/announcements unverified. |
| Motion | No field-owned animation. Native menu/picker Reduce Motion behavior unverified. |
| Content | Single date with explicit end-of-day or end-of-month meaning. No pagination. |
| Search | Fixed twelve-month/date values; search is not applicable. |
| Loading | Parent passes disabled during an operation; field has no independent query or loading spinner. |
| Recovery | Incomplete month/year reports invalid; complete empty value is allowed. Clear removes all precision drafts. |
| Editing | Changes update the parent draft; clearing and changing precision do not revive removed dates. M134 protects year draft while disabled. |
| Privacy | No network or permission request in field. Parent owns asset authorization and save; no boundary certification here. |
| Notifications | Date edits do not directly schedule notifications; outside this field. |
| Media | No camera/files/audio use; not applicable. |
| Lifecycle | Initial seed is retained locally; parent keys reset asset/type or restored draft. Focus/background/native event timing unverified. |

Existing mounted cases cover month validity/clear, Android cancel, native iOS
change, precision conversion, removed-date non-revival and disabled date callback.
M134 adds disabled-year publication and local-draft preservation plus editing after
re-enable. Native acceptance must exercise both Add and Edit, actual year typing,
Save pending, picker dismissal and ordinary-text phone/iPad scroll clearance.
