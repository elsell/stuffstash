# Native search acceptance — current diagnosis

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
production behavior change is allowed to make this assertion pass. iPhone and the
remaining post-search steps still need their own result.
