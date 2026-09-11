package httpserver
import("encoding/json";"net/http";"strings";"testing")
func coverNotificationDeviceScenarios(t *testing.T,coverage executedScenarioCoverage,adversarial bool){
 t.Helper()
 server:=NewServer(":0",newSeededTestApp(t,seededState{tenants:[]seedTenant{{id:"home",name:"Home",owner:"owner"}},inventories:[]seedInventory{{id:"main",tenantID:"home",name:"Main",owner:"owner"}}}))
 const path="/tenants/home/inventories/main/notification-devices"
 const template="/tenants/{tenantId}/inventories/{inventoryId}/notification-devices"
 body:=map[string]any{"installationId":"phone","transport":"apns","token":strings.Repeat("ab",32),"revision":0}
 response:=performRequest(server,http.MethodPost,path,"Bearer dev:owner",body);requireStatus(t,response,http.StatusOK)
 var data struct{Data struct{ID string}};if err:=json.Unmarshal(response.Body.Bytes(),&data);err!=nil{t.Fatal(err)}
 token,status,ownedStatus:="Bearer dev:owner",http.StatusOK,http.StatusOK
 if adversarial{token,status,ownedStatus="Bearer dev:outsider",http.StatusForbidden,http.StatusNotFound}
 coverage.request(t,server,http.MethodPost,template,path,token,body,status)
 coverage.request(t,server,http.MethodGet,template+"/by-installation/{installationId}",path+"/by-installation/phone",token,nil,ownedStatus)
 coverage.request(t,server,http.MethodDelete,template+"/{deviceId}",path+"/"+data.Data.ID+"?revision=1",token,nil,ownedStatus)
}
