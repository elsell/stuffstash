---
name: stuffstash-ui-design
description: Design, review, audit, and implement Stuff Stash mobile and web UI using appropriate platform interactions. Use for screens, controls, navigation, forms, accessibility, and whole-product UI audits; includes iOS HIG review and SvelteKit candidate workflows.
---

# Stuff Stash UI Design

Choose the familiar platform interaction before its implementation. A native
component can implement the wrong interaction; a custom component can resemble a
native control without providing its behavior. Review both questions independently.

## Grounding and mode

Read `specs/platform/platform-interaction-review.spec.md`, the relevant workflow
spec, and the applicable brand/client/navigation specs. User instructions and
existing authorization govern scope. Do not turn an audit into UI implementation
or require a new approval for an already authorized change.

- **Design or implementation:** read [platform interaction decisions](references/platform-interaction-decisions.md).
  Establish the task and pattern before routes, screen layout, or component choice.
- **Audit:** also read [whole-product audit](references/platform-audit.md).
  A full audit needs a coverage inventory, not just a list of obvious problems.
- **Substantial new web direction:** use [web workshop](references/web-workshop.md)
  and [SvelteKit candidate workflow](references/sveltekit-candidate-workflow.md).
  A Svelte candidate is not evidence of native mobile behavior.
- **All modes:** use the relevant lenses in [review rubrics](references/ui-review-rubrics.md),
  [product principles](references/stuffstash-ui-principles.md), and
  [engineering principles](references/frontend-engineering-principles.md).

## Decision standard

For each changed interaction, establish:

1. The user's task: navigate, choose a value, issue a command, edit content, or
   inspect status. Do not let an existing generic component answer this question.
2. The expected platform pattern and why it fits option count, descriptions,
   hierarchy, frequency, risk, and available space. Cite relevant current guidance
   when the choice is disputed, novel, or part of an audit.
3. Commit/cancel semantics, return destination, focus, keyboard behavior, and
   loading/error/empty/denied states. Preserve work across accidental dismissal.
4. The existing native adapter or accessible web primitive. Search its consumers
   before extending it. Document concrete limitations before a custom substitute.
5. Evidence that verifies the interaction choice and its actual implementation.

Do not impose a universal menu, bottom button, sheet, or navigation rule from one
screenshot. Small flat single choices usually belong in place; searchable,
hierarchical, descriptive, or complex choices may justify a selection view.

Apple's guidance governs iOS/iPadOS. Android follows Android conventions; web
follows browser semantics and accessible web patterns. Brand styling must not
replace the platform's interaction vocabulary. Preserve explicit user preferences
and record intentional departures as project choices rather than Apple rules.

## Review and completion

Review connected everyday workflows before isolated controls. Prioritize screen
structure/stability, core task/pattern fit, visual coherence, then detailed states
and edge cases. Tests support this experience judgment; easy test coverage must
not set the product work queue. Apply the same standard to all
consumers of shared controls. Spec before code; meaningful tests before changes;
code critic before finalization, per repository instructions.

For native visual or lifecycle changes, verify a named build on a native runtime:
entry, scrolling, keyboard, navigation return, dismissal, and the affected states.
Use narrow and enlarged-text layouts; include iPad when the affected layout ships
there. Simulator checks do not establish physical push, audio, or camera behavior.
If access is unavailable, finish independent work and explicitly report the missing
evidence. Never equate prop tests, a successful archive, or upload with device QA.

Report concrete findings with source locations, user impact, intended pattern,
priority, evidence level, and acceptance scenario. Distinguish Apple guidance,
project preferences, engineering inference, and observed behavior. Report coverage
and omissions; never certify an entire app from sampled screenshots.
