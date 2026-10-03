import { labelWebMessages } from './labels-web';
import { labelMobileMessages } from './labels-mobile';
import { archiveMessages } from './archive';
import { workflowMessages } from './workflow';
import { mobileMessages } from './mobile';
import { webMessages } from './web';
import type { Catalog } from './translator';

/** English source catalog. Keys describe presentation context, never wire values. */
export const en = {
  ...labelWebMessages,
  ...labelMobileMessages,
  ...archiveMessages,
  ...workflowMessages,
  ...mobileMessages,
  ...webMessages,
  'browse.addAsset': 'Add an asset',
  'browse.title': 'Browse',
  'pagination.position': '{position} of {total}',
  'browse.filters': 'Filters',
  'browse.filtersApplied': { one: 'Filters, {count} applied', other: 'Filters, {count} applied' },
  "assets.visibleCount": {"one": "{count} visible asset", "other": "{count} visible assets"},
  "assets.activeCount": {"one": "{count} active asset", "other": "{count} active assets"},
  "assets.rootCount": {"one": "{count} root item", "other": "{count} root items"},
  "map.overview": "{assets} \u00b7 {roots}",
  "photos.none": "No photos",
  "photos.count": {"one": "{count} photo", "other": "{count} photos"},
  "photos.position": "{position} of {total}",
  "photos.openPosition": "Open photo {position} of {total}",
  "photos.noneAttached": "No photos are attached.",
  "photos.deleteCount": {"one": "{count} photo will be removed with it.", "other": "{count} photos will be removed with it."},
  "assets.deleteContents": " Current contents: {contents}. Deletion will not continue while active things are inside it.",
  "assets.deleteWarning": "This permanently removes {title}. {photos}{contents} Audit history remains, but the asset itself cannot be restored.",
  "fields.showUnset": {"one": "Show {count} unset field", "other": "Show {count} unset fields"},
  "fields.showEmpty": {"one": "Show {count} empty field", "other": "Show {count} empty fields"},
  "import.items": {"one": "{count} item", "other": "{count} items"},
  "import.affectedRecords": {"one": "{count} affected record", "other": "{count} affected records"},
  "import.records": {"one": "{count} record", "other": "{count} records"},
  "import.requiresAction": {"one": "{count} import requires action", "other": "{count} imports require action"},
  "import.blockingIssues": {"one": "A blocking issue or cleanup failure is flagged below.", "other": "Blocking issues or cleanup failures are flagged below."},
  "assets.count": {"one": "{count} asset", "other": "{count} assets"},
  "spaces.count": {"one": "{count} space", "other": "{count} spaces"},
  "items.count": {"one": "{count} item", "other": "{count} items"},
} as const satisfies Catalog;
