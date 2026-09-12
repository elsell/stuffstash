import { expect, it } from 'vitest';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { refreshExpirationHome } from './RefreshExpirationHome';
it('Home refresh fetches active expiration data in this session without touching another account',async()=>{
 const client=new QueryClient({defaultOptions:{queries:{retry:false,staleTime:Infinity}}});
 let current=0,other=0;
 const a=new QueryObserver(client,{queryKey:[...mobileQueryKeys.inventory('current','tenant','inventory'),'expiration','home'],queryFn:async()=>++current});
 const b=new QueryObserver(client,{queryKey:[...mobileQueryKeys.inventory('other','tenant','inventory'),'expiration','home'],queryFn:async()=>++other});
 const stopA=a.subscribe(()=>{}),stopB=b.subscribe(()=>{});await a.refetch();await b.refetch();const before=current,otherBefore=other;
 await refreshExpirationHome(client,'current');expect(current).toBe(before+1);expect(other).toBe(otherBefore);
 stopA();stopB();client.clear();
});
