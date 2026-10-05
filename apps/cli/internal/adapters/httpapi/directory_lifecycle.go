package httpapi

import (
 "context"
 "net/http"
 "github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
 "github.com/stuffstash/stuff-stash/cli/internal/ports"
)
func(c *Client)ChangeDirectoryLifecycle(ctx context.Context,resource ports.DirectoryResource,action ports.LifecycleAction,s ports.Scope)(any,error){
 var response *http.Response;var err error
 if resource==ports.HouseholdResource{
  switch action{
  case ports.Archive:response,err=c.sdk.PatchTenantsByTenantIdArchive(ctx,s.Tenant,nil)
  case ports.Restore:response,err=c.sdk.PatchTenantsByTenantIdRestore(ctx,s.Tenant,nil)
  case ports.Delete:response,err=c.sdk.DeleteTenantsByTenantId(ctx,s.Tenant,nil)
  default:return nil,invalidLifecycle()
  }
 }else if resource==ports.InventoryResource{
  switch action{
  case ports.Archive:response,err=c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdArchive(ctx,s.Tenant,s.Inventory,nil)
  case ports.Restore:response,err=c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdRestore(ctx,s.Tenant,s.Inventory,nil)
  case ports.Delete:response,err=c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryId(ctx,s.Tenant,s.Inventory,nil)
  default:return nil,invalidLifecycle()
  }
 }else{return nil,invalidLifecycle()}
 if action==ports.Delete{
  if err:=noContent(response,err);err!=nil{return nil,err}
  result:=map[string]string{"status":"deleted","tenantId":s.Tenant}
  if resource==ports.InventoryResource{result["inventoryId"]=s.Inventory}
  return result,nil
 }
 if resource==ports.HouseholdResource{
  r,err:=read[generated.SuccessEnvelopeTenantResponse](response,err);if err!=nil{return nil,err}
  return ports.Result[ports.Tenant]{Data:tenant(r.Data),Schema:r.Schema,Meta:metadata(r.Meta)},nil
 }
 r,err:=read[generated.SuccessEnvelopeInventoryResponse](response,err);if err!=nil{return nil,err}
 return ports.Result[ports.Inventory]{Data:inventory(r.Data),Schema:r.Schema,Meta:metadata(r.Meta)},nil
}
func invalidLifecycle()error{return ports.Failure("usage","Unknown lifecycle action. Use --help to choose a command.")}
func noContent(response *http.Response,err error)error{
 if err==nil && response.StatusCode==http.StatusNoContent{response.Body.Close();return nil}
 _,readErr:=read[any](response,err);if readErr!=nil{return readErr}
 return ports.Failure("protocol","The server returned an unexpected delete response. Check the resource before you retry.")
}
