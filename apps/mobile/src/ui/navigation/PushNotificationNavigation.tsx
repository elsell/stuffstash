import {useEffect} from 'react';
import {useRootNavigationState,useRouter} from 'expo-router';
import {useAppServices} from './AppServicesContext';
import {useAppFeedback} from '../feedback/AppFeedback';
import {NotificationFailure} from '../../application/notifications/NotificationFailure';
import {assetDetailHref} from '../screens/AssetDetailNavigation';

export function PushNotificationNavigation(){
 const services=useAppServices();const router=useRouter();const navigation=useRootNavigationState();const feedback=useAppFeedback();
 useEffect(()=>{
  if(!navigation?.key)return;
  let current:AbortController|undefined;
  const unsubscribe=services.pushNotificationResponses.subscribe(payload=>{
   current?.abort();const controller=new AbortController();current=controller;
   return services.openPushNotification.execute(payload,{signal:controller.signal}).then(assetId=>{
    if(!controller.signal.aborted)router.push(assetDetailHref(assetId));
   }).catch(error=>{
    if(!controller.signal.aborted)feedback.showDialog({title:'Notification unavailable',message:error instanceof NotificationFailure?error.message:'This notification could not be opened. Check your notifications inbox and try again.',primaryAction:{label:'OK'}});
   });
  },()=>{services.notificationObserver.record({operation:'push-open',outcome:'failed'});});
  return ()=>{current?.abort();unsubscribe();};
 },[services,router,feedback,navigation?.key]);
 return null;
}
