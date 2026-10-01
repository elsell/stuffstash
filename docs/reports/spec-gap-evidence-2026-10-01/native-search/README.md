# Native search acceptance — current diagnosis

PR #233 changes test observation, fixtures, evidence and documentation; it does not
change production interaction behavior. Product fixes shipped in v0.27.3. The
remaining runner failure is tracked separately in [issue #236](https://github.com/elsell/stuffstash/issues/236)
and does not indefinitely gate this verification/documentation PR.

[Targeted run36933539569](https://github.com/elsell/stuffstash/actions/runs/36933539569),
source `7f20184e6ce0f18e653703910a26f7bac3eeda7c`: iPhone passed tag search,
expiration search and static query/clear/retype. iPad passed static search; tag
and expiration failed inside `app.launch()` before product assertions. The same
iPad setup boundary failed in run36928351670. Those two workflows remain
**unverified on iPad**. Stop unchanged retries; do not weaken assertions or change
production behavior based on these failures.

The earlier normal-search follow-up36928351670 passed seven iPhone workflows and
six iPad workflows, including the managed React-state/native-title checks and
proposal location retry/selection/return. Its iPhone static comparison immediately
read `mi`; the [final screenshot](iphone-static-query-retained.png) shows complete
`missing`. The shared bounded observer correction subsequently passed that fixture
on both devices. It retains exact equality, the time budget and input behavior.

[Expansion run36930043292](https://github.com/elsell/stuffstash/actions/runs/36930043292),
source `f8f5c172cbb7190be039e405d43c2bf1d79af61b`, passed all three Add/recovery,
protected-proposal and location-search workflows on iPhone and iPad. Inspected
screenshots show the complete query, preserved user location names and visible
completion actions:

- iPad: [empty search](ipad-expanded-empty-search.png), [selected location](ipad-expanded-selected-location.png), [returned proposal](ipad-expanded-returned-proposal.png).
- iPhone: [selected location](iphone-expanded-followup-location.png), [returned proposal](iphone-expanded-followup-proposal.png).

This is normal-text simulator fixture evidence, not connected voice, physical
integration or whole-app localization acceptance. RTL run36936112035 was already
started; its result remains unverified here and is not a new merge gate. No more
runs are authorized merely to make this PR's native ledger green.

## Earlier evidence

Run [36923519340](https://github.com/elsell/stuffstash/actions/runs/36923519340),
source `58bb40a81d0ce7a70863222a10db78964e2ee6df`, English, normal text size.
Both iPhone17 and iPad mini (A17 Pro) completed eight tests: seven passed and one
failed on each device. The suite did not pass overall.

- iPhone: the managed-search probe still displayed its original title after the
  reconfigure tap. The fixture exposed no React-state marker, so the evidence
  cannot distinguish a missed state change from a native presentation failure.
  Production place search, proposal location retry/return, settings search,
  tag search/keyboard clearance, expiration search and static placement passed.
- iPad: the query observer timed out on a previously captured search element.
  The final hierarchy and screenshot show `value: Tools`, keyboard focus, and the
  correctly filtered Tools row. This contradicts lost-input diagnosis; it does not
  prove the later apply/return steps, which the failed test never executed.
  The managed probe and six other workflows passed.

Inspected evidence:
- [iPhone probe unchanged](iphone-probe-unchanged.png)
- [iPad retained query and result](ipad-query-retained.png)

One bounded follow-up changes test observation only: re-query the current search
field in the predicate, and expose/assert the probe's React reconfiguration state
before asserting its native title. If the body marker fails, investigate input
activation; if it passes but the title fails, investigate header propagation.
Do not change production search behavior based on these inconclusive failures.
A fresh native result is still required; no retry result is claimed here.

Expansion run36926429000, sourcec7a92147, iPad completed Add/recovery and
protected-proposal close/reset, but location search's immediate assertion read
`m` after typing `missing`. The retained [final screenshot](ipad-expanded-query-retained.png)
and hierarchy show the full `missing` query and correct empty result. Replace the
immediate assertion with the same bounded live-field observer used above; retain
exact value equality and the existing timeout. No input replay, forced value or
production behavior change is allowed to make this assertion pass. The iPad post-search steps remain unverified by this run.

The iPhone job in the same run passed all three workflows. Inspected screenshots
show the [selected Garage / Garage bin destination](iphone-expanded-selected-location.png)
and the [preserved proposal after closing and returning](iphone-expanded-returned-proposal.png).
Expanded app labels fit the visible proposal and its completion actions at normal
text size; user-provided names remain unchanged. This is simulator fixture evidence,
not connected voice, physical-device or full-app localization acceptance.
