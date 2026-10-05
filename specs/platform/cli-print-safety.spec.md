# CLI print cancellation safety

`print-jobs cancel JOB_ID` displays server, household, inventory and job ID,
then asks for confirmation. Scripts and redirected input require `--yes`.
Do not read or change the job before confirmation. Once confirmed, read its
current revision and submit exactly one cancellation with that revision. Never
retry a conflict or uncertain result automatically. Return the complete job and
metadata. Explain that cancellation does not prove a physical label did not print.
Critical tests cover zero requests without confirmation, scoped revision use,
safe error handling and no automatic retry.
