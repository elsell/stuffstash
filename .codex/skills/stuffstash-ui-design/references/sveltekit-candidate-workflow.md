# SvelteKit Candidate Workflow

Use a temporary SvelteKit candidate when the product owner needs to react to a real UI direction.

## Candidate Rules

- Create the candidate outside the repo's production app unless explicitly promoting approved work.
- Prefer `/tmp/stuffstash-ui-candidate-*` or the current task's temporary directory.
- Use the production-intended stack: SvelteKit and Svelte-compatible shadcn-style composition.
- Use mock data that reflects Stuff Stash domain language and likely household records.
- Include responsive mobile and desktop behavior.
- Include realistic page states, not only the happy path.
- Run the dev server and provide the local URL when possible.

## Mock Data Expectations

Use records such as:

- Inventories for a home, workshop, or move.
- Locations like Garage, Kitchen pantry, Medicine cabinet, Office closet.
- Containers like Clear bin A4, Document safe, Cable drawer.
- Items like passports, allergy medicine, USB-C chargers, winter gloves, paint rollers.
- Sharing states like owner, editor, viewer, pending invitation.
- Audit or undo states when relevant.

Avoid lorem ipsum and fake warehouse SKUs unless designing import/export or power-user metadata.

## Implementation Shape

Prefer this structure:

- `src/lib/mockData.ts` for realistic mock records.
- `src/lib/types.ts` for frontend domain types.
- `src/lib/config.ts` for candidate-local runtime-style configuration when the design needs configurable values.
- `src/lib/observability.ts` for domain-oriented candidate events when observability is part of the workflow.
- `src/lib/components/ui/` for generic primitives or local shadcn-style equivalents.
- `src/lib/components/stuffstash/` for product-specific composition.
- `src/routes/+page.svelte` for the first candidate route.

Keep candidate components small enough that the design can later be promoted or discarded deliberately.

Follow `references/frontend-engineering-principles.md` for typed domain models, adapter boundaries, DRY decisions, configuration, observability expectations, and file separation.

## Verification

Before asking for product-owner feedback:

- Start the SvelteKit dev server.
- Check a desktop viewport.
- Check a mobile viewport.
- Exercise major interactive states.
- Check keyboard focus order for core actions.
- Check for hard-coded environment values, loose domain strings, generated DTO leakage, and raw logging.
- Check for catch-all files that mix route composition, mock data, mapping, config, observability, and generic UI primitives.
- Scan for text overflow and overlapping UI.
- Note any dependency, browser, or environment checks that could not run.

## Promotion

Do not copy a candidate into `apps/web` mechanically.

Before promotion:

- Ask the product owner to approve the direction.
- Update the relevant specs.
- Rebuild the implementation inside `apps/web` using existing repo boundaries.
- Keep generated API clients and DTOs behind the web adapter boundary.
- Add tests for real behavior, not only snapshots or rendering smoke.
