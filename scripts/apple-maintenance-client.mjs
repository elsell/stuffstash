import { sign } from 'node:crypto';
const origin='https://api.appstoreconnect.apple.com';
export function appleMaintenanceClient({privateKey,keyId,issuerId,fetch:fetcher=fetch,now=()=>new Date()}) {
 if(!/^[A-Z0-9]{10}$/.test(keyId)||!/^[0-9a-fA-F-]{36}$/.test(issuerId)) throw new Error('Invalid Apple credential identifiers');
 return async(path,method='GET',body)=>{
  const url=new URL(path,origin);
  if(url.origin!==origin||url.username||url.password||!url.pathname.startsWith('/v1/'))throw new Error('Invalid Apple maintenance URL');
  const iat=Math.floor(now().getTime()/1000);
  const encode=value=>Buffer.from(JSON.stringify(value)).toString('base64url');
  const input=`${encode({alg:'ES256',kid:keyId,typ:'JWT'})}.${encode({iss:issuerId,iat,exp:iat+300,aud:'appstoreconnect-v1'})}`;
  const signature=sign('sha256',Buffer.from(input),{key:privateKey,dsaEncoding:'ieee-p1363'}).toString('base64url');
  let response;
  try {response=await fetcher(url.href,{method,redirect:'error',signal:AbortSignal.timeout(30000),headers:{Authorization:`Bearer ${input}.${signature}`,'Content-Type':'application/json'},body:body===undefined?undefined:JSON.stringify(body)});}
  catch {throw new Error('Apple maintenance request did not complete; inspect before retrying writes');}
  if(!response.ok)throw new Error(`Apple maintenance request failed (HTTP ${response.status})`);
  try{return await response.json();}catch{throw new Error('Invalid Apple maintenance response');}
 };
}
