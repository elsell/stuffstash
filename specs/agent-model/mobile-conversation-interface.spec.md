# Mobile Conversation Interface

## Approved B1 interaction

The native mobile conversation sheet replaces separate transcript and progress sections with a bounded chronological conversation. Each exchange shows the user's words, the assistant response, and compact inline result or action widgets. Prior exchanges remain visible while the user continues. Keep at most 20 completed exchanges in memory; no durable transcript storage. Scope changes and sign-out clear all conversation content and drafts. A fresh server session must be identified as a new context when the previous follow-up window has ended; visible history does not imply the model remembers it.

The native bottom accessory opens the sheet for typing; its microphone starts recording. The sheet uses the app's shared native text input, buttons, photo controls, theme, keyboard handling and router form-sheet presentation. A persistent composer supports keyboard and microphone input. Recording has a single finish/send action. Switching to typing cancels capture without sending; opening asset details stops recording and playback without automatically restarting them. Processing uses one descriptive activity row. Never fabricate a percentage for inference; show percentages only when the upload adapter reports measured bytes.

## Typed transport

The authenticated realtime WebSocket accepts `text.input` with `sessionId`, monotonically increasing `seq`, and nonblank `text` of at most 8,000 Unicode characters. A turn contains text or audio, never both. Typed input bypasses speech recognition and enters the same application conversation, authorization, turn budgets, provider ports, structured response and action-plan review path. The initial protocol and required capabilities remain compatible with audio clients. Text and audio may alternate within the negotiated continuity window. No input is accepted while a proposal awaits approval. Blank, oversized, mixed-mode, stale-sequence and wrong-session input fail safely. Authentication, tenant/inventory access and current permissions are checked identically for both modes.

## Shared session state and navigation

Conversation state lives above the sheet route and includes completed exchanges, current work, unsent text, proposal edits, selected photos, scroll position and result-rail positions. Collapse and native navigation must preserve it. Explicit reset clears it. Requests are single-flight; pending approval cannot be replaced by a new request. Async results from a previous scope cannot populate the new scope.

Only trusted structured asset references become links. Item, container and location mentions use the same entity-aware text renderer. Links dismiss the keyboard, pause microphone/playback, dismiss the conversation sheet, and push the ordinary scoped asset detail route. Native back remains native back; a return-to-conversation control reopens the retained sheet from asset detail, including nested navigation. Processing may finish while details are open. Unsaved proposal entries stay in review and are never routed using invented IDs.

## Rich inventory answers

Responses with authorized asset references show a compact horizontal rail of cards with title, kind, available location context and real inventory thumbnail. Load details and photos through existing scoped query ports; no provider URLs or synthetic photos. Missing or unavailable photos use the shared fallback. Cards support swipe, snap, accessible previous/next controls and a position count without autoplay. A single result and its container/location context must not be presented as multiple matching items. Card presses use the same native asset navigation. Retain rail position on return and bound fetched cards to 12 references.

## Action widgets and accessibility

Proposals retain editable titles, parent selection, photos, risk disclosure, explicit approval/cancellation, execution outcomes and attachment retry. Keep them compact and inline with the assistant exchange. Review decisions remain reachable above the keyboard. Saving must disable duplicate submission and must not imply completion until the API confirms execution. Use native accessibility roles, labels, dynamic text, reduced-motion behavior, light/dark theme, and sufficiently large touch targets.

## Verification and release

Tests cover typed/audio continuity, invalid text frames, unauthorized and cross-scope access, retained conversation/draft state, entity links and bounded result presentation. Run mobile checks, relevant API security tests, structural checks and code-critic review before merging. Build exclusively in CI for this change. Release through the stable-tag workflow and its TestFlight job, then update the Stuff Stash GitOps image pins in `~/code/infra`. Upload success and Apple processing availability are separate release evidence.

### Readable message layout and formatting

Message bubbles size to their text and never grow to fill the scroll viewport.
The conversation list has a bounded, flexible viewport between the fixed header
and composer; long answers and result rails remain reachable by scrolling.
Only the viewport owns vertical scrolling. Response text beside an icon gets its
width from that row, without imposing vertical growth on text used in bubbles.
User messages link only asset names actually mentioned, without adding answer
result buttons to the user's bubble. Assistant messages retain fallback controls
for ambiguous or otherwise unplaced resolved references.

Display responses preserve paragraph and list line breaks up to the API's 1,000
character display-response limit. A shared native text renderer supports paragraphs,
bulleted/numbered lists, headings, bold, emphasis and inline code. Formatting is
presentation only: links still come exclusively from authorized response artifacts;
model-provided URLs do not become navigation targets. User transcripts stay literal.
The same renderer applies to current and previous assistant messages.
