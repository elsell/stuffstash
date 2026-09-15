# Conversation response — 24 source axes

S121 at722ca46c. Reviewed current and historical response composition,
VoiceConversationExchange/References, VoiceResponseEntityText/Links/Markdown,
VoiceResultRailNavigation, plan history summary, retry state mapping and the
controller's plan-keyed retry lookup. This is source evidence, not full voice QA.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Read the answer, inspect referenced inventory assets, and review saved plan history. Photo retry is a distinct command after partial media failure. |
| Navigation | Inline references and cards open native asset details after pausing media and dismissing the transient sheet. M171: late pause completion lacks visit ownership. |
| Selection | Text is selectable; references are links/navigation, not editable value choices. Duplicate titles use disambiguating fallback references. |
| Modality | Response reading stays in the conversation. Native asset navigation follows sheet dismissal; race and actual transition need acceptance. |
| Layout | Content-sized response text and bounded horizontal rail; shared parent conversation scroll. Long responses, keyboard and footer reachability remain native work. |
| Adaptation | Fixed264-point card interval with viewport padding. Narrow windows, tablet and RTL snap behavior require runtime checks. |
| Typography | Markdown is parsed into native text spans, including lists and emphasis; no WebView. Normal-size long/link-heavy paragraphs and later enlarged text need acceptance. |
| Appearance | Semantic response surfaces; M170 uses native rail and retry commands. Message/link/card appearance remains unverified on device. |
| Localization | Original response text preserved; link tests cover Unicode offsets and case distinctions. English status/action copy; RTL remains open. |
| Imagery | Real card photos through scoped detail query; shared fallback when no photo. No provider-generated image URL is loaded here. |
| Targets | Native Previous/Next plus shared cards, inline text links and fallback actions. Actual inline link hit regions and VoiceOver alternatives need runtime evidence. |
| Gestures | Rail supports swipe/snap and explicit Previous/Next; native text selection/copy. No autoplay. |
| Keyboard | Response is not an editor. Link navigation dismisses keyboard and pauses audio; asynchronous navigation ownership is M171. |
| Accessibility | Link labels, fallback disambiguation, polite response text updates and position count. Reading order across streaming text/cards and announcement verbosity remain open. |
| Motion | Rail commands respect reduced motion; restored offsets are not animated. Native transitions and long response updates unverified. |
| Content | User transcript is separate from assistant answer; historical plan commands/status remain readable. Response parsing tests preserve content and remove markdown syntax appropriately. |
| Search | Response references are results, not local search controls. Rail deduplicates by asset ID and bounds cards to12; remaining references may appear as text links. |
| Loading | Card detail reads show loading/unavailable text. Parent owns request progress. Native loading/recovery layout needs verification. |
| Recovery | Photo retry is plan-keyed, single-flight per plan and updates matching current/history exchange. Missing retry payload returns a non-retry status; late results are lifetime-scoped. |
| Editing | Historical response is read-only, selectable. Proposal editing is a separate surface and is not certified here. |
| Privacy | Asset reads are inventory-scoped; scope changes clear conversation ownership. Client link presentation is not authorization proof. |
| Notifications | No response-owned notification action; interruption belongs to lifecycle review. |
| Media | Historical photo retry retains original plan context. Asset-link opening pauses media. Physical audio and upload behavior require separate runtime evidence. |
| Lifecycle | History/rail offsets retained above sheet, bounded history20. M171 affects late asset navigation; native background/return acceptance remains open. |

All29 focused response/link/markdown/history/retry-state and rail tests pass on
paul. This does not reproduce the M171 race yet and does not establish native
layout, audio, accessibility or real server authorization acceptance.
