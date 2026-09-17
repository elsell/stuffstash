# Paired search-order comparison —35225554369

Source955f645d85635bfee73fda286ae7fa0eed012090 differs from baseline4420dc30
only in report files. App, native test and workflow code are identical. Selection
search-placement-order applies the runner-only native ordering transformation.
Phone105216113714 passes7/7, versus baseline6/7. iPad105216113286 passes6/7,
versus baseline7/7. Exact case outcomes are in native-search-paired-352255-results.csv.

The baseline phone Settings failure is now visually confirmed: persistent bottom
Search tags field at(33,803,336,38), no header Search button. Its header spans
Y62–116. The candidate's phone Settings journey passes again, strengthening the
placement hypothesis across two candidate samples. Retained baseline image and
hierarchy are in evidence/phone-settings-baseline-352224.*.

Candidate iPad Expiration now passes the corrected Search-to-keyboard flow.
Preconfigured Place instead fails after Clear text: results restore Tool0, but the
five-second predicate requiring a hittable field or absent-field/hittable-Search
button does not complete. The reviewed final capture has the restored20-item list,
a collapsed Search icon in the header and no SearchField in its hierarchy. This
supports eventual restoration, not timely interaction readiness or completion of
subsequent re-entry/navigation checks. Do not mark the entire workflow accepted or
relax the predicate based on a later capture. Retain image/hierarchy as
ipad-place-cleared-order-352255.*.

No production patch is promoted yet. The next investigation is clear/re-entry
lifecycle and repeated navigation-item configuration, keeping the existing query
and return assertions. Source attachment ordering remains a plausible candidate,
not a complete solution proven by these intermittent samples. Both jobs terminated
normally; a sleeping observer collected results without restarts.
