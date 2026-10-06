package presentation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"strconv"
)

type Output struct {
	Stdout, Stderr io.Writer
	JSON           bool
}

func (o Output) Result(value any) error {
	if o.JSON {
		return json.NewEncoder(o.Stdout).Encode(value)
	}
	switch v := value.(type) {
	case ports.Result[[]ports.SearchResult]:
		return o.searchResults(v)
	case ports.Result[ports.PairingReview]:
		return o.pairingReview(v.Data)
	case ports.Result[ports.ApprovedPairing]:
		return o.details([][2]string{{"Pairing", v.Data.ID}, {"State", v.Data.State}, {"Expires", v.Data.ExpiresAt}})
	case ports.Result[ports.ApprovedConnector]:
		return o.approvedConnector(v.Data)

	case ports.Result[[]ports.ConsumerPrinter]:
		for _, item := range v.Data {
			if err := o.consumerPrinter(item); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.ConsumerAttempt]:
		for _, item := range v.Data {
			if err := o.consumerAttempt(&item); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[*ports.ConsumerAttempt]:
		return o.consumerAttempt(v.Data)

	case ports.Result[*ports.VoiceProviderConfiguration]:
		return o.voiceProvider(v.Data)
	case ports.Result[[]ports.EvaluationCaseHead]:
		return o.evaluationCases(v)
	case ports.Result[[]ports.EvaluationRunHead]:
		return o.evaluationRuns(v)
	case ports.Result[ports.EvaluationCaseRevision]:
		return o.evaluationRevision(v.Data)
	case ports.Result[ports.EvaluationRun]:
		return o.evaluationRun(v.Data)
	case ports.Result[[]ports.EvaluationCaseRevision]:
		for _, revision := range v.Data {
			if err := o.evaluationRevision(revision); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)

	case ports.Result[[]ports.WorkflowHead]:
		return o.workflows(v)
	case ports.Result[[]ports.WorkflowRevision]:
		return o.workflowRevisions(v)
	case ports.Result[ports.WorkflowRevision]:
		return o.workflowRevision(v.Data)
	case ports.Result[*ports.WorkflowSelection]:
		return o.workflowSelection(v.Data)
	case ports.Result[ports.PrintConnector]:
		return o.printConnector(v.Data)
	case ports.Result[[]ports.PrintConnector]:
		for _, c := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", strconv.Quote(c.ID), strconv.Quote(c.Name), strconv.Quote(c.State), strconv.Quote(c.Availability)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)

	case ports.Result[ports.PrintSettings]:
		printer := "Not set"
		if v.Data.DefaultPrinterID != nil {
			printer = *v.Data.DefaultPrinterID
		}
		return o.details([][2]string{{"Default printer", printer}, {"Print on create", strconv.FormatBool(v.Data.PrintOnCreateDefault)}, {"Template", v.Data.Template.ID}, {"Template version", strconv.FormatInt(int64(v.Data.Template.Version), 10)}, {"Show reference", strconv.FormatBool(v.Data.Template.Options.ShowReference)}, {"Revision", strconv.FormatUint(v.Data.Revision, 10)}})

	case ports.Result[ports.ProviderTest]:
		return o.details([][2]string{{"Profile", v.Data.ProfileID}, {"Status", v.Data.Status}, {"Message", v.Data.Message}, {"Provider", v.Data.ProviderKind}, {"Capability", v.Data.Capability}, {"Tested", v.Data.TestedAt}})
	case ports.Result[ports.ProviderProfile]:
		return o.providerProfile(v.Data)
	case ports.Result[[]ports.ProviderProfile]:
		return o.providerProfiles(v.Data)
	case ports.Result[ports.ImportJob]:
		return o.importJob(v.Data)
	case ports.Result[ports.ImportJobList]:
		return o.importJobs(v.Data.Jobs)
	case ports.Result[ports.ServerInfo]:
		return o.details([][2]string{{"Instance ID", v.Data.InstanceID}, {"Protocol version", strconv.FormatInt(v.Data.ProtocolVersion, 10)}})
	case ports.Result[ports.ServerAuthConfig]:
		return o.serverAuthConfig(v.Data)
	case ports.Result[ports.InvitationPreview]:
		return o.invitationPreview(v.Data)
	case ports.Result[ports.InvitationAcceptance]:
		if err := o.invitation(v.Data.Invitation); err != nil {
			return err
		}
		return o.accessGrant(v.Data.Grant)
	case ports.Result[ports.Invitation]:
		return o.invitation(v.Data)
	case ports.Result[[]ports.Invitation]:
		return o.invitations(v)
	case ports.Result[ports.AccessGrant]:
		return o.accessGrant(v.Data)
	case ports.Result[[]ports.AccessGrant]:
		return o.accessGrants(v)
	case ports.Result[[]ports.Activity]:
		return o.activity(v)
	case ports.Result[[]ports.AuditRecord]:
		return o.auditRecords(v)
	case ports.Result[ports.NotificationPreferences]:
		return o.notificationPreferences(v.Data)
	case ports.Result[ports.NotificationDevice]:
		return o.details([][2]string{{"Device ID", v.Data.ID}, {"Installation", v.Data.InstallationID}, {"Transport", v.Data.Transport}, {"Revision", strconv.FormatInt(v.Data.Revision, 10)}, {"Active", strconv.FormatBool(v.Data.Active)}})
	case ports.Result[[]ports.Notification]:
		return o.notificationList(v)
	case ports.Result[ports.Notification]:
		return o.notificationDetails(v.Data)
	case ports.Result[ports.NotificationRead]:
		return o.details([][2]string{{"ID", v.Data.ID}, {"Read", strconv.FormatBool(v.Data.Read)}})
	case ports.Result[ports.NotificationCount]:
		if err := o.details([][2]string{{"Unread", strconv.FormatInt(v.Data.Count, 10)}}); err != nil {
			return err
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.NotificationReadAll]:
		if err := o.details([][2]string{{"Complete", strconv.FormatBool(v.Data.Complete)}}); err != nil {
			return err
		}
		return o.pagination(v.Pagination)

	case ports.Result[ports.Attachment]:
		return o.attachmentDetails(v.Data)
	case ports.Result[[]ports.Attachment]:
		return o.attachmentList(v)
	case ports.Result[[]ports.Tag]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", strconv.Quote(item.ID), strconv.Quote(item.DisplayName), strconv.Quote(item.Key), strconv.Quote(item.Lifecycle)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.ExpirationWorkspace]:
		if err := o.expiration(v.Data); err != nil {
			return err
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.CheckedOutAsset]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s  %s  %s  %s\n", strconv.Quote(item.Asset.ID), strconv.Quote(item.Asset.Title), strconv.Quote(item.Checkout.CheckedOutByPrincipalID), strconv.Quote(item.Checkout.CheckedOutAt)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.Checkout]:
		return o.checkoutDetails(v.Data)
	case ports.Result[[]ports.Checkout]:
		for i, item := range v.Data {
			if i > 0 {
				if _, err := io.WriteString(o.Stdout, "\n"); err != nil {
					return err
				}
			}
			if err := o.checkoutDetails(item); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.Tag]:
		fields := [][2]string{{"Tag", v.Data.DisplayName}, {"ID", v.Data.ID}, {"Key", v.Data.Key}, {"State", v.Data.Lifecycle}}
		if v.Data.Color != nil {
			fields = append(fields, [2]string{"Color", *v.Data.Color})
		}
		return o.details(fields)
	case ports.Result[ports.Principal]:
		fields := [][2]string{{"ID", v.Data.ID}}
		if v.Data.DisplayName != nil {
			fields = append(fields, [2]string{"Name", *v.Data.DisplayName})
		}
		if v.Data.Email != nil {
			fields = append(fields, [2]string{"Email", *v.Data.Email})
		}
		return o.details(fields)
	case ports.Result[ports.Tenant]:
		return o.details([][2]string{{"Household", v.Data.Name}, {"ID", v.Data.ID}, {"State", v.Data.Lifecycle}, {"Access", v.Data.Access.Relationship}})
	case ports.Result[ports.Inventory]:
		return o.details([][2]string{{"Inventory", v.Data.Name}, {"ID", v.Data.ID}, {"Household ID", v.Data.TenantID}, {"State", v.Data.Lifecycle}, {"Access", v.Data.Access.Relationship}})
	case ports.Result[[]ports.Tenant]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\n", strconv.Quote(item.ID), strconv.Quote(item.Name), strconv.Quote(item.Lifecycle)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.Inventory]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\n", item.ID, strconv.Quote(item.Name), item.Lifecycle); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.Asset]:
		for _, item := range v.Data {
			if err := o.asset(item); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.PrintJobSummary]:
		return o.printJob(v.Data)
	case ports.Result[[]ports.PrintJobSummary]:
		for _, j := range v.Data {
			if err := o.printJobRow(j); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.RegisteredPrinter]:
		return o.registeredPrinter(v.Data)
	case ports.Result[[]ports.RegisteredPrinter]:
		for _, p := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", strconv.Quote(p.ID), strconv.Quote(p.Name), strconv.Quote(p.Readiness), strconv.Quote(p.MediaName)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.LabelTemplate]:
		return o.labelTemplates(v.Data)
	case ports.Result[[]ports.PrinterProfile]:
		return o.printerProfiles(v.Data)
	case ports.Result[ports.ResolvedLabel]:
		return o.details([][2]string{{"Asset", v.Data.AssetID}, {"Household", v.Data.TenantID}, {"Inventory", v.Data.InventoryID}, {"Lifecycle", v.Data.Lifecycle}, {"Instance", v.Data.InstanceID}, {"Label", v.Data.LabelID}, {"URL", v.Data.URL}})
	case ports.LabelFileResult:
		_, err := fmt.Fprintf(o.Stdout, "Saved %s (%s), sha256=%s\n", strconv.Quote(v.Path), v.Format, v.SHA256)
		return err
	case ports.Result[ports.TelemetryAccepted]:
		return o.details([][2]string{{"Measurements accepted", strconv.Itoa(v.Data.Accepted)}})
	case ports.Result[ports.Asset]:
		return o.assetDetails(v.Data)
	default:
		encoder := json.NewEncoder(o.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
}
func (o Output) asset(a ports.Asset) error {
	_, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", strconv.Quote(a.ID), strconv.Quote(a.Kind), strconv.Quote(a.Title), strconv.Quote(a.Lifecycle))
	if err == nil && a.PrintJobID != nil {
		_, err = fmt.Fprintf(o.Stdout, "Label job: %s\n", strconv.Quote(*a.PrintJobID))
	}
	return err
}
func (o Output) pagination(p *ports.Pagination) error {
	if p != nil && p.HasMore && p.NextCursor != nil {
		_, err := fmt.Fprintf(o.Stdout, "More results: --cursor %s\n", strconv.Quote(*p.NextCursor))
		return err
	}
	return nil
}
func (o Output) Notice(message string) error { _, err := fmt.Fprintln(o.Stderr, message); return err }
func (o Output) Error(category, message string) {
	if o.JSON {
		_ = json.NewEncoder(o.Stderr).Encode(map[string]any{"error": map[string]string{"category": category, "message": message}})
		return
	}
	_, _ = fmt.Fprintf(o.Stderr, "%s: %s\n", category, message)
}

type SilentObserver struct{}

func (SilentObserver) Event(context.Context, string) {}

func (o Output) details(fields [][2]string) error {
	for _, field := range fields {
		if _, err := fmt.Fprintf(o.Stdout, "%-13s %s\n", field[0]+":", strconv.Quote(field[1])); err != nil {
			return err
		}
	}
	return nil
}
