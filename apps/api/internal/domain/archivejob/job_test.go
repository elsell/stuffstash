package archivejob

import (
	"testing"
	"time"
)

func startTime() time.Time { return time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC) }
func exportRequest() Request {
	return Request{RequestKey: "request", ID: "job", TenantID: "tenant", PrincipalID: "owner", Kind: Export, SourceInventoryID: "source", Photos: true, OtherFiles: true}
}
func TestExpiredClaimCannotCompleteAfterReclaim(t *testing.T) {
	now := startTime()
	job, err := New(exportRequest(), now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	first, err := job.Claim("worker-one", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = first.Claim("worker-two", now.Add(time.Second), now.Add(time.Minute)); err == nil {
		t.Fatal("live lease stolen")
	}
	second, err := first.Claim("worker-two", now.Add(2*time.Minute), now.Add(3*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = second.Complete("worker-one", now.Add(2*time.Minute), "artifact"); err == nil {
		t.Fatal("stale worker published")
	}
	done, err := second.Complete("worker-two", now.Add(2*time.Minute), "artifact")
	if err != nil || done.State != Ready || done.Revision <= second.Revision {
		t.Fatalf("completion: %#v %v", done, err)
	}
}
func TestRestoreRequiresApprovalAndRetainsDestinationOnRetry(t *testing.T) {
	now := startTime()
	r := Request{RequestKey: "request", ID: "job", TenantID: "tenant", PrincipalID: "owner", Kind: Restore, SourceArtifactID: "upload", SourceSHA256: "aabb"}
	if _, err := New(r, now, now.Add(time.Hour)); err == nil {
		t.Fatal("invalid source hash accepted")
	}
	r.SourceSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	job, err := New(r, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	claim, err := job.Claim("validate", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = claim.Complete("validate", now, "inventory"); err == nil {
		t.Fatal("validation bypassed approval")
	}
	preview, err := claim.PreviewReady("validate", now, "plan", r.SourceSHA256, "new-inventory")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = preview.Claim("restore", now, now.Add(time.Minute)); err == nil {
		t.Fatal("unapproved restore started")
	}
	approved, err := preview.Approve("My inventory", now)
	if err != nil {
		t.Fatal(err)
	}
	running, err := approved.Claim("restore", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	failed, err := running.Fail("restore", now, FailureStorage)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := failed.Retry(now)
	if err != nil || retry.DestinationInventoryID != "new-inventory" || retry.DestinationName != "My inventory" || retry.Phase != Execution {
		t.Fatalf("retry changed approval: %#v %v", retry, err)
	}
}
func TestCancellationAndExpiryFencePublication(t *testing.T) {
	now := startTime()
	job, err := New(exportRequest(), now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	running, err := job.Claim("worker", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := running.Cancel(now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cancelled.Complete("worker", now, "artifact"); err == nil {
		t.Fatal("cancelled job published")
	}
	if _, err = running.Complete("worker", now.Add(time.Minute), "artifact"); err == nil {
		t.Fatal("expired lease published")
	}
	expired, err := running.Expire(now.Add(time.Hour))
	if err != nil || expired.State != Expired {
		t.Fatalf("expiry: %v", err)
	}
	if _, err = expired.Retry(now.Add(time.Hour)); err == nil {
		t.Fatal("expired job retried")
	}
}

func TestApprovalCannotReplaceValidatedPlan(t *testing.T) {
	now := startTime()
	r := Request{RequestKey: "request", ID: "job", TenantID: "tenant", PrincipalID: "owner", Kind: Restore, SourceArtifactID: "source", SourceSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	job, err := New(r, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	running, err := job.Claim("validator", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := running.PreviewReady("validator", now, "plan", r.SourceSHA256, "new-inventory")
	if err != nil {
		t.Fatal(err)
	}
	approved, err := preview.Approve("Restored", now)
	if err != nil {
		t.Fatal(err)
	}
	if approved.PlanArtifactID != "plan" || approved.PlanSHA256 != r.SourceSHA256 || approved.DestinationInventoryID != "new-inventory" {
		t.Fatal("approval replaced validated plan")
	}
	tampered := approved
	tampered.PlanArtifactID = "different-plan"
	if err := ValidateSuccessor(preview, tampered); err == nil {
		t.Fatal("plan substitution accepted")
	}
}

func TestPublishedRestoreCannotCompleteOrCancelBeforeFinalization(t *testing.T) {
	now := startTime()
	r, _ := New(Request{RequestKey: "restore", ID: "job", TenantID: "tenant", PrincipalID: "owner", Kind: Restore, SourceArtifactID: "upload", SourceSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, now, now.Add(time.Hour))
	r, _ = r.Claim("validate", now, now.Add(time.Minute))
	r, _ = r.PreviewReady("validate", now, "plan", r.SourceSHA256, "destination")
	r, _ = r.Approve("Restored", now)
	r, _ = r.Claim("publish", now, now.Add(time.Minute))
	if _, err := r.Complete("publish", now, ""); err == nil {
		t.Fatal("ready before authorization finalization")
	}
	published, err := r.Published("publish", now, "grant")
	if err != nil {
		t.Fatal(err)
	}
	if err = ValidateSuccessor(r, published); err != nil {
		t.Fatal(err)
	}
	if _, err = published.Cancel(now); err == nil {
		t.Fatal("cancelled published inventory")
	}
	claimed, err := published.Claim("finalize", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	pending, err := claimed.DeferFinalization("finalize", now, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if err = ValidateSuccessor(claimed, pending); err != nil {
		t.Fatal(err)
	}
	if _, err := pending.Claim("early", now, now.Add(time.Minute)); err == nil {
		t.Fatal("claimed before delay")
	}
	restarted, err := pending.Claim("restarted", now.Add(time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = restarted.Complete("finalize", now.Add(time.Second), ""); err == nil {
		t.Fatal("stale finalizer completed")
	}
	ready, err := restarted.Complete("restarted", now.Add(time.Second), "")
	if err != nil || ready.State != Ready || ready.OwnerGrantEventID != "grant" {
		t.Fatal("finalization lost publication", err)
	}
}
