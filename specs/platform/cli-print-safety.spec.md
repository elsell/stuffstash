# CLI print cancellation safety

`print-jobs cancel JOB_ID` displays server, household, inventory and job ID,
then asks for confirmation. Scripts and redirected input require `--yes`.
Do not read or change the job before confirmation. Once confirmed, read its
current revision and submit exactly one cancellation with that revision. Never
retry a conflict or uncertain result automatically. Return the complete job and
metadata. Explain that cancellation does not prove a physical label did not print.
Critical tests cover zero requests without confirmation, scoped revision use,
safe error handling and no automatic retry.

## Resolve uncertain output

`print-jobs resolve JOB_ID` offers keyboard choices for printed, not printed,
or unknown (unknown first). Read the current revision only after the user has
chosen an outcome, then confirm the scoped job and outcome while explaining that
this is a manual report, not proof of physical output. Send acknowledgeUncertainty
true only after confirmation. Scripts provide `--input FILE|-` with reportedOutcome,
positive revision and acknowledgeUncertainty true, plus `--yes`; preserve the
explicit revision so stale input cannot silently overwrite concurrent changes.
No command prints or retries automatically. Preserve the complete updated job.
Critical tests cover invalid input before network, explicit revision, outcome,
confirmation, conflict and cross-scope denial.
