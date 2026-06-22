# UI Design Workshop Spec

## Purpose

Stuff Stash needs a repeatable UI design workflow before the web frontend expands beyond the current tracer-bullet screens.

This spec defines the workshop process for brainstorming, designing, validating, and promoting production-grade web UI direction.

## Scope

This spec covers:

- The design workflow used before substantial web UI implementation.
- The required decision gates with the product owner.
- The role of realistic SvelteKit UI candidates with mock data.
- Accessibility, mobile, desktop, and usability review expectations.
- The engineering principles that UI designs and candidates must preserve.
- The project-scoped Codex skill that guides this workflow.

This spec does not define final screen layouts, final navigation, mobile app UI, backend API behavior, or a complete design token package.

## Decisions

- The project must maintain a repo-scoped Codex skill at `.codex/skills/stuffstash-ui-design`.
- Use the skill when brainstorming, designing, reviewing, or implementing substantial Stuff Stash web UI workflows.
- The workflow must begin from current specs before UI decisions are made.
- The product owner must be used as a sounding board and decision gate before implementation direction is treated as approved.
- Design review must produce real artifacts, not only prose.
- For meaningful UI direction, the artifact must be a working SvelteKit candidate with realistic mock data.
- Temporary UI candidates may live under `/tmp` or another explicitly temporary directory.
- Temporary UI candidates must use the production-intended web stack unless the product owner approves an exception:
  - SvelteKit.
  - Svelte-compatible shadcn-style primitives.
  - Stuff Stash brand direction.
  - Responsive mobile and desktop layouts.
- Temporary candidates must not be treated as disposable prototypes if their purpose is design validation. They should be production-shaped enough to expose real layout, state, accessibility, and implementation tradeoffs.
- The current `apps/web` tracer-bullet UI remains disposable until a design workshop approves a replacement direction.
- Approved UI direction must be reflected in specs before promotion into `apps/web`.
- UI candidates and promoted web implementation must preserve the same engineering discipline as the API:
  - Domain-driven language and typed concepts.
  - Hexagonal boundaries between UI, frontend domain models, generated API clients, adapters, authentication, runtime configuration, and observability.
  - Configuration through runtime or environment-backed boundaries rather than hard-coded environment values.
  - Enumerations or typed value objects instead of loose strings for meaningful domain states and roles.
  - DRY composition where shared behavior is real, while avoiding premature abstractions.
  - Domain-oriented observability through explicit ports or injectable helpers, not raw `print`, `console.log`, or ad hoc logging in production paths.
  - Small, cohesive files with clean separation of concerns. Catch-all "god" files are not allowed.

## Workshop Flow

Each substantial UI workflow must pass through these phases:

1. Grounding.
   - Read relevant platform, brand, client technology, and domain specs.
   - Identify the user workflow, user goal, domain language, and current implementation constraints.
2. Product framing.
   - Define the primary user, job to be done, entry points, success state, failure states, and non-goals.
   - Ask the product owner to confirm or correct the framing.
3. Interaction model.
   - Define routes, page states, major components, information hierarchy, and mobile and desktop behavior.
   - Ask the product owner to choose or revise the direction before building the candidate.
4. Real candidate implementation.
   - Build a working SvelteKit candidate with realistic mock data in a temporary directory.
   - Include loading, empty, error, permission, and dense-data states when they affect the workflow.
   - Run the candidate locally when possible and provide the review URL.
5. Adversarial review.
   - Review the candidate through separate usability, accessibility, mobile, information architecture, visual system, and implementation feasibility lenses.
   - Synthesize the findings into must-fix, should-fix, and acceptable tradeoff groups.
6. Iteration gate.
   - Ask the product owner to decide whether to iterate, redirect, or approve the direction.
7. Promotion.
   - Update the relevant specs first.
   - Implement the approved direction in `apps/web`.
   - Run the relevant tests, accessibility checks, responsive verification, and the code critic agent.

## Review Standards

The workshop must evaluate:

- User mental model.
- Task clarity.
- Navigation and wayfinding.
- Recognition over recall.
- Error prevention and recovery.
- Saved-state and undo clarity.
- Keyboard and screen-reader accessibility.
- WCAG 2.2-aligned contrast, focus, target-size, and form behavior.
- Mobile ergonomics.
- Desktop information density.
- Performance-sensitive rendering choices.
- Use of shadcn-style primitives for generic controls.
- Alignment with Stuff Stash brand guidance.
- Preservation of frontend domain boundaries, typed vocabulary, configuration boundaries, and domain-oriented observability.
- File organization that keeps each file focused on one concern or a few small, tightly related concerns.
- Product copy that makes state and actions clear without narrating, pitching, or explaining the interface to the user.

Named thought leaders may be used as inspiration for review lenses, but the workflow must state findings as concrete product and interface critique rather than pretending to speak as a specific person.

## Verification

Before a workshop result is accepted:

- The temporary candidate must run locally unless blocked by dependency or environment constraints.
- Desktop and mobile viewports must be reviewed.
- Critical interactive states must be exercised.
- Accessibility issues found during review must be either fixed or recorded as explicit tradeoffs.
- The product owner must approve the direction before repo implementation begins.

## Relationship To Other Specs

- `specs/platform/brand-guidelines.spec.md` owns brand intent.
- `specs/platform/client-technology.spec.md` owns web technology direction.
- `specs/platform/web-frontend-tracer-bullet.spec.md` owns the current tracer-bullet constraints and explains why the current screens are disposable.
- Future workflow-specific UI specs may own final screen behavior once approved.
