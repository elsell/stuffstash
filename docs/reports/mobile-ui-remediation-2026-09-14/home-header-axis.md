# Home action header — all 24 axes

S066, source checkpoint `3d61163d`. Reviewed HomeNavigationHeader,
HomeHeaderLayout, HomeScreen, NotificationHomeEntry/Bell, the platform header
adapters and Home stack options. The inventory switcher destination and dashboard
content have separate reviews. No new product defect is established by this pass.

The user's order is **Add, Notifications, Profile**. The source preserves it,
omitting Add without create permission and omitting the bell while its selected
scope is unavailable. Profile remains present. Inventory switching is a separate
leading control; the user explicitly accepted that exception to native action
items and its Home-only placement.

Apple recommends system components for evolving navigation materials and warns
that custom backgrounds can interfere with scroll-edge effects
([Adopting Liquid Glass](https://developer.apple.com/documentation/technologyoverviews/adopting-liquid-glass)).
The iOS adapter installs actual native bar items. Home's stack uses the shared
transparent header policy, with a soft scroll edge on iOS26 and system material on
older iOS. This establishes implementation intent, not visible transparency or
control availability on a named device.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Frequent creation, inbox and account destinations use recognizable action symbols. Inventory context has its own selector. Requested ordering is preserved; platform conventions do not establish that ordering independently of the user's preference. |
| Navigation | Add pushes /add, bell opens /notifications, Profile pushes /settings, selector opens /tenant-switcher. Actions stay outside dashboard scroll content. Native return, repeated taps and tab switching remain acceptance work. |
| Selection | Inventory selection opens a separate scoped workspace task, justified by tenant/inventory hierarchy and its management actions. Header itself has no additional flat-choice picker. |
| Modality | Header does not create its own dialog. Destinations use root route presentation policy. Review dismissal in those destinations; do not infer it from these callbacks. |
| Layout | HomeInventoryControlWidth reserves52 points per current action plus group allowance, caps selector at180 and floors it at44. Header items declare44-point width. These are estimates, not measured native bounds; long-name and three-action clearance need captures. |
| Adaptation | Window width and action count recompute selector width. Font scale hides the tenant subtitle above1.05. Normal-size phone/iPad and window reflow remain first; enlarged-text work stays queued. |
| Typography | Inventory/tenant text uses one line each, with truncation and full accessible context. Internal padding is12 horizontal/4 vertical with a10-point gap. Check ordinary long inventory names and whether context remains recognizable. |
| Appearance | iOS uses system bar backgrounds/symbols and a transparent stack; Android uses Compose icon actions and an opaque themed bar. Check scrolling photos/text, light/dark and reduced transparency. Do not impose iOS glass on Android. |
| Localization | Inventory/tenant names are external content. Labels are English and geometry is not RTL-tested. Notification count is capped visually at99+ while the accessible label gives the count. Broader localization findings remain tracked. |
| Imagery | iOS uses plus, bell and person.crop.circle SF Symbols; Android uses native vector resources. Inventory disclosure remains a Lucide chevron in the accepted custom selector. No content photos are loaded by the header. |
| Targets | Selector minimum height44; native item widths44; Compose hosts48square. Source metrics do not establish hit regions or prevent overflow into an ellipsis. Measure every visible action, especially with three actions and long context. |
| Gestures | Every task has an explicit button. No custom swipe or long-press is required. Scroll-edge behavior belongs to the stack and must be observed with actual dashboard scrolling. |
| Keyboard | N/A for direct input: header has no text field. Returning with a keyboard from another route is a lifecycle acceptance case. |
| Accessibility | Selector includes full inventory and tenant names plus the switch action. Native actions receive descriptive labels; bell describes loading/error/count. Verify VoiceOver/TalkBack order, badge announcements and focus on return. |
| Motion | No custom header animation. Shared native material/navigation response should follow system settings; custom selector text still needs reduced-motion/transparency review in context. |
| Content | Small stable action set; no collection/pagination. Inventory and tenant subtitle provide workspace context. Missing dashboard uses Home title and preserves Profile. Native overflow must not hide normally available Add/Profile. |
| Search | N/A: Home has no search button. Browse owns search; this follows the established app structure and user preference. |
| Loading | Add and selector wait for dashboard data; Profile remains available. Notification loading changes its accessible description, not its ability to open the inbox. Background refresh must not show a pull spinner; Home uses usePullRefresh independently. |
| Recovery | Bell opens the inbox even if registration/count fails and attempts refresh on activation. Count failure is described accessibly; cached badge data can remain visible. Missing dashboard recovery is in Home content. No header-specific error modal is introduced. |
| Editing | N/A for drafts: header only enters tasks. Draft protection belongs to Add and destination editors. |
| Privacy | Add follows dashboard canAdd. The shared scoped query hides data on access failure; notification queries use session, tenant and inventory keys. UI permission visibility is not server authorization evidence. |
| Notifications | Bell is an in-app inbox entry and unread indicator. It does not handle OS cold/warm notification taps; S130 must be audited separately. Count polling is30seconds and disabled in background. |
| Media | N/A: no media acquisition or playback in this header. |
| Lifecycle | Shared header hook keeps committed callbacks current and rejects removed, disabled or unmounted actions. Notification component is keyed by service/tenant/inventory. Native tab return and focus during query changes remain open. |

## Native coverage gap

HomeReturnFixture mounts the production HomeScreen, but supplies no notification
action and uses a root audit stack with an ordinary Home route, not the production
Home tab header options. Its return-sheet scenarios therefore cannot establish
three-action geometry or transparent Home scroll edges. Do not promote S066 from
those results. A representative Home header scenario needs all three actions,
production stack appearance, the correct root back-button policy, long context,
scrollable dashboard content, and measured before/after-scroll button bounds.

The current combined mobile suite at5501152d passed remotely, including width
calculation, native action adapter and committed-callback behavior cases. These
checks do not replace that missing native scenario or certify always-visible
buttons. Android runtime remains unavailable.

## Acceptance scenario added

A separate HomeHeaderFixture now supplies three native actions, a long inventory
name and production Home/Expiration row components with enough domain data to
scroll. It uses production header options and hides the fixture root Back button.
The native test asserts content movement, action order/size/hittability, selector
separation and stable header position, retaining before/after screenshots. The
existing Home Return fixture is unchanged in its default mode.

The fixture still omits production tabs and the voice accessory. Its geometry
results will not establish full tab composition or material appearance; screenshots
need review. The new native test is not yet executed, and no runtime status is
promoted. Notification tapping is outside this fixture: its supplied action is a
controlled no-op for header geometry.

Remote TypeScript, both fixture-isolation tests and the structural check pass.
Critic confirmed the fixture limits and requested explicit menu-item reveal plus
scroll-view lookup by Home content; both procedure corrections are included.
Swift compilation and runtime assertions remain for macOS CI.
