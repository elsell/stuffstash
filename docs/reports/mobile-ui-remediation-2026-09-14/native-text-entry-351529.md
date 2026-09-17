# Native key-event evidence — run35152978881

Source aba0ebb811064010ea919e26835715a48a3babe7. Phone job104985587676
finishes9/14 comparisons, with5 failures in461 seconds. iPad job104985587501
finishes11/14, with3 failures in558 seconds. Phone artifact10470638215 is retained on paul only at
`/tmp/native351529-focused-phone.zip`; compact
[raw traces](evidence/phone-input-events-351529.json) are preserved here.
The complete local log is `/tmp/native351529-phone.log`.

## What changed in the evidence

All four observed character-loss failures receive the full key sequence
`Native draft name`, with zero trace-buffer drops. Ordinary controlled name,
controlled name without assistance and seeded ordinary name end as
`Nve draft name`. The seeded no-accessory comparison ends as `Nive draft name`.

For the first three, events begin: key N, change N/count1, React commit N,
keys a/t/i/v, then change Nv/count2. For the no-accessory case, keys a/t/i
precede change Ni/count2. The missing letters therefore reached the native
keypress path but never appeared in the received text-change values. This narrows
the investigation beyond the prior change-only traces; it does not identify the
specific native operation that discards them.

Pinned React Native0.83.6 emits keypress in `textInputShouldChangeText:inRange:`
before the backing field's edit and emits change from `textInputDidChange`.
The fixtures do not set maxLength. Inspect the backing field/delegate, state/text
synchronization and native change suppression next. Do not infer that pacing or
disabling assistance fixes production behavior: the unassisted controlled case
still loses letters, and earlier runs retain additional intermittent failures.

The other failure is different: controlled name without accessory aborts in
`waitForKeyboard` when XCTest evaluates hittability of a key with an infinite,
zero-size frame. No typing or trace export occurs for that case. It is missing
comparison evidence, not another observed character-loss result. Eleven RN traces
are exported; the two SwiftUI comparisons do not use this tracer.

Controlled, seeded and system address comparisons pass; multiline, native-default
assisted entry, seeded unassisted entry and both paced diagnostics also pass.
Passing this sample does not overturn earlier failures. No production input change
or whole-app native acceptance is claimed.

## iPad comparison

Artifact10470798004 is retained on paul at `/tmp/native351529-focused-ipad.zip`;
[compact raw traces](evidence/ipad-input-events-351529.json) preserve all12 RN
comparisons with zero drops. Log: `/tmp/native351529-ipad.log`.

All failed inputs again receive the complete expected key sequence. Controlled
address ends as `h/example.invalid`: after change h/count1 and React commit h,
keys t/t/p/s/:/slash/slash precede change h/slash/count2. Controlled name without
accessory ends as `Ndraft name`: keys a/t/i/v/e/space/d precede change Nd/count2.
Ordinary controlled name reorders text to `Native aft namedr`; after change
`Native `/count7 and its React commit, keys d/r/a precede change `Native adr`/count8
with the selection at8. Thus this case also shows an insertion-position problem,
not only missing text. Other comparisons pass in this sample.

These paired traces justify investigating native text/selection synchronization
around pending edits, including the single-line delegate's pending-change state.
They do not justify a production workaround or a claim that XCTest alone caused
the problem. The focused filter run35154627907 has now started on both targets;
full native run35148054814 remains active.

## Next controlled comparison

The no-accessory cases still mount AppKeyboardProvider. A runner-only
`text-entry-no-provider` selection now omits both provider and accessory while
running the same fourteen comparisons. Each test retains a configuration attachment
from the root's native identifier so provider absence can be checked in the result.
The normal/full fixture root retains the production provider. Installation tests
first fail for the missing distinct root, then pass after the selector is added.

The library's [upstream issue1588](https://github.com/kirillzyusko/react-native-keyboard-controller/issues/1588)
reports delegate forwarding to the wrong input after navigation. Its reported
missing key events differ from our complete key sequences; it is not an established
explanation here. It does establish why merely hiding the accessory is insufficient
to exclude the provider's delegate hooks. A provider-free comparison can narrow
that boundary before any framework patch or production input replacement.
