# iPad native audit — run35140471580

Source a1b827e0, tested merge a14a86aaa33e5a8169feab902988bea15a9bac15;
iPad mini (A17 Pro), job104943442379 completed79/86 tests with7 failures in
3483.193 seconds. Complete log: `/tmp/native351404-ipad-complete.log`.
Artifact10468360267 (1,088,121,846 bytes) is retained only on paul at
`/tmp/native351404-ipad.zip`. Selected captures/hierarchies are retained on both
hosts at `/tmp/ipad351404-selected`. Both onboarding jobs pass separately.

## Normal-size results

- Controlled address retains `hexample.invalid`, controlled name without accessory
  retains `Nati draft nameve`, and ordinary controlled name retains `Nve draft name`.
  Inspected4D2BB158-C53B-490D-A2AB-3384AAF9067B.png shows the last value in both the
  input and observed-value mirror. No trace attachment was exported; the capture
  mechanism correction and its evidence are described in native-phone-351404.md.
- Ordinary color-picker opening fails again; M51 remains open.
- Add in a navigation stack stops at keyboard readiness before typing. Its later
  final screenshot4FD165E1-7515-421E-B14F-8F7AE1AE844A.png shows an empty focused
  field and visible keyboard. That capture does not establish earlier readiness
  or prove that the unexecuted typing/save workflow passed.
- Unfinished-tag collapse/reopen retention passes with the bounded observation
  correction. Home return recovery, Settings search, preconfigured Place search,
  and managed enable/title/native-action search comparison pass.
- Ordinary seeded name, native-default name, paced controlled name, seeded address
  and seeded address without accessory pass. These contrasts inform diagnosis;
  none proves a product-wide text-entry fix.

Edit metadata and Edit tags enlarged-text journeys are the other two failures.
Move Here enlarged-text recovery passes this run; earlier failures remain recorded.
Normal-size work retains priority. This remains a failed full native release gate.

Follow-up at a01fc760: focused text-entry run35148050909 and full
run35148054814 are active/queued. They preserve exact text acceptance and add
reliable diagnostic-command delivery. No application text-entry change is claimed.

Storage update (September 16): the downloaded full fixture ZIP copies for runs
35131892834 and35140471580 were removed from paul to recover temporary space.
GitHub confirms the original artifacts remain unexpired through September 30.
Retained selected captures, hierarchies and logs are unchanged.
