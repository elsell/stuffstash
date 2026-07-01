import { describe, expect, it } from 'vitest';
import {
  appendVoicePlanPhotoDrafts,
  removeVoicePlanPhotoDraft,
  type VoicePlanPhotoDrafts
} from './VoicePlanPhotoDraftState';

describe('VoicePlanPhotoDrafts', () => {
  it('appends selected photos to the requested plan command without touching other rows', () => {
    const current: VoicePlanPhotoDrafts = {
      'cmd-location': [{
        id: 'photo-existing',
        uri: 'file:///existing.jpg',
        fileName: 'existing.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZXhpc3Rpbmc=',
        sizeBytes: 8
      }]
    };

    expect(appendVoicePlanPhotoDrafts(current, 'cmd-item', [{
      id: 'photo-new',
      uri: 'file:///new.jpg',
      fileName: 'new.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'bmV3',
      sizeBytes: 3
    }])).toEqual({
      'cmd-location': current['cmd-location'],
      'cmd-item': [{
        id: 'photo-new',
        uri: 'file:///new.jpg',
        fileName: 'new.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'bmV3',
        sizeBytes: 3
      }]
    });
  });

  it('keeps current draft state unchanged when the picker returns no photos', () => {
    const current: VoicePlanPhotoDrafts = {};

    expect(appendVoicePlanPhotoDrafts(current, 'cmd-item', [])).toBe(current);
  });

  it('removes one draft photo from a plan command and drops empty rows', () => {
    const current: VoicePlanPhotoDrafts = {
      'cmd-item': [{
        id: 'photo-one',
        uri: 'file:///one.jpg',
        fileName: 'one.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'b25l',
        sizeBytes: 3
      }, {
        id: 'photo-two',
        uri: 'file:///two.jpg',
        fileName: 'two.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'dHdv',
        sizeBytes: 3
      }]
    };

    const withOnePhoto = removeVoicePlanPhotoDraft(current, 'cmd-item', 'photo-one');
    expect(withOnePhoto['cmd-item']?.map((photo) => photo.id)).toEqual(['photo-two']);
    expect(removeVoicePlanPhotoDraft(withOnePhoto, 'cmd-item', 'photo-two')).toEqual({});
  });
});
