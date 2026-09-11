import { useEffect } from 'react';
import { useAppServices } from './AppServicesContext';
export function PushRegistrationLifecycle() {
 const services=useAppServices();
 useEffect(()=>services.pushReconciliation.start(),[services]);
 return null;
}
