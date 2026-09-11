export interface PushNotificationResponses {
 subscribe(listener:(payload:unknown)=>void | Promise<void>,onFailure:()=>void):()=>void;
}
