package presentation

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func definitionFields(id, tenant string, inventory *string, scope, key, name, lifecycle string) [][2]string {
	fields := [][2]string{{"ID", id}, {"Household", tenant}, {"Scope", scope}, {"Key", key}, {"Display name", name}, {"Lifecycle", lifecycle}}
	if inventory != nil {
		fields = append(fields, [2]string{"Inventory", *inventory})
	}
	return fields
}
func (o Output) assetType(v ports.AssetType) error {
	fields := definitionFields(v.ID, v.TenantID, v.InventoryID, v.Scope, v.Key, v.DisplayName, v.Lifecycle)
	return o.details(append(fields, [2]string{"Description", v.Description}, [2]string{"Expiration enabled", strconv.FormatBool(v.ExpirationEnabled)}))
}
func (o Output) fieldDefinition(v ports.FieldDefinition) error {
	fields := definitionFields(v.ID, v.TenantID, v.InventoryID, v.Scope, v.Key, v.DisplayName, v.Lifecycle)
	options, _ := json.Marshal(v.EnumOptions)
	targets, _ := json.Marshal(v.CustomAssetTypeIDs)
	return o.details(append(fields, [2]string{"Type", v.Type}, [2]string{"Enum options", string(options)}, [2]string{"Applicability", v.Applicability}, [2]string{"Custom asset type IDs", string(targets)}))
}
