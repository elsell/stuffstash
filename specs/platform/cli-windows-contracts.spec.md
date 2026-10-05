# Windows CLI contract verification

The pinned Windows runner must execute the existing critical CLI command tests,
not only cross-compile the binary. Include application and bootstrap boundaries,
request input, credential storage, output, saved contexts and terminal parsing.
These tests exercise scripted behavior, scope, safe diagnostics and confirmations
on Windows without live accounts or physical printers. Reuse existing tests;
do not add an implementation-mirroring Windows-only test matrix.

The workflow retains bounded test and job timeouts. A successful run proves the
covered automated contracts on Windows. It does not prove real console rendering,
keyboard interaction, browser sign-in, or printer hardware behavior; retain those
as explicit verification gaps until observed separately.
