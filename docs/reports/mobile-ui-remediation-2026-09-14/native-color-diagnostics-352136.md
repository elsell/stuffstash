# Native color diagnostics —35213668166

Sourcee5b96d4cc6cf6cf5e57478d397ecc2f52717ffb2. Focused color suite passes7/8
on both devices. Only the without-pre-tap-capture comparison fails. Both released
RGB and lock/unlock workflows, original opening journey and all nine delivered
hit-region probes pass. Prior intermittent failures remain valid evidence.

The added xcresult diagnostic export succeeds on both. Export receipts are in
evidence/. Artifacts10494755730(phone) and10495505411(iPad) contain the application
StandardOutputAndStandardError logs even though the booted-simulator collector
produces none. Range downloads retrieved about11.4MB compressed in total, avoiding
both whole archives. Selected logs were examined locally; unrelated extracted
system diagnostics were removed after inspection of the archive index.

Phone XCTest records entry tap11:18:45 and color tap11:19:11. The accessibility
existence query between them takes substantially longer than its nominal timeout.
The tap therefore occurs roughly26 seconds after entry, weakening an explanation
based solely on needing a brief mount delay. App-log inspection found no explicit
attempt-to-present/already-presenting/window-hierarchy rejection around the failed
opening. This absence is not proof that presentation was requested successfully;
logs do not establish delivery to the control. No touch-receipt instrumentation
exists yet. No production delay, retry or default-color substitution is justified.

M51 remains open. Both runs352107 and352136 fail the no-pre-tap-capture comparison
on phone/iPad while passing the capture variant. A next investigation needs to
separate accessibility snapshot effects, native touch delivery and presentation
ownership, rather than assuming a timing cause. These are assertion/log findings;
this run's screenshots have not been reviewed.

The observer slept120 seconds between terminal-state checks; jobs were not
restarted. The broader142-by24 audit and normal-text-first priorities remain.
