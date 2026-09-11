import { createHash } from 'node:crypto';
const appleOrigin = 'https://api.appstoreconnect.apple.com';
const maxPages = 20;
async function collect(api, path) {
  const values = [];
  const visited = new Set();
  while (path) {
    const url = new URL(path, appleOrigin);
    if (url.origin !== appleOrigin || url.username || url.password || visited.has(url.href) || visited.size >= maxPages) throw new Error('Invalid Apple pagination');
    visited.add(url.href);
    const response = await api(url.pathname + url.search);
    if (!Array.isArray(response?.data)) throw new Error('Invalid Apple response');
    values.push(...response.data);
    path = response.links?.next;
    if (path != null && typeof path !== 'string') throw new Error('Invalid Apple pagination');
  }
  return values;
}
export async function inspectProfileRepair(api, profile, expected, now) {
  if (!expected.teamId || !expected.bundleId || profile.teamId !== expected.teamId || profile.bundleId !== expected.bundleId || !Array.isArray(profile.certificateSha256) || !profile.certificateSha256.length) throw new Error('Stored profile does not match application');
  const query = new URLSearchParams({'filter[identifier]':expected.bundleId});
  const bundles = await collect(api, `/v1/bundleIds?${query}`);
  if (bundles.length !== 1 || bundles[0]?.attributes?.identifier !== expected.bundleId || !bundles[0].id) throw new Error('Expected one matching Apple bundle');
  const certificates = await collect(api, '/v1/certificates');
  const matching = certificates.filter(value => {
    const attrs = value?.attributes;
    return value?.id && ['DISTRIBUTION','IOS_DISTRIBUTION'].includes(attrs?.certificateType)
      && new Date(attrs.expirationDate).getTime() > now.getTime()
      && typeof attrs.certificateContent === 'string'
      && profile.certificateSha256.includes(createHash('sha256').update(Buffer.from(attrs.certificateContent,'base64')).digest('hex'));
  });
  if (matching.length !== 1) throw new Error('Expected one existing valid distribution certificate');
  const bundleId = bundles[0].id;
  const capabilities = await collect(api, `/v1/bundleIds/${encodeURIComponent(bundleId)}/bundleIdCapabilities`);
  return {bundleId, certificateId:matching[0].id, pushEnabled:capabilities.some(value=>value?.attributes?.capabilityType==='PUSH_NOTIFICATIONS')};
}
export async function repairPushProfile(api, plan, name) {
  if (!plan.bundleId || !plan.certificateId || !name) throw new Error('Incomplete profile plan');
  if (!plan.pushEnabled) await api('/v1/bundleIdCapabilities', 'POST', {data:{type:'bundleIdCapabilities',attributes:{capabilityType:'PUSH_NOTIFICATIONS'},relationships:{bundleId:{data:{type:'bundleIds',id:plan.bundleId}}}}});
  return api('/v1/profiles','POST',{data:{type:'profiles',attributes:{name,profileType:'IOS_APP_STORE'},relationships:{bundleId:{data:{type:'bundleIds',id:plan.bundleId}},certificates:{data:[{type:'certificates',id:plan.certificateId}]}}}});
}
