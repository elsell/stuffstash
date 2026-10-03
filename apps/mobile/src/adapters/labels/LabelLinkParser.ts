import { parseLabelLink } from '@stuff-stash/api-client';
import { LabelFailure, type LabelReference } from '../../application/labels/LabelWorkspace';
export function parseMobileLabelLink(source: string): LabelReference {
  try { return parseLabelLink(source); }
  catch { throw new LabelFailure('invalid_label'); }
}
