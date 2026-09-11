import { createPrivateKey } from 'node:crypto';
import { readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { appleMaintenanceClient } from './apple-maintenance-client.mjs';
import { inspectProfileRepair, repairPushProfile } from './ios-profile-maintenance.mjs';
try {
 const [mode,directory]=process.argv.slice(2);
 if(!['inspect','repair'].includes(mode)||!directory)throw new Error('Invalid maintenance arguments');
 const original=JSON.parse(await readFile(join(directory,'original.json'),'utf8'));
 const privateKey=createPrivateKey(Buffer.from(process.env.APP_STORE_CONNECT_API_KEY_BASE64,'base64'));
 const api=appleMaintenanceClient({privateKey,keyId:process.env.APP_STORE_CONNECT_KEY_ID,issuerId:process.env.APP_STORE_CONNECT_ISSUER_ID});
 const plan=await inspectProfileRepair(api,original,{teamId:original.teamId,bundleId:original.bundleId},new Date());
 await writeFile(join(directory,'inspection.json'),JSON.stringify({pushEnabled:plan.pushEnabled,existingCertificateMatched:true}));
 if(mode==='repair') {
  const response=await repairPushProfile(api,plan,`Stuff Stash Push ${process.env.GITHUB_RUN_ID}-${process.env.GITHUB_RUN_ATTEMPT}`);
  const content=response?.data?.attributes?.profileContent;
  if(typeof content!=='string'||!content.length)throw new Error('Apple did not return a profile');
  await writeFile(join(directory,'replacement.mobileprovision'),Buffer.from(content,'base64'),{mode:0o600,flag:'wx'});
 }
} catch(error) {
 process.stderr.write(`${error.message.startsWith('Apple maintenance request')?error.message:'iOS profile maintenance failed'}\n`);
 process.exitCode=1;
}
