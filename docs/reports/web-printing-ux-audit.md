# Web printing settings audit

Reviewed October 4, 2026 against main `6d09cbce8`, the three supplied website
photos, and the resulting browser implementation. Scope: website printing settings
and its shared printer, test-label, asset-print and job-recovery components.

## Findings and changes

| Finding | User cost | Correction |
| --- | --- | --- |
| Defaults precede printer status; every control spans the page | Users must read setup fields to find whether printing works | Printer status/actions first, bounded defaults beside it on desktop and below on narrow screens |
| Report timestamps, CLI instructions and computers compete with daily tasks | Technical content obscures the action | Named keyboard disclosures; offline and authorization problems remain visible |
| Raw identity strings dominate resolved jobs | Overflow and misleading emphasis on implementation details | Reported outcome leads; report details retain actor, time and device confirmation separately |
| Printer edits disappear on dismissal | Accidental Escape or outside click loses work | Keep editing / Discard changes; save/loading lock dismissal; restore focus after both dialogs settle |
| Saved-defaults notice survives checkbox changes | Unsaved work looks saved | Notice reflects the saved draft, including checkbox changes |
| Registered media omits its friendly name | 29 × 89.8 mm looks different from the purchased 29 × 90 mm roll | Match catalog display name by adapter, preset and version; never change physical geometry |

The hierarchy and disclosure findings are observed in the supplied photos and
confirmed in source. Draft dismissal and nested focus handling were reproduced
and corrected in Chromium. This is not a certification of the whole website.

## Interaction decisions

Checking readiness, editing defaults, printing a test and checking the most recent
outcome are closely related household tasks. Keep one destination for this small
inventory workflow, with distinct sections, rather than requiring a tab switch to
notice an uncertain job. Use disclosures only for supplementary information.
If history becomes a frequently searched operational destination, a separate
history route may earn its navigation cost; it is not needed for this refactor.

The defaults are a single explicit-save draft: native checkboxes and the existing
accessible short-choice select remain appropriate. Printer editing is a bounded
modal task. Existing revision conflicts require explicit reload. Polling does not
replace defaults or open edits. Test printing remains an explicit command; opening
a dialog never prints. Shared job rows continue to distinguish device confirmation
from a human report and leave uncertain recovery visible.

These choices apply [Krug's usability approach](https://sensible.com/dont-make-me-think/)
as a review lens, not a claimed endorsement. Apple's current
[layout guidance](https://developer.apple.com/design/human-interface-guidelines/layout)
was read through its official DocC data on October 4: order by importance, group
related content, and disclose supplementary information. Apple-specific native
materials and controls are not imposed on the browser. Disclosure keyboard behavior
follows the [WAI pattern](https://www.w3.org/WAI/ARIA/apg/patterns/disclosure/)
through native HTML details/summary, with existing Bits/shadcn dialogs for modality.

## Coverage and evidence

- Production workspace route, shell and adapter: desktop and mobile Chromium
  acceptance covers catalog display names, report disclosure, long-ID reflow,
  Escape → Keep editing, persisted save, explicit discard and trigger focus.
- Real production components with a stateful repository: 390px and 1280px fixtures
  cover ready, empty/setup, loading, unavailable/retry, viewer, offline/paper error,
  uncertain recovery, test dialog, and report detail expansion. No hardware output.
- Existing critical component tests cover revision conflicts, permission changes,
  ambiguous print requests, linked reprints, polling and manual resolution.
- Shared consumers inspected: InventoryPrintingSettings, PrinterTestDialog and
  AssetPrintDialog all use PrintJobList; no domain, API or permission changes.
- Asset menu/scanner/pairing/rotation entry points inspected for unchanged
  composition; native clients, camera scanning, screen-reader speech and physical
  printer behavior are outside this web refactor's runtime evidence.

Screenshots from the workspace acceptance accompany the pull request. Supplied
before photos remain in the review conversation. Generated fixture routes and
local dependency-serving configuration are not shipped.

![Desktop settings in the production workspace shell](images/web-printing-ux/desktop.png)

![Narrow settings in the production workspace shell](images/web-printing-ux/mobile.png)

Validation: 1,235 web tests across 187 files; both Chromium workspace projects;
web production build and type checking; shadcn, visual-token and localization
checks; docs production build and rendered-table check. The pre-commit runner was
not installed in this worktree; its applicable commands were run directly.
Type checking retains one unrelated, existing unused `h4` selector warning.
The final code-critic review reported no confirmed blockers.
