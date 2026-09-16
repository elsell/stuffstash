# Home header native run 35059882579

Source 28c131bb. iPad job104679402657 finishes 3/4, phone job104679402797 finishes 2/4.
Both geometry tests stop on Add's 36-point accessibility width, before the scroll
checks. These totals are retained; this report does not relabel failed tests.

| Action | iPad | Phone |
| --- | --- | --- |
| Add | All nine destination/return probes pass | All nine destination/return assertions and final header existence complete; subsequent screenshot request times out |
| Notifications | All nine counted callback probes pass | All nine counted callback probes pass |
| Profile | All nine destination/return probes pass | All nine destination/return probes pass |

The probes sample the center and points 21pt from it along each axis and diagonal,
inside a centered 44-point square. They establish delivery at those sampled points,
not a measurement of the entire hit area or VoiceOver acceptance. Phone Add's
failure occurs in `capture`, after the final header-existence assertion; the log
and teardown capture remain evidence rather than converting its status to green.
The inspected [phone teardown](evidence/phone-home-add-return-350598.png) and
[iPad return capture](evidence/ipad-home-add-return-350598.png) show the three
separate actions in Add/notifications/Profile order, alongside the inventory
switcher, after navigation returns. These fixtures omit the full tab/accessory
composition; they do not certify that surrounding layout.

[Apple's Buttons guidance](https://developer.apple.com/design/human-interface-guidelines/buttons)
describes a 44×44pt hit region. The observed native action delivery outside the
reported accessibility bounds means those bounds are an inadequate proxy for the
touch region in this fixture. The layout journey now keeps positive bounds,
hittability, onscreen containment, action order, selector clearance and fixed
header position while scrolling. It retains all independent touch probes and raw
frames in hierarchy captures. No production control is resized for this assertion.
Native rerun is required; the previously unreached scroll checks remain open.

Logs: `/tmp/native350598-home-ipad.log`, `/tmp/native350598-home-phone.log`.
Artifacts10433015936 and10433196016 retain the full results. Remote mobile
structural validation passes for the test/spec correction; Swift compilation
and execution remain pending.
