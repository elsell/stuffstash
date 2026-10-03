import {describe, expect, it} from 'vitest';
import {parseLabelLink} from './labelLink';
const instance = '01ARZ3NDEKTSV4RRFFQ69G5FAV';
const label = '01ARZ3NDEKTSV4RRFFQ69G5FAW';
describe('label links are identities, never network destinations', () => {
  it('accepts a migrated HTTPS origin and a configured prefix without retaining either', () => {
    expect(parseLabelLink(`https://old.example/stash/l/v1/${instance}/${label}`)).toEqual({instanceId:instance,labelId:label});
    expect(parseLabelLink(`stuffstash://labels/v1/${instance}/${label}`)).toEqual({instanceId:instance,labelId:label});
  });
  it.each([
    `http://old.example/l/v1/${instance}/${label}`, `https://user:pass@old.example/l/v1/${instance}/${label}`,
    `https://old.example/l/v1/${instance}/${label}?token=x`, `https://old.example/l/v1/${instance}/${label}#x`,
    `https://old.example/%2e%2e/l/v1/${instance}/${label}`, `https://old.example/a/../l/v1/${instance}/${label}`,
    `https://old.example/l/v1/${instance}/${label}/`, `https://old.example/l/v1/${instance}/%30${label.slice(1)}`,
    `https://old.example\\@evil.example/l/v1/${instance}/${label}`, `stuffstash://invitations/v1/${instance}/${label}`,
    `javascript:alert(1)`, `https://old.example/l/v2/${instance}/${label}`, `https://old.example/l/v1/${instance}/private-name`,
  ])('rejects ambiguous or unrelated input %s', value => expect(() => parseLabelLink(value)).toThrow());
});
