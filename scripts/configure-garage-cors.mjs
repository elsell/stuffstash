import { createHash, createHmac } from 'node:crypto';
import http from 'node:http';
import https from 'node:https';

const endpoint = requiredEnv('STUFF_STASH_GARAGE_CORS_ENDPOINT');
const bucket = requiredEnv('STUFF_STASH_S3_BUCKET');
const accessKey = requiredEnv('STUFF_STASH_S3_ACCESS_KEY');
const secretKey = requiredEnv('STUFF_STASH_S3_SECRET_KEY');
const region = process.env.STUFF_STASH_S3_REGION || 'garage';
const allowedOrigin = requiredEnv('STUFF_STASH_WEB_ORIGIN');
const maxAttempts = Number.parseInt(process.env.STUFF_STASH_GARAGE_CORS_ATTEMPTS || '20', 10);

const corsXML = `<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <CORSRule>
    <AllowedOrigin>${escapeXML(allowedOrigin)}</AllowedOrigin>
    <AllowedMethod>GET</AllowedMethod>
    <AllowedMethod>POST</AllowedMethod>
    <AllowedMethod>PUT</AllowedMethod>
    <AllowedHeader>*</AllowedHeader>
    <ExposeHeader>ETag</ExposeHeader>
  </CORSRule>
</CORSConfiguration>`;

for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
  try {
    await putBucketCORS();
    console.log(`Configured Garage CORS for ${bucket} and ${allowedOrigin}`);
    process.exit(0);
  } catch (error) {
    if (attempt === maxAttempts) {
      throw error;
    }
    console.error(`Garage CORS configuration attempt ${attempt} failed: ${error.message}`);
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
}

async function putBucketCORS() {
  const url = new URL(endpoint);
  const body = Buffer.from(corsXML, 'utf8');
  const amzDate = new Date().toISOString().replace(/[:-]|\.\d{3}/g, '');
  const dateStamp = amzDate.slice(0, 8);
  const payloadHash = sha256Hex(body);
  const canonicalURI = `/${encodeURIComponent(bucket)}`;
  const canonicalQueryString = 'cors=';
  const host = url.host;
  const canonicalHeaders = `host:${host}\nx-amz-content-sha256:${payloadHash}\nx-amz-date:${amzDate}\n`;
  const signedHeaders = 'host;x-amz-content-sha256;x-amz-date';
  const canonicalRequest = ['PUT', canonicalURI, canonicalQueryString, canonicalHeaders, signedHeaders, payloadHash].join('\n');
  const credentialScope = `${dateStamp}/${region}/s3/aws4_request`;
  const stringToSign = ['AWS4-HMAC-SHA256', amzDate, credentialScope, sha256Hex(canonicalRequest)].join('\n');
  const signingKey = signatureKey(secretKey, dateStamp, region, 's3');
  const signature = hmacHex(signingKey, stringToSign);
  const authorization = `AWS4-HMAC-SHA256 Credential=${accessKey}/${credentialScope}, SignedHeaders=${signedHeaders}, Signature=${signature}`;

  await request(url, {
    method: 'PUT',
    path: `${canonicalURI}?cors`,
    body,
    headers: {
      Authorization: authorization,
      Host: host,
      'Content-Length': String(body.length),
      'Content-Type': 'application/xml',
      'X-Amz-Content-Sha256': payloadHash,
      'X-Amz-Date': amzDate
    }
  });
}

function request(url, options) {
  return new Promise((resolve, reject) => {
    const client = url.protocol === 'https:' ? https : http;
    const req = client.request({
      hostname: url.hostname,
      port: url.port || (url.protocol === 'https:' ? 443 : 80),
      method: options.method,
      path: options.path,
      headers: options.headers
    }, (res) => {
      const chunks = [];
      res.on('data', (chunk) => chunks.push(chunk));
      res.on('end', () => {
        const body = Buffer.concat(chunks).toString('utf8');
        if (res.statusCode && res.statusCode >= 200 && res.statusCode < 300) {
          resolve();
          return;
        }
        reject(new Error(`Garage returned HTTP ${res.statusCode}: ${body}`));
      });
    });
    req.on('error', reject);
    req.end(options.body);
  });
}

function signatureKey(key, dateStamp, regionName, serviceName) {
  const kDate = hmac(`AWS4${key}`, dateStamp);
  const kRegion = hmac(kDate, regionName);
  const kService = hmac(kRegion, serviceName);
  return hmac(kService, 'aws4_request');
}

function hmac(key, data) {
  return createHmac('sha256', key).update(data, 'utf8').digest();
}

function hmacHex(key, data) {
  return createHmac('sha256', key).update(data, 'utf8').digest('hex');
}

function sha256Hex(value) {
  return createHash('sha256').update(value).digest('hex');
}

function requiredEnv(name) {
  const value = process.env[name]?.trim();
  if (!value) {
    throw new Error(`${name} is required`);
  }
  return value;
}

function escapeXML(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&apos;');
}
