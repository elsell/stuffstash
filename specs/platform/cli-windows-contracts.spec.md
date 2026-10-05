# Windows CLI contract verification

The pinned Windows runner must execute the existing critical CLI command tests,
not only cross-compile the binary. Include application contracts, local bootstrap help, request input, output,
saved contexts and terminal parsing.
These tests exercise scripted application behavior, scope, safe diagnostics and confirmations
on Windows without live accounts or physical printers. Reuse existing tests;
do not add an implementation-mirroring Windows-only test matrix.

The workflow retains bounded test and job timeouts. A successful run proves the
covered automated contracts on Windows. It does not prove real console rendering,
keyboard interaction, browser sign-in, or printer hardware behavior; retain those
as explicit verification gaps until observed separately.

Credential-file fixtures are Unix-only because the Windows product deliberately
requires the OS credential store. Do not disable this guard to run Unix fixtures.
Authenticated Windows bootstrap/keyring verification remains outstanding; the
bootstrap job selects only local help tests that require no credentials.
