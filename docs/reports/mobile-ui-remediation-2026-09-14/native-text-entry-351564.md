# Provider-free input comparison — run35156439952

Sourcee83bfc138810e8a7cc3113fe9951c1176dff5e74. Both targets finish11/14 cases:
phone job105001089504 (484 seconds), iPad job105001089799 (510 seconds).
This is diagnostic evidence, not a release workflow or a production correction.

The [iPad configuration attachments](evidence/ipad-provider-omitted-351564.json)
confirm all14 tests actually mounted the root with the keyboard provider omitted.
The fixture root also omits its accessory. Three incorrect final values remain:

- Controlled address: `h/example.invalid` instead of `https://example.invalid`.
- Controlled name without accessory: `N` instead of `Native draft name`.
- Ordinary controlled name: `Nativdraft namee ` instead of `Native draft name`.

Thus removing the provider does not resolve the observed input loss/reordering.
Do not remove it from production or infer the exact native root cause from this
comparison. Both paced comparisons and SwiftUI comparisons pass in this sample;
that does not prove natural input reliability or justify weakening exact-value
acceptance in actual Add/settings/onboarding workflows.

Phone logs record controlled address `h//example.invalid` and both ordinary seeded
name and seeded no-accessory name `Nve draft name`. Its configuration attachments
and event traces have not yet been inspected, so the iPad attachments are the
verified provider-absence evidence. No new production input change is included
in the frozen release batch on the basis of these comparisons.

Logs: `/tmp/native351564-input-phone.log` and `/tmp/native351564-input-ipad.log`.
iPad artifact10473010922 (247MB) is retained only on paul at
`/tmp/native351564-input-ipad.zip`; phone artifact10472970482 (210MB) remains in
GitHub. Only the compact configuration evidence is copied locally. Further native
text investigation stays in the ongoing audit while required batch workflows run.
