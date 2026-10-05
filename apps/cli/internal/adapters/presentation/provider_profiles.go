package presentation

import (
	"encoding/json"
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"text/tabwriter"
)

func (o Output) providerProfiles(profiles []ports.ProviderProfile) error {
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tNAME\tCAPABILITY\tPROVIDER\tMODEL\tSTATE"); err != nil {
		return err
	}
	for _, v := range profiles {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", strconv.Quote(v.ID), strconv.Quote(v.DisplayName), strconv.Quote(v.Capability), strconv.Quote(v.ProviderKind), strconv.Quote(v.ModelName), strconv.Quote(v.LifecycleState)); err != nil {
			return err
		}
	}
	return w.Flush()
}
func (o Output) providerProfile(v ports.ProviderProfile) error {
	fields := [][2]string{{"ID", v.ID}, {"Household", v.TenantID}, {"Name", v.DisplayName}, {"Capability", v.Capability}, {"Provider", v.ProviderKind}, {"Model", v.ModelName}, {"Endpoint", v.EndpointURL}, {"State", v.LifecycleState}, {"Credential status", v.CredentialStatus}, {"Created", v.CreatedAt}, {"Updated", v.UpdatedAt}}
	if v.LastTestedAt != nil {
		fields = append(fields, [2]string{"Last tested", *v.LastTestedAt})
	}
	if v.PromptTemplate != nil {
		fields = append(fields, [2]string{"Prompt template", *v.PromptTemplate})
	}
	for _, entry := range []struct {
		label string
		value map[string]interface{}
	}{{"Runtime options", v.RuntimeOptions}, {"Capability metadata", v.CapabilityMetadata}} {
		data, err := json.Marshal(entry.value)
		if err != nil {
			return err
		}
		fields = append(fields, [2]string{entry.label, string(data)})
	}
	return o.details(fields)
}
