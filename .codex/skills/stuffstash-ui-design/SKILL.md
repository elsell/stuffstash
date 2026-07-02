---
name: stuffstash-ui-design
description: Production-grade Stuff Stash web UI brainstorming, design, review, and implementation workflow. Use when Codex is asked to design, redesign, critique, brainstorm, workshop, prototype, build, or implement substantial Stuff Stash web UI surfaces, SvelteKit screens, responsive layouts, user journeys, visual direction, accessibility behavior, or shadcn-svelte component composition.
---

# StuffStash UI Design

## Overview

Use this skill to run a spec-grounded UI design workshop for Stuff Stash before expanding the real web app. Produce concrete design artifacts, build a working SvelteKit candidate when the direction needs review, and use the product owner as the decision gate before promotion into `apps/web`.

## Required References

Read these files before starting substantial UI work:

- `specs/platform/ui-design-workshop.spec.md`
- `specs/platform/brand-guidelines.spec.md`
- `specs/platform/client-technology.spec.md`
- `specs/platform/web-frontend-tracer-bullet.spec.md`
- Relevant domain specs for the workflow being designed

Read bundled references as needed:

- `references/ui-review-rubrics.md` for adversarial review lenses.
- `references/stuffstash-ui-principles.md` for Stuff Stash-specific UI constraints.
- `references/frontend-engineering-principles.md` for domain, architecture, observability, typing, DRY, and configuration expectations.
- `references/sveltekit-candidate-workflow.md` before building a temporary SvelteKit candidate.

## Workflow

### 1. Ground The Work

Identify the workflow being designed and read the relevant specs before proposing UI behavior. Do not invent product behavior, domain concepts, dependencies, or architecture that the specs do not support.

Summarize:

- Primary user and job to be done.
- Entry points and expected success state.
- Known domain constraints.
- Open questions that affect design direction.
- Whether the current request needs a real SvelteKit candidate.

### 2. Use The Product Owner As Decision Gate

Ask the product owner for decisions when direction materially affects the product:

- Workflow priority.
- Mental model or domain-language choice.
- Navigation model.
- Visual density.
- Risk tolerance for destructive or state-changing actions.
- Whether to build, iterate, or promote a candidate.

Prefer one compact decision prompt at a time. Continue with reasonable assumptions only for low-risk details.

### 3. Design The Interaction Model

Before building, define:

- Routes and major states.
- Mobile and desktop layout behavior.
- Main components and their responsibilities.
- Loading, empty, error, denied, saved, and undo states.
- Keyboard, focus, and screen-reader expectations.
- Data shape for realistic mock records.
- Frontend domain types, adapter boundaries, configuration needs, and observability points.
- File ownership and split points so no route, component, adapter, helper, or mock-data file becomes a catch-all.

Keep the first screen as the usable product surface, not a marketing page.

### 4. Build A Real Candidate When Direction Needs Review

If visual or interaction direction needs product-owner feedback, create a temporary working SvelteKit candidate instead of a static mockup.

Requirements:

- Use SvelteKit.
- Use Svelte-compatible shadcn-style primitives or local equivalents matching that composition model.
- Use realistic mock data, not lorem ipsum.
- Include mobile and desktop responsive behavior.
- Include meaningful interactive states.
- Run the candidate locally when possible and provide the URL.
- Treat the candidate as production-shaped design evidence, not a disposable tracer bullet.

Read `references/sveltekit-candidate-workflow.md` before creating the candidate.

### 5. Run Adversarial Reviews

After a candidate or detailed design exists, run separate review passes using `references/ui-review-rubrics.md`.

Minimum lenses:

- Clarity and cognitive load.
- Usability heuristics.
- Accessibility and inclusive interaction.
- Mobile ergonomics.
- Information architecture.
- Visual system.
- Implementation feasibility.
- Frontend engineering discipline.
- File organization and separation of concerns.

State findings as concrete issues with risks and fixes. Do not impersonate living people; use named schools of thought only as shorthand for principles.

### 6. Synthesize And Gate

Group review output into:

- Must fix before approval.
- Should fix during iteration.
- Acceptable tradeoffs.
- Open product decisions.

Ask the product owner whether to iterate, redirect, or approve the direction.

### 7. Promote Only After Approval

Before modifying `apps/web`, update the relevant spec first. Then implement through the existing web architecture:

- Keep generated API DTOs behind adapter boundaries.
- Keep frontend domain models separate when product behavior needs them.
- Use typed domain concepts and enums instead of meaningful loose strings.
- Use shadcn-svelte primitives for generic controls.
- Preserve runtime configuration boundaries.
- Preserve domain-oriented observability through explicit, injectable helpers or ports.
- Keep files cohesive; split routes, components, adapters, mappers, state helpers, fixtures, and observability helpers before they become catch-all files.
- Verify with tests, responsive checks, accessibility checks, and the code critic agent.

## Non-Negotiables

- Do not treat brainstorming prose as sufficient for a major UI direction.
- Do not promote temporary candidate code into `apps/web` without spec updates and product-owner approval.
- Do not expand the current tracer-bullet UI as if it were the final product direction.
- Do not use React shadcn components in the SvelteKit web app.
- Do not use decorative AI gradients, mascot-like behavior, beige/rustic palettes, enterprise inventory styling, or marketing-first landing pages for product workflows.
- Do not hide destructive, saved, denied, or failed states.
- Do not use in-product copy to pitch, explain, or narrate the UI. Design the surface so available actions and state are apparent.
- Do not bypass frontend ports, adapters, generated-client boundaries, auth boundaries, runtime configuration, or observability boundaries for convenience.
- Do not hard-code environment-specific URLs, tenant IDs, inventory IDs, OIDC values, feature flags, or operational settings.
- Do not use raw `console.log`, `print`, or one-off diagnostics in production candidate paths.
- Do not create "god" files that mix unrelated UI, state, adapter, config, observability, mock-data, and mapping concerns.
