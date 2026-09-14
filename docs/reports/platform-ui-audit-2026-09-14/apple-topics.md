# Apple HIG applicability ledger

Snapshot: 2026-09-14. 170 topic/category documents. “Applicable” means
review scope, not a compliance pass. “Conditional” records platform/feature
conditions. “N/A” requires reassessment when product scope changes. Runtime
verification is pending throughout; see the main report for source findings.

The index was traversed through Apple DocC `topicSections`, including nested
component categories. API documentation linked from articles is not included.

| Apple topic | Scope | Product mapping and evidence |
| --- | --- | --- |
| [Accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Action button](https://developer.apple.com/design/human-interface-guidelines/action-button) | Conditional | Optional system shortcut integration; no requirement to implement. |
| [Action sheets](https://developer.apple.com/design/human-interface-guidelines/action-sheets) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Activity rings](https://developer.apple.com/design/human-interface-guidelines/activity-rings) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Activity views](https://developer.apple.com/design/human-interface-guidelines/activity-views) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [AirPlay](https://developer.apple.com/design/human-interface-guidelines/airplay) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Alerts](https://developer.apple.com/design/human-interface-guidelines/alerts) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Always On](https://developer.apple.com/design/human-interface-guidelines/always-on) | Conditional | Relevant only if an always-visible extension ships. |
| [App Clips](https://developer.apple.com/design/human-interface-guidelines/app-clips) | N/A | No corresponding system/media connectivity integration is shipped. |
| [App icons](https://developer.apple.com/design/human-interface-guidelines/app-icons) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [App Shortcuts](https://developer.apple.com/design/human-interface-guidelines/app-shortcuts) | Conditional | Optional automation entrypoint; not currently established. |
| [Apple Pay](https://developer.apple.com/design/human-interface-guidelines/apple-pay) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Apple Pencil and Scribble](https://developer.apple.com/design/human-interface-guidelines/apple-pencil-and-scribble) | Conditional | iPad text entry interoperability to verify, no custom drawing needed. |
| [Augmented reality](https://developer.apple.com/design/human-interface-guidelines/augmented-reality) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Boxes](https://developer.apple.com/design/human-interface-guidelines/boxes) | Conditional | Web grouping and native form sections relevant; avoid importing macOS chrome. |
| [Branding](https://developer.apple.com/design/human-interface-guidelines/branding) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Buttons](https://developer.apple.com/design/human-interface-guidelines/buttons) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Camera Control](https://developer.apple.com/design/human-interface-guidelines/camera-control) | Conditional | Camera acquisition is relevant; dedicated device control integration unverified. |
| [CareKit](https://developer.apple.com/design/human-interface-guidelines/carekit) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [CarPlay](https://developer.apple.com/design/human-interface-guidelines/carplay) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Charting data](https://developer.apple.com/design/human-interface-guidelines/charting-data) | Conditional | Apply if quantitative charts are shipped; do not add visualizations just for coverage. |
| [Charts](https://developer.apple.com/design/human-interface-guidelines/charts) | Conditional | Apply if quantitative charts are shipped; import counts are not automatically charts. |
| [Collaboration and sharing](https://developer.apple.com/design/human-interface-guidelines/collaboration-and-sharing) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Collections](https://developer.apple.com/design/human-interface-guidelines/collections) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Color wells](https://developer.apple.com/design/human-interface-guidelines/color-wells) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Color](https://developer.apple.com/design/human-interface-guidelines/color) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Column views](https://developer.apple.com/design/human-interface-guidelines/column-views) | Conditional | Containment hierarchy merits comparison; macOS column view is not automatically a phone pattern. |
| [Combo boxes](https://developer.apple.com/design/human-interface-guidelines/combo-boxes) | Conditional | Web searchable choices may fit; do not copy macOS control appearance. |
| [Complications](https://developer.apple.com/design/human-interface-guidelines/complications) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Components](https://developer.apple.com/design/human-interface-guidelines/components) | Index | Coverage category; children assessed individually. |
| [Content](https://developer.apple.com/design/human-interface-guidelines/content) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Context menus](https://developer.apple.com/design/human-interface-guidelines/context-menus) | Conditional | Assess existing overflow affordances; long-press menus are optional. |
| [Controls](https://developer.apple.com/design/human-interface-guidelines/controls) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Dark Mode](https://developer.apple.com/design/human-interface-guidelines/dark-mode) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Design principles](https://developer.apple.com/design/human-interface-guidelines/design-principles) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Designing for games](https://developer.apple.com/design/human-interface-guidelines/designing-for-games) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Designing for iOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-ios) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Designing for iPadOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-ipados) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Designing for iPhone Duo](https://developer.apple.com/design/human-interface-guidelines/designing-for-iphone-duo) | Conditional | Foldable/adaptive layout guidance; device support and runtime verification not established. |
| [Designing for macOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-macos) | Conditional | Web on desktop is in scope; no native macOS client declared. |
| [Designing for tvOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-tvos) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Designing for visionOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-visionos) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Designing for watchOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-watchos) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Digit entry views](https://developer.apple.com/design/human-interface-guidelines/digit-entry-views) | N/A | No corresponding product control or media task is currently shipped. |
| [Digital Crown](https://developer.apple.com/design/human-interface-guidelines/digital-crown) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Disclosure controls](https://developer.apple.com/design/human-interface-guidelines/disclosure-controls) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Dock menus](https://developer.apple.com/design/human-interface-guidelines/dock-menus) | N/A | No native Mac client is shipped; web uses browser equivalents. |
| [Drag and drop](https://developer.apple.com/design/human-interface-guidelines/drag-and-drop) | Conditional | Potential media/containment enhancement; no requirement to add. |
| [Edit menus](https://developer.apple.com/design/human-interface-guidelines/edit-menus) | Conditional | Retain native text editing; custom editing commands only when needed. |
| [Entering data](https://developer.apple.com/design/human-interface-guidelines/entering-data) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Eyes](https://developer.apple.com/design/human-interface-guidelines/eyes) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Feedback](https://developer.apple.com/design/human-interface-guidelines/feedback) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [File management](https://developer.apple.com/design/human-interface-guidelines/file-management) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Focus and selection](https://developer.apple.com/design/human-interface-guidelines/focus-and-selection) | Applicable | All interactive surfaces; keyboard/assistive-tech/pointer runtime checks pending (F09). |
| [Foundations](https://developer.apple.com/design/human-interface-guidelines/foundations) | Index | Coverage category; children assessed individually. |
| [Game Center](https://developer.apple.com/design/human-interface-guidelines/game-center) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Game controls](https://developer.apple.com/design/human-interface-guidelines/game-controls) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Gauges](https://developer.apple.com/design/human-interface-guidelines/gauges) | N/A | No corresponding product control or media task is currently shipped. |
| [Generative AI](https://developer.apple.com/design/human-interface-guidelines/generative-ai) | Applicable | Notifications/auth/AI voice; inspect user control/data-use and verify system entrypoints. |
| [Gestures](https://developer.apple.com/design/human-interface-guidelines/gestures) | Applicable | All interactive surfaces; keyboard/assistive-tech/pointer runtime checks pending (F09). |
| [Getting started](https://developer.apple.com/design/human-interface-guidelines/getting-started) | Index | Coverage category; children assessed individually. |
| [Going full screen](https://developer.apple.com/design/human-interface-guidelines/going-full-screen) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Gyroscope and accelerometer](https://developer.apple.com/design/human-interface-guidelines/gyro-and-accelerometer) | N/A | No corresponding system/media connectivity integration is shipped. |
| [HealthKit](https://developer.apple.com/design/human-interface-guidelines/healthkit) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Home Screen quick actions](https://developer.apple.com/design/human-interface-guidelines/home-screen-quick-actions) | Conditional | Optional entrypoint; not currently established. |
| [HomeKit](https://developer.apple.com/design/human-interface-guidelines/homekit) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [iCloud](https://developer.apple.com/design/human-interface-guidelines/icloud) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Icons](https://developer.apple.com/design/human-interface-guidelines/icons) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [ID Verifier](https://developer.apple.com/design/human-interface-guidelines/id-verifier) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Image views](https://developer.apple.com/design/human-interface-guidelines/image-views) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Image wells](https://developer.apple.com/design/human-interface-guidelines/image-wells) | Conditional | Media acquisition applies; macOS image-well styling is not a mobile requirement. |
| [Images](https://developer.apple.com/design/human-interface-guidelines/images) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [iMessage apps and stickers](https://developer.apple.com/design/human-interface-guidelines/imessage-apps-and-stickers) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Immersive experiences](https://developer.apple.com/design/human-interface-guidelines/immersive-experiences) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [In-app purchase](https://developer.apple.com/design/human-interface-guidelines/in-app-purchase) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Inclusion](https://developer.apple.com/design/human-interface-guidelines/inclusion) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Inputs](https://developer.apple.com/design/human-interface-guidelines/inputs) | Applicable | All interactive surfaces; keyboard/assistive-tech/pointer runtime checks pending (F09). |
| [Keyboards](https://developer.apple.com/design/human-interface-guidelines/keyboards) | Applicable | All interactive surfaces; keyboard/assistive-tech/pointer runtime checks pending (F09). |
| [Labels](https://developer.apple.com/design/human-interface-guidelines/labels) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Launching](https://developer.apple.com/design/human-interface-guidelines/launching) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Layout and organization](https://developer.apple.com/design/human-interface-guidelines/layout-and-organization) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Layout](https://developer.apple.com/design/human-interface-guidelines/layout) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Lists and tables](https://developer.apple.com/design/human-interface-guidelines/lists-and-tables) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Live Activities](https://developer.apple.com/design/human-interface-guidelines/live-activities) | Conditional | Optional background task/status extension; not currently established. |
| [Live Photos](https://developer.apple.com/design/human-interface-guidelines/live-photos) | Conditional | Photo acquisition may encounter this media; dedicated Live Photo behavior not established. |
| [Live-viewing apps](https://developer.apple.com/design/human-interface-guidelines/live-viewing-apps) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Loading](https://developer.apple.com/design/human-interface-guidelines/loading) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Lockups](https://developer.apple.com/design/human-interface-guidelines/lockups) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Mac Catalyst](https://developer.apple.com/design/human-interface-guidelines/mac-catalyst) | N/A | No native Mac client is shipped; web uses browser equivalents. |
| [Machine learning](https://developer.apple.com/design/human-interface-guidelines/machine-learning) | Applicable | Notifications/auth/AI voice; inspect user control/data-use and verify system entrypoints. |
| [Managing accounts](https://developer.apple.com/design/human-interface-guidelines/managing-accounts) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Managing notifications](https://developer.apple.com/design/human-interface-guidelines/managing-notifications) | Applicable | Notifications/auth/AI voice; inspect user control/data-use and verify system entrypoints. |
| [Maps](https://developer.apple.com/design/human-interface-guidelines/maps) | N/A | Containment Map is a hierarchy, not geographical mapping. |
| [Materials](https://developer.apple.com/design/human-interface-guidelines/materials) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Menus and actions](https://developer.apple.com/design/human-interface-guidelines/menus-and-actions) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Menus](https://developer.apple.com/design/human-interface-guidelines/menus) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Modality](https://developer.apple.com/design/human-interface-guidelines/modality) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Motion](https://developer.apple.com/design/human-interface-guidelines/motion) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Multitasking](https://developer.apple.com/design/human-interface-guidelines/multitasking) | Conditional | iPad resizing/background return relevant; multiwindow feature not established. |
| [Navigation and search](https://developer.apple.com/design/human-interface-guidelines/navigation-and-search) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Nearby interactions](https://developer.apple.com/design/human-interface-guidelines/nearby-interactions) | N/A | No corresponding system/media connectivity integration is shipped. |
| [NFC](https://developer.apple.com/design/human-interface-guidelines/nfc) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Notifications](https://developer.apple.com/design/human-interface-guidelines/notifications) | Applicable | Notifications/auth/AI voice; inspect user control/data-use and verify system entrypoints. |
| [Offering help](https://developer.apple.com/design/human-interface-guidelines/offering-help) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Onboarding](https://developer.apple.com/design/human-interface-guidelines/onboarding) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Ornaments](https://developer.apple.com/design/human-interface-guidelines/ornaments) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Outline views](https://developer.apple.com/design/human-interface-guidelines/outline-views) | Conditional | Hierarchy semantics relevant to containment; native macOS control not required. |
| [Page controls](https://developer.apple.com/design/human-interface-guidelines/page-controls) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Panels](https://developer.apple.com/design/human-interface-guidelines/panels) | Conditional | Supplemental wider-layout tools may fit; not a default phone modal. |
| [Path controls](https://developer.apple.com/design/human-interface-guidelines/path-controls) | Conditional | Ancestor context is relevant; native macOS path control is not required. |
| [Patterns](https://developer.apple.com/design/human-interface-guidelines/patterns) | Index | Coverage category; children assessed individually. |
| [Photo editing](https://developer.apple.com/design/human-interface-guidelines/photo-editing) | Conditional | Acquisition/crop behavior relevant; no standalone editing feature required. |
| [Pickers](https://developer.apple.com/design/human-interface-guidelines/pickers) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Playing audio](https://developer.apple.com/design/human-interface-guidelines/playing-audio) | Applicable | Photos/files/sharing/voice; source entrypoints surveyed, system/device checks pending. |
| [Playing haptics](https://developer.apple.com/design/human-interface-guidelines/playing-haptics) | Conditional | Check any shipped feedback; haptics are optional and cannot carry sole meaning. |
| [Playing video](https://developer.apple.com/design/human-interface-guidelines/playing-video) | N/A | No corresponding product control or media task is currently shipped. |
| [Pointing devices](https://developer.apple.com/design/human-interface-guidelines/pointing-devices) | Applicable | All interactive surfaces; keyboard/assistive-tech/pointer runtime checks pending (F09). |
| [Pop-up buttons](https://developer.apple.com/design/human-interface-guidelines/pop-up-buttons) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Popovers](https://developer.apple.com/design/human-interface-guidelines/popovers) | Conditional | Appropriate for anchored choices on wider layouts; native iPad behavior unverified. |
| [Presentation](https://developer.apple.com/design/human-interface-guidelines/presentation) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Printing](https://developer.apple.com/design/human-interface-guidelines/printing) | Conditional | Optional export/print workflow; not currently established. |
| [Privacy](https://developer.apple.com/design/human-interface-guidelines/privacy) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Progress indicators](https://developer.apple.com/design/human-interface-guidelines/progress-indicators) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Pull-down buttons](https://developer.apple.com/design/human-interface-guidelines/pull-down-buttons) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Rating indicators](https://developer.apple.com/design/human-interface-guidelines/rating-indicators) | N/A | No corresponding product control or media task is currently shipped. |
| [Ratings and reviews](https://developer.apple.com/design/human-interface-guidelines/ratings-and-reviews) | Conditional | Optional store review request; no need to add to satisfy HIG. |
| [Remotes](https://developer.apple.com/design/human-interface-guidelines/remotes) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [ResearchKit](https://developer.apple.com/design/human-interface-guidelines/researchkit) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Right to left](https://developer.apple.com/design/human-interface-guidelines/right-to-left) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Scroll views](https://developer.apple.com/design/human-interface-guidelines/scroll-views) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Search fields](https://developer.apple.com/design/human-interface-guidelines/search-fields) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Searching](https://developer.apple.com/design/human-interface-guidelines/searching) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Segmented controls](https://developer.apple.com/design/human-interface-guidelines/segmented-controls) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Selection and input](https://developer.apple.com/design/human-interface-guidelines/selection-and-input) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Settings](https://developer.apple.com/design/human-interface-guidelines/settings) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [SF Symbols](https://developer.apple.com/design/human-interface-guidelines/sf-symbols) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [SharePlay](https://developer.apple.com/design/human-interface-guidelines/shareplay) | N/A | No corresponding system/media connectivity integration is shipped. |
| [ShazamKit](https://developer.apple.com/design/human-interface-guidelines/shazamkit) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Sheets](https://developer.apple.com/design/human-interface-guidelines/sheets) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Sidebars](https://developer.apple.com/design/human-interface-guidelines/sidebars) | Conditional | Consider for wider task hierarchy; not mandatory just because iPad is supported. |
| [Sliders](https://developer.apple.com/design/human-interface-guidelines/sliders) | Conditional | Candidate for continuous values, not arbitrary discrete choices. |
| [Snippets](https://developer.apple.com/design/human-interface-guidelines/snippets) | N/A | No corresponding system/media connectivity integration is shipped. |
| [Spatial layout](https://developer.apple.com/design/human-interface-guidelines/spatial-layout) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Split views](https://developer.apple.com/design/human-interface-guidelines/split-views) | Conditional | Consider for simultaneous hierarchy/detail tasks; do not add without a user need. |
| [Status bars](https://developer.apple.com/design/human-interface-guidelines/status-bars) | Conditional | Respect system status/safe areas; no custom status bar needed. |
| [Status](https://developer.apple.com/design/human-interface-guidelines/status) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Steppers](https://developer.apple.com/design/human-interface-guidelines/steppers) | Conditional | Candidate for bounded numeric input when useful; custom reminder days may need direct typing. |
| [System experiences](https://developer.apple.com/design/human-interface-guidelines/system-experiences) | Index | Coverage category; children assessed individually. |
| [Tab bars](https://developer.apple.com/design/human-interface-guidelines/tab-bars) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Tab views](https://developer.apple.com/design/human-interface-guidelines/tab-views) | Conditional | Web peer panels relevant; do not confuse them with primary iOS tab navigation. |
| [Tap to Pay on iPhone](https://developer.apple.com/design/human-interface-guidelines/tap-to-pay-on-iphone) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Technologies](https://developer.apple.com/design/human-interface-guidelines/technologies) | Index | Coverage category; children assessed individually. |
| [Text fields](https://developer.apple.com/design/human-interface-guidelines/text-fields) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Text views](https://developer.apple.com/design/human-interface-guidelines/text-views) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [The menu bar](https://developer.apple.com/design/human-interface-guidelines/the-menu-bar) | Conditional | Native iPad/desktop command coverage depends on supported environments; web is not a macOS app. |
| [Toggles](https://developer.apple.com/design/human-interface-guidelines/toggles) | Applicable | Filters/Add/edit/settings; F01–F04/F10; source surveyed, native interaction checks pending. |
| [Token fields](https://developer.apple.com/design/human-interface-guidelines/token-fields) | Conditional | Filter tokens may help web search; macOS token control is not an iOS requirement. |
| [Toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars) | Applicable | Home/Browse/Map/detail/settings; source surveyed, native interaction matrix pending. |
| [Top Shelf](https://developer.apple.com/design/human-interface-guidelines/top-shelf) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Typography](https://developer.apple.com/design/human-interface-guidelines/typography) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
| [Undo and redo](https://developer.apple.com/design/human-interface-guidelines/undo-and-redo) | Applicable | Onboarding/editors/feedback/switching; F05–F08; runtime verification pending. |
| [Virtual keyboards](https://developer.apple.com/design/human-interface-guidelines/virtual-keyboards) | Applicable | All interactive surfaces; keyboard/assistive-tech/pointer runtime checks pending (F09). |
| [VoiceOver](https://developer.apple.com/design/human-interface-guidelines/voiceover) | Applicable | Notifications/auth/AI voice; inspect user control/data-use and verify system entrypoints. |
| [Wallet](https://developer.apple.com/design/human-interface-guidelines/wallet) | N/A | No corresponding commerce, health, home-device, or identity-verification feature is shipped. |
| [Watch faces](https://developer.apple.com/design/human-interface-guidelines/watch-faces) | N/A | No tvOS/watchOS/visionOS/game client or extension is shipped. |
| [Web views](https://developer.apple.com/design/human-interface-guidelines/web-views) | Applicable | Notifications/auth/AI voice; inspect user control/data-use and verify system entrypoints. |
| [Widgets](https://developer.apple.com/design/human-interface-guidelines/widgets) | Conditional | Optional extension; not currently established. |
| [Windows](https://developer.apple.com/design/human-interface-guidelines/windows) | Conditional | iPad window resizing is relevant; separate-window capability not established. |
| [Workouts](https://developer.apple.com/design/human-interface-guidelines/workouts) | N/A | No corresponding product control or media task is currently shipped. |
| [Writing](https://developer.apple.com/design/human-interface-guidelines/writing) | Applicable | All user-facing surfaces; inspect semantics/adaptation and validate rendering on device. |
