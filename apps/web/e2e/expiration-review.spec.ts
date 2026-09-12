import { expect, test } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

test('reviews every date, retains filters and returns from item detail at narrow and desktop widths', async ({page},testInfo)=>{
 resetWorkspaceApiState(page);await installAuthenticatedWorkspace(page);
 const items=Array.from({length:34},(_,index)=>({id:`expiry-${index}`,tenantId:'tenant-home',inventoryId:'inventory-household',kind:'item',lifecycleState:'active',title:index===0?'ANUCORT-HC 25MG Rectal Suppository':index===1?'Clinere Earwax Cleaners':`Contact lens solution ${index}`,description:'Medicine cabinet supplies',parentAssetId:null,customAssetTypeId:'medicine',customFields:{},tags:[],expiration:{date:index<2?'2026-08':'2026-10',precision:'month'},expirationContext:{state:index<2?'expired':'upcoming',trackingEnabled:true,advanceDays:30,timezone:'UTC'},ancestorPath:[{id:'location-garage',title:'Hall Medicine Closet'},{id:'asset-bin',title:'Cold/Cough'}],createdAt:'2026-01-01T12:00:00Z',updatedAt:'2026-09-12T12:00:00Z'}));
 await page.route('http://127.0.0.1:18080/**',async route=>{
  const url=new URL(route.request().url());
  if(url.pathname.endsWith('/expiration-assets')){
   let filtered=items.filter(item=>!url.searchParams.get('q')||item.title.toLowerCase().includes(url.searchParams.get('q')!.toLowerCase()));
   if(url.searchParams.get('fromDate'))filtered=filtered.filter(item=>item.expiration.date>='2026-10');
   const counts={all:filtered.length,expired:filtered.filter(item=>item.expirationContext.state==='expired').length,soon:filtered.filter(item=>item.expirationContext.state==='upcoming').length};
   const mode=url.searchParams.get('mode');if(mode==='expired')filtered=filtered.filter(item=>item.expirationContext.state==='expired');if(mode==='soon')filtered=filtered.filter(item=>item.expirationContext.state==='upcoming');
   const start=Number(url.searchParams.get('cursor')||0),limit=Number(url.searchParams.get('limit')||30);const next=start+limit;
   return route.fulfill({json:{data:{items:filtered.slice(start,next),counts,timezone:'UTC'},meta:{pagination:{limit,hasMore:next<filtered.length,nextCursor:next<filtered.length?String(next):null}}}});
  }
  const id=url.pathname.match(/\/assets\/(expiry-\d+)$/)?.[1];if(id)return route.fulfill({json:{data:items.find(item=>item.id===id),meta:{}}});
  return route.fallback();
 });
 await page.goto('/');const home=page.getByRole('region',{name:'Expiration',exact:true});
 await expect(home.getByText('Expired · 2',{exact:true})).toBeVisible();await expect(home.getByText('Expiring soon · 32',{exact:true})).toBeVisible();
 await page.screenshot({path:testInfo.outputPath('expiration-home.png'),fullPage:true});
 await home.getByRole('link',{name:'See all',exact:true}).click();await expect(page).toHaveURL(/\/expiration\?expiration=all/);
 await expect(page.getByRole('heading',{name:'Expiration',exact:true})).toBeVisible();await expect(page.getByText('August 2026 (end of month)',{exact:false}).first()).toBeVisible();
 await page.getByRole('button',{name:'Load more',exact:true}).click();await expect(page.getByRole('link',{name:/Contact lens solution 33/})).toBeVisible();
 await page.getByRole('link',{name:/Contact lens solution 33/}).click();await expect(page).toHaveURL(/assets\/expiry-33/);
 await page.getByRole('link',{name:'Back',exact:true}).click();await expect(page).toHaveURL(/\/expiration\?expiration=all/);await expect(page.getByRole('link',{name:/Contact lens solution 33/})).toBeVisible();
 await page.getByRole('button',{name:'Filters',exact:true}).click();await expect(page.getByRole('dialog')).toBeVisible();
 await page.getByLabel('From',{exact:true}).fill('2026-10-01');await page.getByRole('button',{name:'Apply filters',exact:true}).click();await expect(page).toHaveURL(/from=2026-10-01/);
 await page.reload();await expect(page.getByText('From 2026-10-01',{exact:false})).toBeVisible();await expect(page.getByText('ANUCORT-HC 25MG Rectal Suppository',{exact:true})).toHaveCount(0);
 await page.screenshot({path:testInfo.outputPath('expiration-filtered.png'),fullPage:true});
 expect(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth)).toBe(true);
});
