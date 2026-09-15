# Appearance and material audit — September 15

This is a source review of shared appearance ownership and selected consumers,
not native visual acceptance of every appearance cell.

Apple recommends system materials that adapt to accessibility preferences, and
checking dark appearance with Increase Contrast and Reduce Transparency both
separately and together. See [Materials](https://developer.apple.com/design/human-interface-guidelines/materials)
and [Dark Mode](https://developer.apple.com/design/human-interface-guidelines/dark-mode).
These are distinct checks: a passing color ratio does not prove material behavior.

| Owner | Source evidence | Remaining native acceptance |
| --- | --- | --- |
| AppearanceProvider | Resolves saved/system appearance; waits for initial contrast read before hydration; subscribes before the read and preserves newer contrast events. Separate iOS/Android contrast APIs. | Switch system and explicit appearance while sheets are open; live contrast changes; cold start without a wrong-theme frame. |
| Semantic palettes | Text/action/control pairs have numerical contrast coverage, including photo-count scrim against white. Increased contrast strengthens structural and interactive borders. | Actual pairings throughout screens, disabled/pressed states, long text, material composition and high-contrast combinations. Token tests do not validate all consumers. |
| NativeTabHeader and NativeTabs | iOS 26 native scroll-edge effects, older iOS system material, Android opaque header; native bottom accessory owns its material. No app-rendered fake glass found here. | Scroll images beneath chrome with Reduce Transparency/Increase Contrast in both appearances; verify runtime native adaptation instead of adding redundant JS overrides. |
| Refinement counts — M66 | Three adapters duplicated white text on accent: 3.39:1 light, 2.22:1 dark. Shared badge now uses action/onAction; all eight rendered contrast cases pass after failing before the fix. | Native count placement, enlarged text, and Android rendering. iOS expiration uses its own native symbol state and does not consume this badge. |
| Full-screen photo viewer | Intentional neutral-black exception is specified. Custom toolbar still uses an alpha background and metadata opacity; gallery position badge uses semantic scrim. | Inspect over very bright/dark photos, both accessibility preferences, text scaling, removal disabled state. Fixed black alone is not an appearance defect. |
| Tag chips | User color decorates a low-alpha fill/border; text uses semantic foreground. | Composite contrast across allowed colors/surfaces and increased contrast; color cannot be the sole distinguishing information. |

Shared refinement consumers inspected: BrowseHeader, fallback/Android expiration
filter header, and all three NativeRefinementButton implementations. The iOS
expiration header uses a UIBarButtonItem symbol instead. Count remains duplicated
in the parent accessible label and is hidden from separate accessibility focus.

No blanket claim that every alpha color must change with Reduce Transparency:
first distinguish native material, decoration, imagery overlay and status opacity.
The custom photo overlay needs rendered evidence before choosing its correction.
