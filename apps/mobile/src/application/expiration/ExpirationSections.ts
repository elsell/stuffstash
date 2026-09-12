import type { AssetCardViewModel } from '../assets/AssetViewModels';
export function groupExpirationItems(items: readonly AssetCardViewModel[]) {
 const groups: { key: string; month: string; expired: boolean; items: AssetCardViewModel[] }[] = [];
 for (const item of items) {
  if (!item.expiration) continue;
  const month = item.expiration.date.slice(0, 7);
  const expired = item.expirationContext?.state === 'expired';
  const key = `${expired ? 'expired' : 'future'}:${month}`;
  let group = groups[groups.length - 1];
  if (group?.key !== key) { group = { key, month, expired, items: [] }; groups.push(group); }
  group.items.push(item);
 }
 return groups;
}
