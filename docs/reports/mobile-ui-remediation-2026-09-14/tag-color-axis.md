# Optional tag color — all 24 source axes

S108 at71d54bb6. Consumers: settings tag create/edit, Add staged-tag creation,
and item Edit tag creation. Reviewed TagColorPicker, FullSpectrumTagColorPicker,
NativeTagColorPicker.ios and their presentation helpers and consumer composition.

[Apple's color-well guidance](https://developer.apple.com/design/human-interface-guidelines/color-wells)
recommends considering the system picker for familiarity. The project chooses it
on iOS; the pinned Expo package has no equivalent Android picker. The fallback
is a documented platform limitation, not permission to substitute custom iOS UI.

| Axis | Source conclusion and remaining acceptance |
| --- | --- |
| Task | Choose optional decorative metadata for a named tag. In-place presets plus a system color well fit; there is no reason for a separate navigation route. |
| Navigation | iOS system presentation returns to the existing draft; Android custom panel expands in context. Actual native return/focus needs repeated phone acceptance. |
| Selection | Six named presets, explicit No tag color, arbitrary RGB selection. Empty values map to no native selection without inventing a persisted color. M51 activation remains unresolved. |
| Modality | Current iOS code opens the native picker directly; the earlier redundant custom panel is removed. Android retains its local Done/Cancel transaction. |
| Layout | Presets wrap; the native host fills available width with a minimum height. The fallback separates spectrum gestures from scrolling supplementary controls. Keyboard/small-window runtime acceptance remains open. |
| Adaptation | Native color presentation owns phone/tablet differences. Fallback compacts its spectrum according to measured space/font scale. This does not establish enlarged-text acceptance. |
| Typography | Text labels supplement swatches; native label allows wrapping. Long labels and large text remain runtime work. |
| Appearance | Semantic palette for neutral controls; selected checkmark contrast derives from each swatch. System picker owns its chrome. Light/dark runtime parity is unverified. |
| Localization | English preset names; arbitrary RGB values remain stable hex. RTL spectrum direction and localized long labels are unverified, not declared compliant. |
| Imagery | No image assets. Checkmark, clear symbol and named colors prevent color being the sole indicator. |
| Targets | Presets declare minimum targets; native well declares44 points but exposes28 on phone/36 on iPad. Actual hit area is separate from AX frame. M51 remains open. |
| Gestures | iOS uses native activation; Android offers adjustable accessibility actions, adjustment buttons and hex input alongside dragging. Native activation still fails on the captured phone run. |
| Keyboard | iOS color editing delegates input to the system. Fallback hex entry disables correction and labels format. Actual keyboard/dismissal/return acceptance is pending. |
| Accessibility | Presets expose names, selected and disabled states. Native capture duplicates Choose any color in the button label; spoken output/order is unverified. Android exposes hue/saturation/brightness adjustments. |
| Motion | No project-owned color transition animation; native presentation/reduced-motion behavior remains unverified. |
| Content | Small finite preset row with optional full picker; named read-only value replaces mutation controls for inherited/read-only definitions. |
| Search | No search task in a small color palette. |
| Loading | Color selection itself is synchronous. Parent mutation state disables interaction; native availability is checked before choosing the adapter. |
| Recovery | Invalid values receive corrective text; clear removes optional color. Parent save failures retain drafts. Opening failure currently has no recovery beyond retry/presets: M51. |
| Editing | iOS selection updates the parent draft, not the server. Explicit clear is distinct from opening. Android Cancel discards its temporary draft and Done commits a valid value. |
| Privacy | No additional service or permission requested. Parent settings/item authorization gates persistence; a color control is not an authorization boundary. |
| Notifications | No notification task or permission in this surface. |
| Media | No camera/library/file/audio operation. System picker internals do not justify adding app media permissions. |
| Lifecycle | Parent draft/workflow owns navigation and save. Component unmount discards fallback-local editing state. Background and system-picker interruption require native acceptance. |

## Evidence

Existing mounted tests exercise presets, clear, invalid input, disabled choices,
fallback arbitrary colors, Cancel/Done, adjustable spectrum and compact geometry.
The complete remote suite at71d54bb6 passes1,824 tests; these do not mount SwiftUI.

Native350465 atb6321dcb: phone system-picker activation fails and leaves the parent
unchanged; iPad open/clear passes. See [phone result](native-phone-350465.md).
No speculative adapter change is justified by that difference alone. Repeat native
center activation, selection, close, explicit clear, disabled state, parent draft
retention and re-entry before closing M51. Source review completes the18 pending
cells for S108 without turning the three existing M51 finding cells into passes.

Run350504 phone follow-up: inspected artifact10429678833, test
testColorPickerOpensDirectlyAndClearPreservesParentDraft, final screenshot
A0767C68-7A0D-4725-9825-E930AC2F94B6.png and hierarchy
8264F8CC-704E-4162-8ABF-7C216E49F56F.txt. After XCTest taps the color button, the
parent remains visible with no picker and Color value: none. The hierarchy places
the button at(346,360.7),28×28 with duplicated Choose any color label. Thus the
failed Sliders assertion corresponds to an unopened picker, not merely a changed
selector. AX bounds alone still do not prove the effective hit region. This
reproduces the earlier phone observation at source802e4955; no production fix or
root-cause attribution is claimed.

M213 separately tracks the duplicated accessible name. The candidate removes the
additional accessibility-label modifier while retaining ColorPicker's native label.
Apple documents that hidden control labels still serve accessibility:
[labelsHidden](https://developer.apple.com/documentation/swiftui/view/labelshidden()).
That guidance supports preserving the native label, not a claim that this edit fixes
activation. The new independent exact-name native test is queued;15 related remote
tests and static checks pass, critic found no blocker. Native name/VoiceOver
verification remains open.

Run350592 confirms one accessible name on both devices and direct opening/clear
on phone. iPad direct activation fails: its inspected final capture and hierarchy
show the unchanged parent without a picker. See `native-fixtures-350592.md` and
`evidence/ipad-color-unopened-350592.png`. M213's naming correction has native
name evidence; VoiceOver output remains unverified. M51 activation remains open,
with different failing devices across runs. No blanket success or source cause
is inferred from the phone pass.

Follow-up native instrumentation retains both existing activation/frame journeys
and adds nine delivered-touch probes centered on the visible well. Every probe
must open and dismiss the system picker without changing the unset parent value;
failed opening/dismissal captures the state and stops that journey. The focused
color-picker workflow includes this diagnostic. Remote structural validation and
critic review pass; execution on phone/iPad is pending. This is evidence-gathering
for M51, not a production correction or a replacement for color-editing acceptance.

Run350695 passes all nine delivered-touch probes on both devices, plus ordinary
opening/clear and single-name checks. AX frames remain28pt on phone and36pt on
iPad. Following review, the separate opening test now requires nonempty, onscreen,
compact AX bounds instead of equating that frame with the effective touch region.
All nine touch probes remain required. This corrects an invalid measurement proxy;
it does not retroactively mark the failed runs green or prove every point in a
44-point region. Remote structural checks pass; the revised Swift assertion awaits
native execution. VoiceOver, color-editing acceptance and inconsistent historical
activation remain open.

## Android custom-color staging sample

September16, Android16 Pixel6, normal text/light, synthetic APK
`d3fc5d55763fc45cb8f6e5c93cd116440570d1613b0646c884c2fad636e0bd32`:
the Android fallback expands inline; scrolling outside the color surface reveals
Hue/Saturation/Brightness adjustments, Hex color, Clear color, Cancel and Done.
Increasing Hue and cancelling leaves No tag color selected. Reopening restores
the original214-degree draft, rather than the cancelled adjustment. Increasing
Hue again and choosing Done selects [Custom and marks the parent draft
unsaved](evidence/android-tag-custom-draft.png). Parent Save reaches the fixture
collection and shows Tag saved. No server or real definition was mutated.

The destination collection uses its own synthetic repository; its rows do not
prove persisted color readback. This sample verifies cancel/staging and successful
command navigation only. Hex keyboard editing, direct color-surface dragging,
Clear, rejection recovery, TalkBack, large text and iOS color activation remain
open. The route-name header in the capture belongs to the fixture configuration;
it is not evidence of the production title.

Evidence: `/tmp/android-tag-{editor,custom,custom-bottom,cancel,reopened,done,saved}.xml`
and `/tmp/android-tag-done.png`. No implementation changed for these checks.

### Hex entry and Clear follow-up

On the same Android build/configuration, entering `123456` updates the color draft.
The [keyboard covers Cancel/Done](evidence/android-tag-hex-keyboard.png); an outside-
color-surface scroll gesture dismisses Gboard and exposes both commands. This is
verified dismissal/reachability, not continuously visible keyboard actions.
Done followed by reopening shows normalized `#123456`, confirming the staged
value. Clear color followed by Done restores the original No tag color selection
and removes Unsaved changes. The parent Save remains unnecessary for that net-zero
edit. Direct color-surface dragging, invalid hex, rejection recovery and assistive
technology remain outside the sample.

Evidence: `/tmp/android-tag-hex-{entry,keyboard,scrolled,retained}.xml`,
`/tmp/android-tag-{cleared,clear-done}.xml`, and keyboard/scrolled PNG captures.

### Invalid hex recovery and direct dragging

September16, Android16 Pixel6, normal text/light, APK
`b76e227a0c215b08da22f929024514b4b0721f9781e1d209811aaff877d1b4c1`.
TagColorPicker and FullSpectrumTagColorPicker source hashes match the current
branch. Using the settings-controls fixture, typing `ZZZZZZ` shows
[inline validation and disabled Done](evidence/android-color-invalid.png).
Actually tapping disabled Done leaves the panel and parent `none` unchanged.
Replacing the input with `123456` removes the error, enables Done, and applying
produces exact [parent value #123456](evidence/android-color-recovered.png).
Android Back dismisses the keyboard before these commands are checked.

Reopening and dragging the [saturation/brightness surface](evidence/android-color-drag.png)
from one-quarter to three-quarters across both axes changes the announced values
from79%/34% to75%/26%. Its frame stays `[98,1396][984,1816]`, demonstrating that this
gesture adjusts color rather than scrolling the parent. Cancel preserves the
previous parent value `#123456`. Evidence XMLs and screenshots are retained as
`/tmp/android-color-{invalid,recovered,drag}.{xml,png}` on both hosts.

This closes the sampled normal-size Android invalid-hex recovery and direct
spectrum-drag checks. It does not establish server persistence/rejection recovery,
TalkBack operation, other layouts, or iOS activation. M51 remains unresolved.
