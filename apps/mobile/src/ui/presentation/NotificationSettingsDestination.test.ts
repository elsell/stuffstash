import {expect,it} from 'vitest';
import {notificationSettingsEditorTarget} from './NotificationSettingsDestination';
it('requires an explicit scalar inventory scope and a supported destination',()=>{
 const scope={tenantId:'tenant',inventoryId:'inventory'};
 expect(notificationSettingsEditorTarget({...scope,view:'timing',typeId:'medicine'})).toEqual({scope,page:{kind:'timing',typeId:'medicine'}});
 expect(notificationSettingsEditorTarget({...scope,view:'timezone'})).toEqual({scope,page:{kind:'timezone'}});
 for(const params of [{view:'timing'},{...scope,view:'unknown'},{...scope,view:'type'},{...scope,view:'type',typeId:['a','b']},{...scope,view:'timing',tenantId:['other']},{...scope,view:'timing',inventoryId:''}])expect(notificationSettingsEditorTarget(params)).toBeNull();
});
