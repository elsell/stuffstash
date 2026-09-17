# Single-activation search acceptance —35232422023

Source fa798486e15e218c716c3365bf2a22dbfce769ab; runner-only attachment-order
candidate. Phone105239676997 passes7/8; iPad105239676456 passes6/8.
Exact outcomes are in native-search-single-activation-352324-results.csv.

Both devices pass ordinary/preconfigured Place, managed/static placement,
Settings and Expiration. Phone also passes voice location selection. Browse tags
fails keyboard readiness on phone and action/accessory clearance on iPad.
Voice location on iPad fails its post-Clear availability predicate.

Three final-state captures and hierarchies were visually reviewed and retained
in evidence/ with352324 suffixes. Unlike352287, phone Browse shows a focused
search field and visible keyboard without an AutoFill menu. Removing the redundant
tap therefore does not fully resolve the readiness failure. iPad Browse visibly
has both actions above the accessory: AX action bottom682.5, dismiss top739.
The predicate additionally requires hit-test availability; the screenshot and
hierarchy do not establish that component or its earlier state. Do not classify
this sample as confirmed visual overlap or weaken hit-test assertions.

The iPad voice capture shows restored locations and the collapsed header Search
button. Its test still requires hidden search fields to disappear and later taps
a focused field unconditionally. This is the same test-contract inconsistency
already corrected in Place, and requires the same usable-control semantics while
preserving query, selection and navigation checks.

No production dependency patch is promoted. M207 remains open. The Bash observer
slept120 seconds between terminal checks; no jobs were restarted. Range extraction
transferred approximately1.6MB instead of downloading the524MB artifact pair.
