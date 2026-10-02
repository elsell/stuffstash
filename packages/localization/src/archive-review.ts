import type { archiveMessages } from './archive';
import type { MessageValues } from './translator';

type ArchiveMessage = (key: keyof typeof archiveMessages, values?: MessageValues) => string;
interface ReviewCounts {
  readonly assets: number;
  readonly tags: number;
  readonly photos: number;
  readonly otherFiles: number;
  readonly customAssetTypes: number;
  readonly customFields: number;
  readonly omittedAttachments: number;
  readonly keyRemappings: readonly unknown[];
}

/** Shared count presentation; catalog templates own ordering and complete notices. */
export function formatArchiveReview(t: ArchiveMessage, counts: ReviewCounts) {
  return {
    content: t('archive.counts', {
      assets: t('archive.items', { count: counts.assets }),
      tags: t('archive.tags', { count: counts.tags }),
      photos: t('archive.photoCount', { count: counts.photos }),
      files: t('archive.fileCount', { count: counts.otherFiles })
    }),
    schema: t('archive.definitions', {
      types: t('archive.typeCount', { count: counts.customAssetTypes }),
      fields: t('archive.fieldCount', { count: counts.customFields })
    }),
    omitted: t('archive.omitted', { count: counts.omittedAttachments }),
    remappings: t('archive.remappings', { count: counts.keyRemappings.length })
  };
}
