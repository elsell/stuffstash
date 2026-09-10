# Mobile conversation lifecycle

This complements `mobile-conversation-interface.spec.md` and governs native
interaction transitions across the provider, controller, and socket adapter.

| Phase | Allowed user actions | Required outcome |
| --- | --- | --- |
| Ready or completed | Type, record, reset, inspect history | One active turn; no duplicate history on startup failure |
| Recording | Finish, switch to typing, dismiss, reset | Capture stops once; unsent recording is never silently submitted |
| Processing | Cancel, dismiss, reset | Late results cannot replace a newer turn |
| Proposed review | Edit, stage photos, approve, reject, dismiss | Exactly one review decision, including rapid repeat taps |
| Decision submitted | Dismiss or explicitly reset | No misleading cancel/undo affordance; retain outcome uncertainty on disconnect |
| Execution confirmed | Inspect saved result; finish photo batch | Publish Saved immediately, before attachment work finishes |
| Photo batch | Observe progress, dismiss; explicit reset stops remaining work | Successful writes remain saved; retries target failed photos only |
| Failed or disconnected | Inspect retained evidence, reset, start another turn | Do not reuse a dead socket or re-submit an accepted message implicitly |

Reset and scope replacement invalidate outstanding photo work:
a request already sent may complete, but no subsequent photo starts and no stale
result may restore cleared retry state or update a newer conversation. Pending
photo payloads for rejected/failed/non-executed plans are released. Retry data
belongs to executed plans with unsuccessful attachments only. Cancelling a later
question preserves earlier committed-plan photo retries. Once execution is
confirmed, a socket error cannot invalidate the independent attachment outcome;
finish that callback before completing the turn. Explicit disposal still wins.

Review submission is guarded synchronously before any asynchronous work. A second
approve/cancel gesture for the same proposal does not send a second decision,
clear its photo metadata, or change its visible status to failure. Local field
validation may restore the proposal for correction.

An idle socket may expire while the user records a follow-up. If its closure is
known before send, submit the captured recording as a fresh turn, showing the new
context boundary. Never automatically replay a request that might have reached
the server.

History contains actual exchanges. Starting capture or failing local readiness
must not duplicate the previous exchange. Preserve accepted text in its message;
restore composer input only when the request failed before acceptance.

Verification uses controlled asynchronous port fakes and the real context and
controller. Cover normal paths and interruption at each await boundary, including
scope changes, rapid gestures, acknowledgement before photo completion, and late
callbacks. Keep authorization enforcement on the server; fresh requests and photo
retries continue using explicit tenant/inventory scope.

## September 10 lifecycle audit evidence

The audit covered capture and typed input, follow-up expiry, review decisions,
execution acknowledgement, staged-photo batches/retries, reset/scope replacement,
connection failure, and history recovery. Regression tests reproduce the observed
follow-up approval rejection and the additional lifecycle defects before their
fixes. The code critic reviewed the implementation and its follow-up corrections.

Remote validation passed 1,132 mobile tests, mobile type checking and structural
checks, plus application and HTTP adapter tests (including existing adversarial
boundary coverage). No build or test ran on the developer's Mac.

Live Google provider evaluation used synthesized recordings and isolated in-memory
inventory fixtures. Human trace review, without an automated judge:

| Scenario and request | Deterministic result / product verdict | Evidence |
| --- | --- | --- |
| Missing bedroom: add a water bottle to the master bedroom | Failed before fix; passed after fix / pass | Empty authorized lookup led directly to a location and dependent item proposal; no writes before approval |
| Existing move: move my cordless drill to the garage | Passed / pass | Retrieved the existing Office drill and Garage, proposed one move using returned IDs |
| Additional item: add another cordless drill to the garage | Passed / pass | Proposed one new item under the retrieved Garage; did not move the existing drill |
| Dependent move: create a blue toolbox in the garage and move my cordless drill into it | Initial speech-provider timeout; isolated rerun passed / pass for planning, timeout retained as reliability evidence | Proposed toolbox under Garage and move of the existing drill using the toolbox command ID |
| Ambiguous move: move my screwdriver to the garage; the one in the kitchen | Passed / pass | Presented Office/Kitchen ambiguity, retained context, proposed the Kitchen screwdriver move |
| Follow-up: where is my cordless drill; what color is it | Passed / pass | Office answer and subsequent red answer both grounded in the retrieved drill, with its asset reference |

Raw provider traces were retained in the remote validation artifacts. The full
live invocation was not wholly green: its speech timeout is not erased by the
successful isolated rerun. These checks do not replace physical-device testing
of microphone capture, photo picker interruption, backgrounding, or sheet layout.
Existing context boundaries after completed writes remain explicit; prior visible
history is not implicitly replayed to a fresh model session.

## In-card save progress

After approval, the action card changes from review to saving, with a native
activity indicator and a concrete label naming the item when there is one change.
Show that confirmation is pending; do not invent a percentage or completed
command count before the server confirms execution. Cancellation of a proposal
uses cancellation language, never saving language.

Once execution is confirmed, mark changed rows complete and keep Saved visible
while photos upload. Publish completed/total photo counts after every attachment
attempt, including retries. The photo progress bar measures successfully attached
photos, not elapsed time or bytes; failed attachments never count toward 100%.
Show the number still needing attention. Retry progress keeps earlier successful
photos in its totals and updates the same card, including archived exchanges.
Use native accessibility progress values and existing theme colors. Keep progress
inside the action card; avoid duplicate spinners in the conversation body and
composer when this card owns progress.
Save labels use the effective reviewed names, including edits made before approval.
Photo counts supplement, rather than replace, actionable upload failure details.
