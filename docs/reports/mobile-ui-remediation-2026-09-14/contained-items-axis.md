# Contained items current source review

S102 reviewed atc1c2a22d: AssetContainedWorkspace, AssetDetailView and its route.
No additional confirmed normal-size defect was established in this pass. Native
search placement remains M207; phone success does not override iPad failure.

| Axes | Evidence and limits |
| --- | --- |
| Task, navigation, selection, modality | Rows open asset details; Add here and Move here are commands with destination context. No arbitrary selection modal. Native stack owns Back. |
| Layout, adaptation, typography | One FlatList owns header and rows; rows have88-point minimum height,64-point decorative thumbnails and flexible wrapping text. Real safe-area, tablet and long-text geometry remain unverified. |
| Appearance, imagery | Theme colors and inversion exclusion for images. Missing image uses kind fallback; failed decorative thumbnails have no dedicated fallback but row identity remains in text. No loss of the row task established. |
| Localization, targets, gestures | Labels retain title/path and use English summaries; RTL not verified. Whole row is a button. Native commands provide explicit actions; no required swipe operation. |
| Keyboard, accessibility, motion | List dismisses keyboard and preserves handled taps. Row label includes title/type/path, section titles are headers, chevrons/images decorative. VoiceOver order, keyboard focus and scrolling under Reduce Motion remain native work. |
| Content, search | Location spaces/items are separate sections; other containable assets use a single section. Search applies to loaded names/relative paths at20 entries and clears when the owner changes. Empty search includes a native Clear search command. This is client filtering of supplied contents, not a global search claim. |
| Loading, recovery | Independent contents-query loading and retry coexist with core/photo regions. Missing contents do not become a misleading empty collection. Background fetching does not directly control the pull spinner. |
| Editing, privacy | Row navigation has no draft; parent capability gates Add/Move-here. Server scope/authorization remains authoritative. No new security-boundary claim from reviewing the rendering layer. |
| Notifications, media, lifecycle | No notification or capture entry in rows. Asset navigation and scoped search owners prevent old search callbacks crossing assets; process death, app-switch and session replacement require native acceptance. |

Run350504: Place search passes phone but fails iPad. Both results remain in the
ledger; the source review is complete while platform acceptance is not.
