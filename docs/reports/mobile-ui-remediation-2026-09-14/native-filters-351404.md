# Filter native review — September 16

Run [35140471580](https://github.com/elsell/stuffstash/actions/runs/35140471580)
at a1b827e0 records four relevant normal-size passes on both iPhone17 and iPad mini:
Browse final-tag selection/application, in-place availability selection, expiration
calendar open/dismiss/Back, and expiration tag search while the keyboard is open.
Visual inspection identifies an unresolved overlap despite the last test's pass.

## M249 — Keyboard accessory overlays the phone Back command

The [phone keyboard capture](evidence/phone-filters-keyboard-351404.png) shows the
Back label washed out beneath the app's keyboard-dismiss accessory. Its
[hierarchy](evidence/phone-filters-keyboard-351404.txt) places Back at Y491–545 and
Dismiss keyboard at Y485–529. Both are44+ point native controls, but their frames
overlap. The prior test only checked hittability and successfully tapped Back;
that does not establish unobscured visual presentation. Normal text, light mode.

The [iPad comparison](evidence/ipad-filters-keyboard-351404.png) shows both commands
fully above the detached keyboard accessory in this sheet configuration. This
single configuration does not establish every iPad window/keyboard arrangement.

The shared iOS sheet hook currently remeasures anticipated keyboard frame changes,
but ignores settled frame/show notifications. A failing mounted regression proves
it retains the old inset when the settled keyboard frame and sheet boundary change.
The candidate listens to those settled notifications and remeasures the unmoved
boundary, retaining generation guards and hide behavior. Eleven focused tests pass
on paul after the failing-first regression. TypeScript, six fixture-preparation
checks and the mobile structural check also pass; code critic found no blocker.
This does not prove that missed settled
notifications caused the captured overlap or that the native defect is fixed.

Native acceptance now requires both entire command frames to remain above the
Dismiss keyboard control after layout settles, then taps Back. No assumed accessory
height is added. Browse and expiration share this hook through NativeFilterSheet;
Android has its own adapter and is unchanged. The new native assertion and candidate
are not included in the already-running jobs and still require execution.

## Scoped passes retained

[Phone final tag](evidence/phone-filters-tags-351404.png) and
[iPad final tag](evidence/ipad-filters-tags-351404.png) show the last list row fully
above both footer commands. The test also selects that row, returns to the overview,
applies, and verifies the exact selected ID. This supports the specific long-list
clearance correction; it does not verify keyboard clearance or assistive reading.

[Phone date range](evidence/phone-filters-dates-351404.png) and
[iPad date range](evidence/ipad-filters-dates-351404.png) show controls and commands
after dismissing the system calendar. The test opens/dismisses the calendar and
returns through the filter overview; it does not select and persist a new date.

The availability case selects Available from the in-place menu, retains the Filters
screen and applies the expected value. No screenshot from that case was inspected
in this pass; its interaction assertions are log evidence only.

The input-search case retains the full Tools query and filters out Holiday supplies.
Its old green result must not close M249. Appearance, larger text, permission/loading
states, VoiceOver and full production route/API integration remain separate gates.
