import { expect, it } from 'vitest';
import { localization } from '../../presentation/localization';
import { photoMetadataLabel } from './AssetPhotoWorkspacePresentation';

it('formats displayed photo sizes in the runtime locale without changing binary scaling', () => {
  expect(photoMetadataLabel({ sizeBytes: 1536 })).toBe(`${localization.number(1.5, { minimumFractionDigits: 1, maximumFractionDigits: 1 })} KB`);
  expect(photoMetadataLabel({ sizeBytes: 12 * 1024 * 1024 })).toBe(`${localization.number(12, { maximumFractionDigits: 0 })} MB`);
  expect(photoMetadataLabel({ sizeBytes: 0 })).toBeUndefined();
});
